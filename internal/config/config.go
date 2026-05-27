package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Database *Database
}

type Database struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func (d *Database) GetDSN() string {
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

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file, reading from environment")
	}

	cfg := &Config{
		Database: &Database{
			Host:     getString("POSTGRES_HOST", "localhost"),
			Port:     getString("POSTGRES_PORT", "5432"),
			User:     getString("POSTGRES_USER", "postgres"),
			Password: getString("POSTGRES_PASSWORD", ""),
			DBName:   getString("POSTGRES_DB", ""),
			SSLMode:  getString("POSTGRES_SSLMODE", "disable"),
		},
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
