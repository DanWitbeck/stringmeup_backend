// internal/users/handler.go
package users

import (
	"encoding/json"
	"net/http"

	"github.com/example/threadcraft-backend/internal/db"
	"github.com/example/threadcraft-backend/internal/middleware"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, svc *Service) {
	r.Get("/users/me", handleGetMe(svc))
	r.Patch("/users/me", handleUpdateMe(svc))
}

func handleGetMe(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := svc.GetByID(r.Context(), middleware.UserID(r))
		if err != nil {
			db.Error(w, http.StatusNotFound, "NOT_FOUND", "user not found")
			return
		}
		db.Data(w, http.StatusOK, user)
	}
}

func handleUpdateMe(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			db.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid body")
			return
		}
		user, err := svc.Update(r.Context(), middleware.UserID(r), body)
		if err != nil {
			db.Error(w, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
			return
		}
		db.Data(w, http.StatusOK, user)
	}
}
