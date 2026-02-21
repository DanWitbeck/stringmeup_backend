// internal/progress/handler.go
package progress

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"stringmeup/backend/internal/db"
	"stringmeup/backend/internal/middleware"
)

func HandleGet(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, err := svc.Get(r.Context(), chi.URLParam(r, "id"), middleware.UserID(r))
		if err != nil {
			db.Error(w, http.StatusNotFound, "NOT_FOUND", "progress not found")
			return
		}
		db.Data(w, http.StatusOK, p)
	}
}

func HandlePut(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var p Progress
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			db.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid body")
			return
		}
		result, err := svc.Put(r.Context(), chi.URLParam(r, "id"), middleware.UserID(r), &p)
		if err != nil {
			db.Error(w, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
			return
		}
		db.Data(w, http.StatusOK, result)
	}
}
