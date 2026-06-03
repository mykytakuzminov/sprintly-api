package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   *ServerConfig
	Database *DatabaseConfig
	JWT      *JWTConfig
	Redis    *RedisConfig
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file, reading from environment")
	}

	cfg := &Config{
		Server: &ServerConfig{
			Host: getString("SERVER_HOST", "localhost"),
			Port: getString("SERVER_PORT", "8080"),
		},
		Database: &DatabaseConfig{
			Host:     getString("POSTGRES_HOST", "localhost"),
			Port:     getString("POSTGRES_PORT", "5432"),
			User:     getString("POSTGRES_USER", "postgres"),
			Password: getString("POSTGRES_PASSWORD", ""),
			DBName:   getString("POSTGRES_DB", ""),
			SSLMode:  getString("POSTGRES_SSLMODE", "disable"),
		},
		JWT: &JWTConfig{
			Secret:     getString("JWT_SECRET", ""),
			AccessTTL:  getDuration("JWT_ACCESS_TTL", 15*time.Minute),
			RefreshTTL: getDuration("JWT_REFRESH_TTL", 168*time.Hour),
		},
		Redis: &RedisConfig{
			Host:     getString("REDIS_HOST", "localhost"),
			Port:     getString("REDIS_PORT", "6379"),
			Password: getString("REDIS_PASSWORD", ""),
			DB:       getInt("REDIS_DB", 0),
		},
	}

	if cfg.Database.Password == "" {
		log.Fatal("POSTGRES_PASSWORD is required")
	}
	if cfg.Database.DBName == "" {
		log.Fatal("POSTGRES_DB is required")
	}
	if cfg.JWT.Secret == "" {
		log.Fatal("JWT_SECRET is required")
	}
	if cfg.Redis.Password == "" {
		log.Fatal("REDIS_PASSWORD is required")
	}

	return cfg
}

func getString(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}

func getInt(key string, fallback int) int {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}

	vali, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	return vali
}

func getDuration(key string, fallback time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}

	vald, err := time.ParseDuration(val)
	if err != nil {
		return fallback
	}
	return vald
}
