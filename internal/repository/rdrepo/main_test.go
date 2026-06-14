package rdrepo

import (
	"context"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/mykytakuzminov/task-manager-api/internal/config"
	"github.com/mykytakuzminov/task-manager-api/internal/infra/rdclient"
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

var testClient *redis.Client

func TestMain(m *testing.M) {
	_ = godotenv.Load("../../../.env.example")

	l := zap.Must(zap.NewDevelopment())
	logger := l.Sugar()

	cfg := config.Load(logger)
	ctx := context.Background()

	container, err := startRedisContainer(ctx, cfg)
	if err != nil {
		logger.Fatalf("start redis container: %v", err)
	}

	testClient, err = rdclient.NewClient(cfg)
	if err != nil {
		terminateRedisContainer(ctx, container, logger)
		logger.Fatalf("create client: %v", err)
	}

	code := m.Run()

	_ = logger.Sync()
	terminateRedisContainer(ctx, container, logger)
	if err := testClient.Close(); err != nil {
		logger.Errorf("failed to close redis client")
	}

	os.Exit(code)
}

func startRedisContainer(
	ctx context.Context,
	cfg *config.Config,
) (testcontainers.Container, error) {
	redisC, err := testcontainers.Run(
		ctx,
		"redis:8-alpine",
		testcontainers.WithExposedPorts("6379/tcp"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections"),
		),
	)
	if err != nil {
		return nil, err
	}

	host, err := redisC.Host(ctx)
	if err != nil {
		return nil, err
	}

	port, err := redisC.MappedPort(ctx, "6379/tcp")
	if err != nil {
		return nil, err
	}

	cfg.Redis.Host = host
	cfg.Redis.Port = port.Port()

	return redisC, nil
}

func terminateRedisContainer(
	ctx context.Context,
	container testcontainers.Container,
	logger *zap.SugaredLogger,
) {
	if err := container.Terminate(ctx); err != nil {
		logger.Errorf("failed to terminate redis container")
	}
}
