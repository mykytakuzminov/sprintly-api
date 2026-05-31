package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/mykytakuzminov/task-manager-api/internal/auth"
)

const UserIDKey = "userID"

func AuthMiddleware(a *auth.Auth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if tokenString == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			userID, err := a.ParseToken(tokenString)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
