package handler

import (
	"log/slog"
	"net/http"
)

// recoverPanics converts a handler panic into a logged, styled 500
// instead of a dropped connection.
func recoverPanics(log *slog.Logger, on500 http.HandlerFunc, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Error("panic recovered", "path", r.URL.Path, "panic", rec)
				on500(w, r)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
