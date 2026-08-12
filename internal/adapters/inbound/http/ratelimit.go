package http

import (
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/heitorsfreitass/job-radar/internal/domain"
)

// rateLimitWindow is the fixed window each client's request count is
// bucketed into.
const rateLimitWindow = time.Minute

// RateLimit builds middleware that caps each client IP to limit requests
// per rateLimitWindow, backed by domain.Cache (Redis). It must run after
// middleware.RealIP so r.RemoteAddr reflects the real client IP behind a
// proxy.
func RateLimit(cache domain.Cache, limit int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr
			}

			bucket := time.Now().UTC().Truncate(rateLimitWindow).Unix()
			key := "ratelimit:" + ip + ":" + strconv.FormatInt(bucket, 10)

			count, err := cache.Increment(r.Context(), key, int(rateLimitWindow.Seconds()))
			if err != nil {
				// Fail open: a cache outage should not take the API down.
				next.ServeHTTP(w, r)
				return
			}
			if count > int64(limit) {
				w.Header().Set("Retry-After", strconv.Itoa(int(rateLimitWindow.Seconds())))
				writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
