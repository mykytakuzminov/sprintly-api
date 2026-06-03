package app

import (
	"context"
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
	"go.uber.org/zap"
)

type App struct {
	cfg    *config.Config
	logger *zap.SugaredLogger
	pool   *pgxpool.Pool
	client *goredis.Client
	server *server.Server
}

func New() *App {
	cfg := config.Load()

	l := zap.Must(zap.NewProduction())
	logger := l.Sugar()

	pool, err := postgres.NewPool(cfg)
	if err != nil {
		logger.Fatalw("database connection failed",
			cfg.Database.LogFieldsWithErr(err)...,
		)
	}
	logger.Infow("database connected successfully",
		cfg.Database.LogFields()...,
	)

	if err := postgres.RunMigrations(cfg); err != nil {
		logger.Fatalw("migrations failed",
			"error", err,
		)
	}
	logger.Infow("migrations applied successfully")

	client, err := redistore.NewClient(cfg)
	if err != nil {
		logger.Fatalw("redis connection failed",
			cfg.Redis.LogFieldsWithErr(err)...,
		)
	}
	logger.Infow("redis connected successfully",
		cfg.Redis.LogFields()...,
	)

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

	router.Mount("/api/v1/auth", authHandler.Routes())
	router.Mount("/api/v1/users", userHandler.Routes())
	router.Mount("/api/v1/users/me/tasks", taskHandler.UserRoutes())

	router.Mount("/api/v1/boards", boardHandler.Routes())

	router.Mount("/api/v1/boards/{boardID}/columns", columnHandler.BoardRoutes())
	router.Mount("/api/v1/columns", columnHandler.Routes())

	router.Mount("/api/v1/columns/{columnID}/tasks", taskHandler.ColumnRoutes())
	router.Mount("/api/v1/tasks", taskHandler.Routes())

	return &App{
		cfg:    cfg,
		logger: logger,
		pool:   pool,
		client: client,
		server: server.New(cfg.Server.GetAddr(), router),
	}
}

func (a *App) Run() {
	defer a.logger.Sync()
	defer a.pool.Close()
	defer a.client.Close()

	errCh := make(chan error, 1)
	go func() {
		errCh <- a.server.Run()
	}()
	a.logger.Infow("server started",
		a.cfg.Server.LogFields()...,
	)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			a.logger.Errorw("unexpected error occured",
				a.cfg.Server.LogFieldsWithErr(err)...,
			)
		}
	case <-quit:
		a.logger.Infow("shutting down the server")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.Errorw("shutting down error occured",
			a.cfg.Server.LogFieldsWithErr(err)...,
		)
	}

	a.logger.Infow("server stopped")
}
