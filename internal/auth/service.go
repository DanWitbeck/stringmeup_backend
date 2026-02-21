// internal/auth/service.go
package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/example/threadcraft-backend/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	db  *pgxpool.Pool
	cfg *config.Config
}

func NewService(db *pgxpool.Pool, cfg *config.Config) *Service {
	return &Service{db: db, cfg: cfg}
}

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Tokens struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func (s *Service) Register(ctx context.Context, email, password, name string) (*User, *Tokens, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, err
	}

	user := &User{
		ID:        uuid.New().String(),
		Email:     email,
		Name:      name,
		CreatedAt: time.Now().UTC(),
	}

	_, err = s.db.Exec(ctx,
		`INSERT INTO users (id, email, password_hash, name, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		user.ID, user.Email, string(hash), user.Name, user.CreatedAt,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("insert user: %w", err)
	}

	tokens, err := s.issueTokens(ctx, user.ID)
	if err != nil {
		return nil, nil, err
	}
	return user, tokens, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*User, *Tokens, error) {
	var user User
	var hash string
	err := s.db.QueryRow(ctx,
		`SELECT id, email, name, password_hash, created_at FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Email, &user.Name, &hash, &user.CreatedAt)
	if err != nil {
		return nil, nil, fmt.Errorf("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return nil, nil, fmt.Errorf("invalid credentials")
	}

	tokens, err := s.issueTokens(ctx, user.ID)
	if err != nil {
		return nil, nil, err
	}
	return &user, tokens, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (*Tokens, error) {
	// Validate the refresh token from DB
	var userID string
	err := s.db.QueryRow(ctx,
		`SELECT user_id FROM refresh_tokens
		 WHERE token = $1 AND expires_at > NOW()`,
		refreshToken,
	).Scan(&userID)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired refresh token")
	}

	// Rotate: delete old token
	s.db.Exec(ctx, `DELETE FROM refresh_tokens WHERE token = $1`, refreshToken)

	return s.issueTokens(ctx, userID)
}

func (s *Service) Logout(ctx context.Context, userID string) error {
	_, err := s.db.Exec(ctx,
		`DELETE FROM refresh_tokens WHERE user_id = $1`, userID)
	return err
}

func (s *Service) issueTokens(ctx context.Context, userID string) (*Tokens, error) {
	expiresAt := time.Now().UTC().Add(time.Hour)

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}).SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return nil, err
	}

	refreshToken := uuid.New().String()
	refreshExpiry := time.Now().UTC().Add(30 * 24 * time.Hour)

	_, err = s.db.Exec(ctx,
		`INSERT INTO refresh_tokens (token, user_id, expires_at)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (token) DO NOTHING`,
		refreshToken, userID, refreshExpiry,
	)
	if err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	return &Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}
