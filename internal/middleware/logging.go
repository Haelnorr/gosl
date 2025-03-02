package middleware

import (
	"gosl/pkg/contexts"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

// Wraps the http.ResponseWriter, adding a statusCode field
type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

// Extends WriteHeader to the ResponseWriter to add the status code
func (w *wrappedWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}

// Middleware to add logs to console with details of the request
func Logging(logger *zerolog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/static/css/output.css" ||
			r.URL.Path == "/static/favicon.ico" {
			next.ServeHTTP(w, r)
			return
		}
		start, err := contexts.GetStartTime(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		wrapped := &wrappedWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		next.ServeHTTP(wrapped, r)
		logger.Info().
			Int("status", wrapped.statusCode).
			Str("method", r.Method).
			Str("resource", r.URL.Path).
			Dur("time_elapsed", time.Since(start)).
			Str("remote_addr", r.Header.Get("X-Forwarded-For")).
			Msg("Served")
	})
}
