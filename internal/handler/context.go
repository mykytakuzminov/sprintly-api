package handler

import (
	"net/http"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type userIDKeyType struct{}
type userRoleKeyType struct{}
type traceIDKeyType struct{}

var UserIDKey = userIDKeyType{}
var UserRoleKey = userRoleKeyType{}
var TraceIDKey = traceIDKeyType{}

func getUserID(r *http.Request) (uuid.UUID, bool) {
	userID, ok := r.Context().Value(UserIDKey).(uuid.UUID)
	return userID, ok
}

func getUserRole(r *http.Request) (string, bool) {
	role, ok := r.Context().Value(UserRoleKey).(string)
	return role, ok
}

func getTraceID(r *http.Request, logger *zap.SugaredLogger) uuid.UUID {
	traceID, ok := r.Context().Value(TraceIDKey).(uuid.UUID)

	if !ok {
		traceID = uuid.Nil
		logger.Warnw("trace id missing")
	}

	return traceID
}
