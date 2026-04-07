package middleware

import (
	"context"
	"net/http"
	"strings"

	"example.com/go-master-web-sample/internal/auth"
	"example.com/go-master-web-sample/internal/httpx"
)

type contextKey string

const claimsKey contextKey = "authClaims"

func RequireAuth(manager *auth.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				httpx.Error(w, http.StatusUnauthorized, "missing bearer token", nil)
				return
			}

			claims, err := manager.ParseToken(strings.TrimPrefix(header, "Bearer "))
			if err != nil {
				httpx.Error(w, http.StatusUnauthorized, "invalid token", err.Error())
				return
			}

			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())
			if !ok {
				httpx.Error(w, http.StatusUnauthorized, "missing auth context", nil)
				return
			}

			if claims.Role != role {
				httpx.Error(w, http.StatusForbidden, "insufficient role", map[string]string{"required": role})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func ClaimsFromContext(ctx context.Context) (*auth.Claims, bool) {
	claims, ok := ctx.Value(claimsKey).(*auth.Claims)
	return claims, ok
}
