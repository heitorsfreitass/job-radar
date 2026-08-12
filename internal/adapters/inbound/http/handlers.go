package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/heitorsfreitass/job-radar/internal/application"
	"github.com/heitorsfreitass/job-radar/internal/domain"
)

type searchResponse struct {
	Data []*domain.Job `json:"data"`
	Meta pageMeta      `json:"meta"`
}

type pageMeta struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Total    int `json:"total"`
}

// handleSearchJobs handles GET /jobs, translating query params into a
// domain.JobFilter and delegating to application.SearchJobs.
func (h *Handler) handleSearchJobs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	filter := domain.JobFilter{
		Country:   q.Get("country"),
		Workplace: domain.WorkplaceType(q.Get("workplace")),
		Seniority: domain.SeniorityLevel(q.Get("seniority")),
		Tag:       q.Get("tag"),
		Keyword:   q.Get("keyword"),
		Page:      atoiOrDefault(q.Get("page"), 1),
		PageSize:  atoiOrDefault(q.Get("page_size"), 20),
	}

	result, err := application.SearchJobs(r.Context(), h.repo, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to search jobs")
		return
	}

	jobs := result.Jobs
	if jobs == nil {
		jobs = []*domain.Job{}
	}

	writeJSON(w, http.StatusOK, searchResponse{
		Data: jobs,
		Meta: pageMeta{Page: result.Page, PageSize: result.PageSize, Total: result.Total},
	})
}

// handleGetJob handles GET /jobs/{id}.
func (h *Handler) handleGetJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	job, err := application.GetJob(r.Context(), h.repo, id)
	if errors.Is(err, application.ErrJobNotFound) {
		writeError(w, http.StatusNotFound, "job not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get job")
		return
	}

	writeJSON(w, http.StatusOK, job)
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func atoiOrDefault(s string, fallback int) int {
	if s == "" {
		return fallback
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return n
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
