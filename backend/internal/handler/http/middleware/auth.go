package middleware

import (
	"context"
	"net/http"
	"strings"

	"backend/internal/pkg/jwt"
	"backend/internal/pkg/response"
)

// RequireAuth middleware verifies the JWT token present in the Authorization header.
func RequireAuth(accessSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Err(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				response.Err(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid authorization header format")
				return
			}

			tokenStr := parts[1]
			userID, err := jwt.ParseToken(tokenStr, accessSecret)
			if err != nil {
				response.Err(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired access token")
				return
			}

			ctx := context.WithValue(r.Context(), "user_id", userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
