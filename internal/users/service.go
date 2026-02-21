// internal/users/service.go
package users

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Preferences struct {
	DefaultNailStyle      string  `json:"default_nail_style"`
	DefaultNailDiameterMM float64 `json:"default_nail_diameter_mm"`
	Units                 string  `json:"units"`
	AutoSaveProgress      bool    `json:"auto_save_progress"`
	HapticFeedback        bool    `json:"haptic_feedback"`
}

type User struct {
	ID          string      `json:"id"`
	Email       string      `json:"email"`
	Name        string      `json:"name"`
	CreatedAt   time.Time   `json:"created_at"`
	Preferences Preferences `json:"preferences"`
}

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service { return &Service{db: db} }

func (s *Service) GetByID(ctx context.Context, id string) (*User, error) {
	u := &User{}
	err := s.db.QueryRow(ctx,
		`SELECT id, email, name, created_at,
		        COALESCE(pref_nail_style, 'top_mounted'),
		        COALESCE(pref_nail_diameter_mm, 1.5),
		        COALESCE(pref_units, 'metric'),
		        COALESCE(pref_auto_save, true),
		        COALESCE(pref_haptic, false)
		 FROM users WHERE id = $1`, id,
	).Scan(
		&u.ID, &u.Email, &u.Name, &u.CreatedAt,
		&u.Preferences.DefaultNailStyle,
		&u.Preferences.DefaultNailDiameterMM,
		&u.Preferences.Units,
		&u.Preferences.AutoSaveProgress,
		&u.Preferences.HapticFeedback,
	)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	return u, nil
}

func (s *Service) Update(ctx context.Context, id string, updates map[string]any) (*User, error) {
	// Apply known fields
	if name, ok := updates["name"].(string); ok {
		s.db.Exec(ctx, `UPDATE users SET name = $1 WHERE id = $2`, name, id)
	}
	if prefs, ok := updates["preferences"].(map[string]any); ok {
		if v, ok := prefs["default_nail_style"].(string); ok {
			s.db.Exec(ctx, `UPDATE users SET pref_nail_style = $1 WHERE id = $2`, v, id)
		}
		if v, ok := prefs["default_nail_diameter_mm"].(float64); ok {
			s.db.Exec(ctx, `UPDATE users SET pref_nail_diameter_mm = $1 WHERE id = $2`, v, id)
		}
		if v, ok := prefs["units"].(string); ok {
			s.db.Exec(ctx, `UPDATE users SET pref_units = $1 WHERE id = $2`, v, id)
		}
		if v, ok := prefs["auto_save_progress"].(bool); ok {
			s.db.Exec(ctx, `UPDATE users SET pref_auto_save = $1 WHERE id = $2`, v, id)
		}
		if v, ok := prefs["haptic_feedback"].(bool); ok {
			s.db.Exec(ctx, `UPDATE users SET pref_haptic = $1 WHERE id = $2`, v, id)
		}
	}
	return s.GetByID(ctx, id)
}
