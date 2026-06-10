package service

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mykytakuzminov/task-manager-api/internal/domain"
	"github.com/redis/go-redis/v9"
)

type HealthSvc struct {
	startTime time.Time
	pool      *pgxpool.Pool
	client    *redis.Client
}

func NewHealthService(pool *pgxpool.Pool, client *redis.Client) *HealthSvc {
	startTime := time.Now()
	return &HealthSvc{startTime: startTime, pool: pool, client: client}
}

func (s *HealthSvc) Check(ctx context.Context) *domain.HealthStats {
	healthStats := &domain.HealthStats{
		Status: "ok",
		DB:     "ok",
		Redis:  "ok",
		Uptime: int64(time.Since(s.startTime).Seconds()),
	}

	if err := s.pool.Ping(ctx); err != nil {
		healthStats.Status = "degraded"
		healthStats.DB = "unavailable"
	}

	if err := s.client.Ping(ctx).Err(); err != nil {
		healthStats.Status = "degraded"
		healthStats.Redis = "unavailable"
	}

	return healthStats
}
