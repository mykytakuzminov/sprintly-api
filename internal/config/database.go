package config

import (
	"fmt"
)

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
