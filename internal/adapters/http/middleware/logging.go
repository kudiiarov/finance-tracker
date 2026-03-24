// Package middleware provides HTTP middleware.
package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/sereozha/finance-tracker/internal/ports/logger"
	"github.com/sereozha/finance-tracker/pkg/contextkey"
)

// RequestLogger logs HTTP requests with structured fields.
func RequestLogger(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Generate or extract correlation ID
			correlationID := r.Header.Get("X-Correlation-ID")
			ctx := contextkey.WithCorrelationID(r.Context(), correlationID)
			r = r.WithContext(ctx)

			// Add correlation ID to response header
			w.Header().Set("X-Correlation-ID", contextkey.CorrelationID(ctx))

			// Wrap response writer to capture status code
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// Process request
			next.ServeHTTP(ww, r)

			// Log request completion
			duration := time.Since(start)
			log.Info("http_request",
				logger.String(logger.FieldCorrelationID, contextkey.CorrelationID(ctx)),
				logger.String(logger.FieldMethod, r.Method),
				logger.String(logger.FieldPath, r.URL.Path),
				logger.String("remote_addr", r.RemoteAddr),
				logger.Int("status", ww.Status()),
				logger.Int64(logger.FieldDuration, duration.Milliseconds()),
				logger.Int("bytes_written", ww.BytesWritten()),
				logger.String("user_agent", r.UserAgent()),
			)
		})
	}
}

// RequestID adds correlation ID to context (backward compatibility).
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := r.Header.Get("X-Correlation-ID")
		ctx := contextkey.WithCorrelationID(r.Context(), correlationID)
		w.Header().Set("X-Correlation-ID", contextkey.CorrelationID(ctx))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// UserID middleware adds user ID to request context.
func UserID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Extract from JWT token when auth is implemented
		ctx := contextkey.WithUserID(r.Context(), "user-123")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
