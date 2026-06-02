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
	_ "github.com/mykytakuzminov/task-manager-api/docs"
	"github.com/mykytakuzminov/task-manager-api/internal/auth"
	"github.com/mykytakuzminov/task-manager-api/internal/config"
	"github.com/mykytakuzminov/task-manager-api/internal/handler"
	"github.com/mykytakuzminov/task-manager-api/internal/repository/postgres"
	redistore "github.com/mykytakuzminov/task-manager-api/internal/repository/redis"
	"github.com/mykytakuzminov/task-manager-api/internal/server"
	"github.com/mykytakuzminov/task-manager-api/internal/service"
	goredis "github.com/redis/go-redis/v9"
	httpSwagger "github.com/swaggo/http-swagger"
)

type App struct {
	cfg    *config.Config
	pool   *pgxpool.Pool
	client *goredis.Client
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

	client, err := redistore.NewClient(cfg)
	if err != nil {
		log.Fatalf("redis connection failed: %v", err)
	}
	log.Println("redis connected successfully")

	auth := auth.NewAuth(cfg.JWT)

	router := chi.NewRouter()

	router.Get("/swagger/*", httpSwagger.WrapHandler)

	userRepo := postgres.NewUserRepository(pool)
	userSvc := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userSvc, auth)

	tokenRepo := redistore.NewTokenRepository(client)
	authSvc := service.NewAuthService(userRepo, tokenRepo, auth)
	authHandler := handler.NewAuthHandler(authSvc)

	boardRepo := postgres.NewBoardRepository(pool)
	boardSvc := service.NewBoardService(boardRepo)
	boardHandler := handler.NewBoardHandler(boardSvc, auth)

	columnRepo := postgres.NewColumnRepository(pool)
	columnSvc := service.NewColumnService(columnRepo, boardRepo)
	columnHandler := handler.NewColumnHandler(columnSvc, auth)

	taskRepo := postgres.NewTaskRepository(pool)
	taskSvc := service.NewTaskService(taskRepo, columnRepo)
	taskHandler := handler.NewTaskHandler(taskSvc, auth)

	router.Mount("/api/v1/users", userHandler.Routes())
	router.Mount("/api/v1/auth", authHandler.Routes())
	router.Mount("/api/v1/boards", boardHandler.Routes())
	router.Mount("/api/v1/boards/{boardID}/columns", columnHandler.BoardRoutes())
	router.Mount("/api/v1/columns", columnHandler.Routes())
	router.Mount("/api/v1/tasks", taskHandler.Routes())
	router.Mount("/api/v1/columns/{columnID}/tasks", taskHandler.ColumnRoutes())
	router.Mount("/api/v1/users/me/tasks", taskHandler.UserRoutes())

	return &App{
		cfg:    cfg,
		pool:   pool,
		client: client,
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
		defer a.client.Close()
		return a.server.Shutdown(ctx)
	}

	return nil
}
