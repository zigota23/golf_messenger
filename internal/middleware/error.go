package middleware

import (
	"net/http"

	"github.com/yourusername/golf_messenger/pkg/response"
	"go.uber.org/zap"
)

func ErrorRecovery(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("panic recovered",
						zap.Any("error", err),
						zap.String("method", r.Method),
						zap.String("path", r.URL.Path),
					)

					response.InternalServerError(w, "Internal server error")
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
