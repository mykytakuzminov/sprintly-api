package config

import (
	"fmt"
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

type ServerConfig struct {
	Host string
	Port string
}

func (s *ServerConfig) GetAddr() string {
	return s.Host + ":" + s.Port
}

func (s *ServerConfig) LogFields() []interface{} {
	logFields := []interface{}{
		"server_host", s.Host,
		"server_port", s.Port,
	}

	return logFields
}

func (s *ServerConfig) LogFieldsWithErr(err error) []interface{} {
	logFields := []interface{}{
		"error", err,
	}

	logFields = append(logFields, s.LogFields()...)

	return logFields
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func (d *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		d.User,
		d.Password,
		d.Host,
		d.Port,
		d.DBName,
		d.SSLMode,
	)
}

func (d *DatabaseConfig) LogFields() []interface{} {
	logFields := []interface{}{
		"db_host", d.Host,
		"db_port", d.Port,
		"db_user", d.User,
		"db_name", d.DBName,
		"db_ssl_mode", d.SSLMode,
	}

	return logFields
}

func (d *DatabaseConfig) LogFieldsWithErr(err error) []interface{} {
	logFields := []interface{}{
		"error", err,
	}

	logFields = append(logFields, d.LogFields()...)

	return logFields
}

type JWTConfig struct {
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func (r *RedisConfig) GetAddr() string {
	return r.Host + ":" + r.Port
}

func (r *RedisConfig) LogFields() []interface{} {
	logFields := []interface{}{
		"redis_host", r.Host,
		"redis_port", r.Port,
		"redis_db", r.DB,
	}

	return logFields
}

func (r *RedisConfig) LogFieldsWithErr(err error) []interface{} {
	logFields := []interface{}{
		"error", err,
	}

	logFields = append(logFields, r.LogFields()...)

	return logFields
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
