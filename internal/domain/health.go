package domain

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type HealthStats struct {
	Status string `json:"status" example:"ok"`
	DB     string `json:"db"     example:"ok"`
	Redis  string `json:"redis"  example:"ok"`
	Uptime int64  `json:"uptime" example:"3600"`
}

type DBPinger interface {
	Ping(ctx context.Context) error
}

type RedisPinger interface {
	Ping(ctx context.Context) error
}

type RedisClientWrapper struct {
	client *redis.Client
}

func NewRedisClientWrapper(client *redis.Client) *RedisClientWrapper {
	return &RedisClientWrapper{client: client}
}

func (w *RedisClientWrapper) Ping(ctx context.Context) error {
	return w.client.Ping(ctx).Err()
}

type HealthService interface {
	Check(ctx context.Context) *HealthStats
}
