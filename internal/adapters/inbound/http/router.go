// Package http contains the inbound HTTP adapter: router, handlers, and
// middleware. Handlers depend only on the application layer's use cases,
// never on the outbound adapters directly.
package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/heitorsfreitass/job-radar/internal/domain"
)

// defaultRateLimitPerMinute caps each client IP to 60 requests/minute
// against the API, backed by Redis via RateLimit.
const defaultRateLimitPerMinute = 60

// Handler holds the dependencies shared by the HTTP handlers.
type Handler struct {
	repo domain.JobRepository
}

// NewRouter builds the HTTP router: job search/filter/pagination routes,
// backed by repo, rate-limited per client IP via cache. frontendOrigin is
// the only origin allowed to call the API cross-origin (see internal/config).
func NewRouter(repo domain.JobRepository, cache domain.Cache, frontendOrigin string) http.Handler {
	h := &Handler{repo: repo}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{frontendOrigin},
		AllowedMethods: []string{http.MethodGet},
	}))
	r.Use(RateLimit(cache, defaultRateLimitPerMinute))

	r.Get("/healthz", handleHealthz)
	r.Get("/jobs", h.handleSearchJobs)
	r.Get("/jobs/{id}", h.handleGetJob)

	return r
}
