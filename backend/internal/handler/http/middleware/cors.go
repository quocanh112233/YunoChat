package middleware

import (
	"net/http"
	"strings"
)

// CORS handles Cross-Origin Resource Sharing logic
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			allowed := false
			isWildcard := len(allowedOrigins) == 1 && allowedOrigins[0] == "*"

			if isWildcard {
				allowed = true
			} else {
				for _, o := range allowedOrigins {
					if strings.EqualFold(o, origin) {
						allowed = true
						break
					}
				}
			}

			if allowed {
				// If it's a wildcard, we shouldn't reflect the origin IF we want to allow credentials
				// Browsers don't allow "*" with credentials.
				// If we reflect the origin, we are effectively allowing any origin with credentials.
				if isWildcard {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
				
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
				w.Header().Set("Access-Control-Expose-Headers", "Link")
				w.Header().Set("Access-Control-Max-Age", "300")
			}

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
