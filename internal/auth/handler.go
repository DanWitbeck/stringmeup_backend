// internal/auth/handler.go
package auth

import (
	"encoding/json"
	"net/http"

	"github.com/example/threadcraft-backend/internal/db"
	"github.com/example/threadcraft-backend/internal/middleware"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, svc *Service) {
	r.Post("/auth/register", handleRegister(svc))
	r.Post("/auth/login", handleLogin(svc))
	r.Post("/auth/refresh", handleRefresh(svc))
	r.Delete("/auth/logout", middleware.Authenticate(svc.cfg.JWTSecret)(
		http.HandlerFunc(handleLogout(svc))).ServeHTTP)
}

func handleRegister(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
			Name     string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			db.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid body")
			return
		}
		if body.Email == "" || body.Password == "" || body.Name == "" {
			db.Error(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "email, password and name required")
			return
		}

		user, tokens, err := svc.Register(r.Context(), body.Email, body.Password, body.Name)
		if err != nil {
			db.Error(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
			return
		}
		db.Data(w, http.StatusCreated, map[string]any{
			"user":   user,
			"tokens": tokens,
		})
	}
}

func handleLogin(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			db.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid body")
			return
		}

		user, tokens, err := svc.Login(r.Context(), body.Email, body.Password)
		if err != nil {
			db.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid credentials")
			return
		}
		db.Data(w, http.StatusOK, map[string]any{
			"user":   user,
			"tokens": tokens,
		})
	}
}

func handleRefresh(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.RefreshToken == "" {
			db.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "refresh_token required")
			return
		}

		tokens, err := svc.Refresh(r.Context(), body.RefreshToken)
		if err != nil {
			db.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
			return
		}
		db.Data(w, http.StatusOK, map[string]any{"tokens": tokens})
	}
}

func handleLogout(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserID(r)
		svc.Logout(r.Context(), userID)
		w.WriteHeader(http.StatusNoContent)
	}
}
