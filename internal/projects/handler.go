// internal/projects/handler.go
package projects

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"stringmeup/backend/internal/db"
	"stringmeup/backend/internal/middleware"
	"stringmeup/backend/internal/progress"
)

func RegisterRoutes(r chi.Router, svc *Service, progressSvc *progress.Service) {
	r.Get("/projects", handleList(svc))
	r.Post("/projects", handleCreate(svc))

	r.Route("/projects/{id}", func(r chi.Router) {
		r.Get("/", handleGet(svc))
		r.Patch("/", handleUpdate(svc))
		r.Delete("/", handleDelete(svc))
		r.Get("/export", handleExport(svc))
		r.Get("/progress", progress.HandleGet(progressSvc))
		r.Put("/progress", progress.HandlePut(progressSvc))
	})
}

func handleList(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if page < 1 {
			page = 1
		}
		if limit < 1 || limit > 100 {
			limit = 20
		}

		projects, meta, err := svc.List(r.Context(), middleware.UserID(r), page, limit)
		if err != nil {
			db.Error(w, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
			return
		}
		db.JSON(w, http.StatusOK, map[string]any{"data": projects, "meta": meta})
	}
}

func handleCreate(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			db.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid body")
			return
		}
		p, err := svc.Create(r.Context(), middleware.UserID(r), body)
		if err != nil {
			db.Error(w, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
			return
		}
		db.Data(w, http.StatusCreated, p)
	}
}

func handleGet(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, err := svc.GetByID(r.Context(), chi.URLParam(r, "id"), middleware.UserID(r))
		if err != nil {
			db.Error(w, http.StatusNotFound, "NOT_FOUND", "project not found")
			return
		}
		db.Data(w, http.StatusOK, p)
	}
}

func handleUpdate(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			db.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid body")
			return
		}
		p, err := svc.Update(r.Context(), chi.URLParam(r, "id"), middleware.UserID(r), body)
		if err != nil {
			db.Error(w, http.StatusNotFound, "NOT_FOUND", "project not found")
			return
		}
		db.Data(w, http.StatusOK, p)
	}
}

func handleDelete(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		svc.Delete(r.Context(), chi.URLParam(r, "id"), middleware.UserID(r))
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleExport(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		format := r.URL.Query().Get("format")
		if format == "" {
			format = "txt"
		}
		content, err := svc.Export(r.Context(), chi.URLParam(r, "id"), middleware.UserID(r), format)
		if err != nil {
			db.Error(w, http.StatusNotFound, "NOT_FOUND", "project not found")
			return
		}
		db.Data(w, http.StatusOK, map[string]string{"content": content})
	}
}
