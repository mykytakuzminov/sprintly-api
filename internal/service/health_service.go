package service

import (
	"context"
	"time"

	"github.com/mykytakuzminov/sprintly-api/internal/domain"
)

type HealthSvc struct {
	startTime time.Time
	db        domain.DBPinger
	client    domain.RedisPinger
}

func NewHealthService(db domain.DBPinger, client domain.RedisPinger) *HealthSvc {
	return &HealthSvc{startTime: time.Now(), db: db, client: client}
}

func (s *HealthSvc) Check(ctx context.Context) *domain.HealthStats {
	healthStats := &domain.HealthStats{
		Status: "ok",
		DB:     "ok",
		Redis:  "ok",
		Uptime: int64(time.Since(s.startTime).Seconds()),
	}

	if err := s.db.Ping(ctx); err != nil {
		healthStats.Status = "degraded"
		healthStats.DB = "unavailable"
	}

	if err := s.client.Ping(ctx); err != nil {
		healthStats.Status = "degraded"
		healthStats.Redis = "unavailable"
	}

	return healthStats
}
