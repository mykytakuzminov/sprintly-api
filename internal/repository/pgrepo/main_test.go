package pgrepo

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/mykytakuzminov/task-manager-api/internal/config"
	"github.com/mykytakuzminov/task-manager-api/internal/infra/pgclient"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	_ = godotenv.Load("../../../.env.test")

	l := zap.Must(zap.NewDevelopment())
	logger := l.Sugar()

	cfg := config.Load(logger)
	ctx := context.Background()

	container, err := startPostgresContainer(ctx, cfg)
	if err != nil {
		logger.Fatalf("start postgres container: %v", err)
	}

	testPool, err = pgclient.NewPool(cfg)
	if err != nil {
		container.Terminate(ctx)
		logger.Fatalf("create pool: %v", err)
	}

	if err = pgclient.RunMigrations(cfg); err != nil {
		container.Terminate(ctx)
		testPool.Close()
		logger.Fatalf("run migrations: %v", err)
	}

	code := m.Run()

	logger.Sync()
	container.Terminate(ctx)
	testPool.Close()

	os.Exit(code)
}

func startPostgresContainer(
	ctx context.Context,
	cfg *config.Config,
) (testcontainers.Container, error) {
	postgresC, err := testcontainers.Run(
		ctx,
		"postgres:18-alpine",
		testcontainers.WithEnv(map[string]string{
			"POSTGRES_USER":     cfg.Database.User,
			"POSTGRES_PASSWORD": cfg.Database.Password,
			"POSTGRES_DB":       cfg.Database.DBName,
		}),
		testcontainers.WithExposedPorts("5432/tcp"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2),
		),
	)
	if err != nil {
		return nil, err
	}

	host, err := postgresC.Host(ctx)
	if err != nil {
		return nil, err
	}

	port, err := postgresC.MappedPort(ctx, "5432/tcp")
	if err != nil {
		return nil, err
	}

	cfg.Database.Host = host
	cfg.Database.Port = port.Port()

	return postgresC, nil
}
