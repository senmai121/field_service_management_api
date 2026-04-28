package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"field_service_management_api/internal/auth"
)

type contextKey string

const claimsKey contextKey = "claims"

// jsonError writes a JSON error response — used internally to avoid import cycles.
func jsonError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// JWTAuth returns middleware that validates the Authorization: Bearer <token> header.
// On success, the parsed Claims are stored in the request context.
// On failure, a 401 JSON error is returned immediately.
func JWTAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				jsonError(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				jsonError(w, http.StatusUnauthorized, "invalid authorization header format")
				return
			}

			claims, err := auth.ValidateToken(parts[1], secret)
			if err != nil {
				jsonError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetClaims retrieves the JWT Claims stored in the request context by JWTAuth.
// Returns nil if no claims are present (i.e., the request bypassed auth middleware).
func GetClaims(r *http.Request) *auth.Claims {
	claims, _ := r.Context().Value(claimsKey).(*auth.Claims)
	return claims
}
