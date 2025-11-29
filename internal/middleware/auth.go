package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/yourusername/golf_messenger/pkg/jwt"
	"github.com/yourusername/golf_messenger/pkg/response"
)

type contextKey string

const (
	UserIDKey  contextKey = "user_id"
	EmailKey   contextKey = "email"
)

func Auth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Unauthorized(w, "Authorization header required")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				response.Unauthorized(w, "Invalid authorization header format")
				return
			}

			tokenString := parts[1]

			claims, err := jwt.ValidateAccessToken(tokenString, jwtSecret)
			if err != nil {
				if err == jwt.ErrExpiredToken {
					response.Unauthorized(w, "Token has expired")
					return
				}
				response.Unauthorized(w, "Invalid token")
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, EmailKey, claims.Email)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
