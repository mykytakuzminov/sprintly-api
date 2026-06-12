package handler

import (
	"net"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mykytakuzminov/task-manager-api/internal/domain"
	"go.uber.org/zap"
)

func getUserID(r *http.Request) (uuid.UUID, bool) {
	userID, ok := r.Context().Value(UserIDKey).(uuid.UUID)
	return userID, ok
}

func getTraceID(r *http.Request, logger *zap.SugaredLogger) uuid.UUID {
	traceID, ok := r.Context().Value(TraceIDKey).(uuid.UUID)

	if !ok {
		traceID = uuid.Nil
		logger.Warnw("trace id missing")
	}

	return traceID
}

func getURLParam(r *http.Request, key string) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, key))
}

func getClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func logInvalidBody(logger *zap.SugaredLogger, traceID uuid.UUID, err error) {
	logger.Warnw(
		"invalid request body",
		"trace_id", traceID,
		"error", err,
	)
}

func logUnexpectedError(logger *zap.SugaredLogger, traceID uuid.UUID, err error) {
	logger.Errorw(
		"unexpected error occured",
		"trace_id", traceID,
		"error", err,
	)
}

func logUnauthorizedAccess(logger *zap.SugaredLogger, traceID uuid.UUID) {
	logger.Warnw(
		"unauthorized access",
		"trace_id", traceID,
		"error", domain.ErrUnauthorized,
	)
}

func logTooManyRequests(logger *zap.SugaredLogger, traceID uuid.UUID) {
	logger.Warnw(
		"too many requests",
		"trace_id", traceID,
		"error", domain.ErrTooManyRequests,
	)
}

func logWarn(logger *zap.SugaredLogger, traceID uuid.UUID, msg string, err error) {
	logger.Warnw(
		msg,
		"trace_id", traceID,
		"error", err,
	)
}

func logSuccess(logger *zap.SugaredLogger, traceID uuid.UUID, msg string, fields ...interface{}) {
	logger.Infow(
		msg,
		append([]interface{}{"trace_id", traceID}, fields...)...,
	)
}

func parseListParams(r *http.Request) *domain.ListParams {
	q := r.URL.Query()

	limit, err := strconv.Atoi(q.Get("limit"))
	if err != nil || limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset, err := strconv.Atoi(q.Get("offset"))
	if err != nil || offset < 0 {
		offset = 0
	}

	return &domain.ListParams{
		Limit:  limit,
		Offset: offset,
		SortBy: q.Get("sort"),
		Order:  q.Get("order"),
	}
}
