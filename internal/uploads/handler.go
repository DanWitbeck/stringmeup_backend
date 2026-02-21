// internal/uploads/handler.go
package uploads

import (
	"encoding/json"
	"net/http"

	"github.com/example/threadcraft-backend/internal/db"
	"github.com/example/threadcraft-backend/internal/middleware"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, svc *Service) {
	r.Post("/uploads/presign", handlePresign(svc))
}

func handlePresign(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			ContentType string `json:"content_type"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ContentType == "" {
			body.ContentType = "image/jpeg"
		}

		result, err := svc.Presign(r.Context(), middleware.UserID(r), body.ContentType)
		if err != nil {
			db.Error(w, http.StatusInternalServerError, "SERVER_ERROR", "could not generate upload URL")
			return
		}
		db.Data(w, http.StatusOK, result)
	}
}
