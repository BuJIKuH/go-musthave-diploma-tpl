// cmd/app/main.go
package main

import (
	"context"
	"fmt"
	"go-musthave-diploma-tpl/internal/handler"
	"go-musthave-diploma-tpl/internal/repository/postgres"
	"go-musthave-diploma-tpl/internal/service"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"go-musthave-diploma-tpl/internal/config"
	customMiddleware "go-musthave-diploma-tpl/internal/middleware"
)

func main() {
	fx.New(
		fx.Provide(
			config.InitConfig,
			newLogger,
			newRouter,
			newStorage,
			NewAuthService,
			NewAuthHandler,
			NewUserRepository,
		),
		fx.Invoke(startServer, startMigrations),
	).Run()
}

func NewUserRepository(store *postgres.DBStorage) *postgres.UserRepository {
	return postgres.NewUserRepository(store.DB)
}

func NewAuthService(repo *postgres.UserRepository, cfg *config.Config, logger *zap.Logger) *service.AuthService {
	return service.NewAuthService(repo, cfg.AuthSecret, logger)
}

func NewAuthHandler(authService *service.AuthService, logger *zap.Logger) *handler.AuthHandler {
	return handler.NewAuthHandler(authService, logger)
}

func startMigrations(cfg *config.Config, logger *zap.Logger) error {
	if err := postgres.RunMigrations(cfg.DatabaseURI, logger); err != nil {
		return fmt.Errorf("migrations failed: %w", err)
	}
	return nil
}

func newStorage(lc fx.Lifecycle, cfg *config.Config, logger *zap.Logger) (*postgres.DBStorage, error) {

	dbStore, err := postgres.NewDBStorage(cfg.DatabaseURI, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %w", err)
	}

	logger.Info("Using PostgreSQL storage")

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info("closing database connection")
			return dbStore.Close()
		},
	})

	return dbStore, nil
}

func newLogger() (*zap.Logger, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
		return nil, err
	}
	return logger, nil
}

func newRouter(cfg *config.Config, logger *zap.Logger, authHandler *handler.AuthHandler) chi.Router {
	if logger != nil {
		logger.Info("router initialized")
	}
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(customMiddleware.Logger(logger))

	r.Get("/health", handler.Health)

	r.Post("/api/user/register", authHandler.Register)
	r.Post("/api/user/login", authHandler.Login)

	r.Group(func(r chi.Router) {
		r.Use(customMiddleware.AuthMiddleware(cfg.AuthSecret, logger))

		// TODO: добавить orderHandler, balanceHandler и т.д.
		// r.Post("/api/user/orders", orderHandler.UploadOrder)
		// r.Get("/api/user/orders", orderHandler.ListOrders)
	})

	return r
}

func startServer(
	lc fx.Lifecycle,
	cfg *config.Config,
	router chi.Router,
	log *zap.Logger,
) {
	srv := &http.Server{
		Addr:    cfg.RunAddress,
		Handler: router,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("starting HTTP server", zap.String("Address", cfg.RunAddress))
			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Error("HTTP server failed", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("shutting down HTTP server")
			return srv.Shutdown(ctx)
		},
	})
}
