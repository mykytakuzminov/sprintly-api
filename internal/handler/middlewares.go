package handler

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mykytakuzminov/task-manager-api/internal/auth"
	"go.uber.org/zap"
)

type contextKey struct{}

var UserIDKey = contextKey{}
var TraceIDKey = contextKey{}

func AuthMiddleware(a *auth.Auth, logger *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			traceID := getTraceID(r, logger)

			tokenString := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if tokenString == "" {
				logUnauthorizedAccess(logger, traceID)
				return
			}

			userID, err := a.ParseToken(tokenString)
			if err != nil {
				logUnauthorizedAccess(logger, traceID)
				return
			}

			logSuccess(logger, traceID, "authorized", "user_id", userID)
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func TraceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := uuid.New()

		w.Header().Set("X-Trace-ID", traceID.String())

		ctx := context.WithValue(r.Context(), TraceIDKey, traceID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func TimeoutMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
