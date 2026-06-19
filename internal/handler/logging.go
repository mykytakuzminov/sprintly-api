package handler

import (
	"github.com/google/uuid"
	"github.com/mykytakuzminov/sprintly-api/internal/domain"
	"go.uber.org/zap"
)

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

func logInvalidRole(logger *zap.SugaredLogger, traceID uuid.UUID) {
	logger.Warnw(
		"invalid user role",
		"trace_id", traceID,
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
