package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   *ServerConfig
	Database *DatabaseConfig
	JWT      *JWTConfig
}

type ServerConfig struct {
	Host string
	Port string
}

func (s *ServerConfig) GetAddr() string {
	return s.Host + ":" + s.Port
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

type JWTConfig struct {
	Secret     string
	AccessTTL  string
	RefreshTTL string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file, reading from environment")
	}

	cfg := &Config{
		Database: &DatabaseConfig{
			Host:     getString("POSTGRES_HOST", "localhost"),
			Port:     getString("POSTGRES_PORT", "5432"),
			User:     getString("POSTGRES_USER", "postgres"),
			Password: getString("POSTGRES_PASSWORD", ""),
			DBName:   getString("POSTGRES_DB", ""),
			SSLMode:  getString("POSTGRES_SSLMODE", "disable"),
		},
		Server: &ServerConfig{
			Host: getString("SERVER_HOST", "localhost"),
			Port: getString("SERVER_PORT", "8080"),
		},
		JWT: &JWTConfig{
			Secret:     getString("JWT_SECRET", ""),
			AccessTTL:  getString("JWT_ACCESS_TTL", "15m"),
			RefreshTTL: getString("JWT_REFRESH_TTL", "168h"),
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

	return cfg
}

func getString(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
