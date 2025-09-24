package middleware

import (
	"net/http"
	"task-processor/internal/infrastructure/shared/logger"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func LoggerMiddleware(logger *logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			reqID := middleware.GetReqID(r.Context())
			
			next.ServeHTTP(w, r)

			logger.Info("request completed",
				zap.String("request_id", reqID),
				zap.String("method", r.Method),
				zap.String("url", r.URL.String()),
				zap.String("remote_addr", r.RemoteAddr),
				zap.Duration("duration", time.Since(start)),
			)
		}
		return http.HandlerFunc(fn)
	}
}
