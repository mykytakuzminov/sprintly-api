package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mykytakuzminov/task-manager-api/internal/config"
	"github.com/mykytakuzminov/task-manager-api/internal/repository/postgres"
	"github.com/mykytakuzminov/task-manager-api/internal/server"
)

type App struct {
	cfg    *config.Config
	pool   *pgxpool.Pool
	server *server.Server
}

func New() *App {
	cfg := config.Load()

	pool, err := postgres.NewPool(cfg)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	log.Println("database connected successfully")

	if err := postgres.RunMigrations(cfg); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}
	log.Println("migrations applied successfully")

	router := chi.NewRouter()

	return &App{
		cfg:    cfg,
		pool:   pool,
		server: server.New(cfg.Server.GetAddr(), router),
	}
}

func (a *App) Run() error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- a.server.Run()
	}()
	log.Println("server is running")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			return err
		}
	case <-quit:
		log.Println("shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		defer a.pool.Close()
		return a.server.Shutdown(ctx)
	}

	return nil
}
