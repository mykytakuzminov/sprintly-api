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
	"github.com/mykytakuzminov/task-manager-api/internal/handler/middleware"
	"github.com/mykytakuzminov/task-manager-api/internal/infra/pgclient"
	"github.com/mykytakuzminov/task-manager-api/internal/infra/rdclient"
	"github.com/mykytakuzminov/task-manager-api/internal/repository/pgrepo"
	"github.com/mykytakuzminov/task-manager-api/internal/repository/rdrepo"
	"github.com/mykytakuzminov/task-manager-api/internal/server"
	"github.com/mykytakuzminov/task-manager-api/internal/service"
	"github.com/redis/go-redis/v9"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

type App struct {
	cfg    *config.Config
	logger *zap.SugaredLogger
	pool   *pgxpool.Pool
	client *redis.Client
	server *server.Server
}

func New() *App {
	l := zap.Must(zap.NewProduction())
	logger := l.Sugar()

	cfg := config.Load(logger)
	auth := auth.NewAuth(cfg.JWT)

	pool := initDB(cfg, logger)
	client := initRedis(cfg, logger)
	router := initRouter(logger, pool, client, auth)

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

func initDB(cfg *config.Config, logger *zap.SugaredLogger) *pgxpool.Pool {
	pool, err := pgclient.NewPool(cfg)
	if err != nil {
		logger.Fatalw("database connection failed",
			cfg.Database.LogFieldsWithErr(err)...,
		)
	}
	logger.Infow("database connected successfully",
		cfg.Database.LogFields()...,
	)

	if err := pgclient.RunMigrations(cfg); err != nil {
		logger.Fatalw("migrations failed",
			"error", err,
		)
	}
	logger.Infow("migrations applied successfully")

	return pool
}

func initRedis(cfg *config.Config, logger *zap.SugaredLogger) *redis.Client {
	client, err := rdclient.NewClient(cfg)
	if err != nil {
		logger.Fatalw("redis connection failed",
			cfg.Redis.LogFieldsWithErr(err)...,
		)
	}
	logger.Infow("redis connected successfully",
		cfg.Redis.LogFields()...,
	)

	return client
}

func initRouter(
	logger *zap.SugaredLogger,
	pool *pgxpool.Pool,
	client *redis.Client,
	auth *auth.Auth,
) chi.Router {
	router := chi.NewRouter()

	router.Get("/swagger/*", httpSwagger.WrapHandler)

	userRepo := pgrepo.NewUserRepository(pool)
	userSvc := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userSvc, auth, logger)

	tokenRepo := rdrepo.NewTokenRepository(client)
	authSvc := service.NewAuthService(userRepo, tokenRepo, auth)
	authHandler := handler.NewAuthHandler(authSvc, logger)

	boardRepo := pgrepo.NewBoardRepository(pool)
	boardSvc := service.NewBoardService(boardRepo)
	boardHandler := handler.NewBoardHandler(boardSvc, auth, logger)

	columnRepo := pgrepo.NewColumnRepository(pool)
	columnSvc := service.NewColumnService(columnRepo, boardRepo)
	columnHandler := handler.NewColumnHandler(columnSvc, auth, logger)

	taskRepo := pgrepo.NewTaskRepository(pool)
	taskSvc := service.NewTaskService(taskRepo, columnRepo)
	taskHandler := handler.NewTaskHandler(taskSvc, auth)

	router.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.TraceMiddleware)

		r.Mount("/auth", authHandler.Routes())
		r.Mount("/users", userHandler.Routes())
		r.Mount("/users/me/tasks", taskHandler.UserRoutes())

		r.Mount("/boards", boardHandler.Routes())

		r.Mount("/boards/{boardID}/columns", columnHandler.BoardRoutes())
		r.Mount("/columns", columnHandler.Routes())

		r.Mount("/columns/{columnID}/tasks", taskHandler.ColumnRoutes())
		r.Mount("/tasks", taskHandler.Routes())
	})

	return router
}
