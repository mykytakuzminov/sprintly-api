package postgres

import (
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/mykytakuzminov/task-manager-api/internal/config"
)

func RunMigrations(cfg *config.Config) error {
	url := cfg.Database.GetDSN()

	m, err := migrate.New("file://migrations", url)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
