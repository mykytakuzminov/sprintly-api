package domain

import "context"

type HealthStats struct {
    Status string `json:"status" example:"ok"`
    DB     string `json:"db"     example:"ok"`
    Redis  string `json:"redis"  example:"ok"`
    Uptime int64  `json:"uptime" example:"3600"`
}

type HealthService interface {
	Check(ctx context.Context) *HealthStats
}
