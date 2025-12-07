// cmd/app/main.go
package main

import (
	"context"
	"go-musthave-diploma-tpl/internal/handler"
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
			NewLogger,
			newRouter,
		),
		fx.Invoke(startServer),
	).Run()
}

func NewLogger() (*zap.Logger, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
		return nil, err
	}
	return logger, nil
}

func newRouter(log *zap.Logger) chi.Router {
	if log != nil {
		log.Info("router initialized")
	}
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(customMiddleware.Logger(log))

	r.Get("/health", handler.Health)

	return r
}

// startServer запускает HTTP-сервер с graceful shutdown
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
