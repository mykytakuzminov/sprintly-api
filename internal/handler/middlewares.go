package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mykytakuzminov/sprintly-api/internal/auth"
	"github.com/mykytakuzminov/sprintly-api/internal/domain"
	"go.uber.org/zap"
)

func AuthMiddleware(a *auth.Auth, logger *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			traceID := getTraceID(r, logger)

			tokenString := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if tokenString == "" {
				logUnauthorizedAccess(logger, traceID)
				errorResponse(w, domain.ErrUnauthorized)
				return
			}

			userID, role, err := a.ParseToken(tokenString)
			if err != nil {
				logUnauthorizedAccess(logger, traceID)
				errorResponse(w, domain.ErrUnauthorized)
				return
			}

			logSuccess(logger, traceID, "authorized", "user_id", userID)
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			ctx = context.WithValue(ctx, UserRoleKey, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireAdminMiddleware(logger *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			traceID := getTraceID(r, logger)

			role, ok := getUserRole(r)
			if !ok {
				logUnexpectedError(logger, traceID, domain.ErrMissingRole)
				errorResponse(w, domain.ErrMissingRole)
				return
			}

			if role != "admin" {
				logInvalidRole(logger, traceID)
				errorResponse(w, domain.ErrForbidden)
				return
			}

			logSuccess(logger, traceID, "", "role", role)
			next.ServeHTTP(w, r)
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

func RateLimiterMiddleware(svc domain.RateLimitService, logger *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			traceID := getTraceID(r, logger)
			var key string

			userID, ok := getUserID(r)
			if !ok {
				ip := getClientIP(r)
				key = fmt.Sprintf("ratelimit:ip:%s", ip)
			} else {
				key = fmt.Sprintf("ratelimit:user:%s", userID.String())
			}

			allowance, err := svc.AllowRequest(r.Context(), key)
			if err != nil {
				logUnexpectedError(logger, traceID, err)
				errorResponse(w, err)
				return
			}
			if !allowance {
				logTooManyRequests(logger, traceID)
				errorResponse(w, domain.ErrTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
