package main

import (
	"context"
	"fmt"
	"go-musthave-diploma-tpl/internal/accrual"
	"go-musthave-diploma-tpl/internal/config"
	"go-musthave-diploma-tpl/internal/handler"
	"go-musthave-diploma-tpl/internal/repository/postgres"
	"go-musthave-diploma-tpl/internal/service"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/fx"
	"go.uber.org/zap"

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

			NewOrderRepository,
			NewOrdersService,
			NewOrdersHandler,

			NewWithdrawalRepository,
			NewBalanceService,
			NewBalanceHandler,

			NewAccrualClient,
			NewAccrualWorker,
		),
		fx.Invoke(startServer, startMigrations, StartAccrualWorker),
	).Run()
}

// ------------------------ Providers ------------------------

func newLogger() (*zap.Logger, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
		return nil, err
	}
	return logger, nil
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

// ------------------------ Repos & Services ------------------------

func NewUserRepository(store *postgres.DBStorage) *postgres.UserRepository {
	return postgres.NewUserRepository(store.DB)
}

func NewOrderRepository(store *postgres.DBStorage) *postgres.OrderRepository {
	return postgres.NewOrderRepository(store.DB)
}

func NewWithdrawalRepository(store *postgres.DBStorage) *postgres.WithdrawalRepository {
	return postgres.NewWithdrawalRepository(store.DB)
}

func NewAuthService(repo *postgres.UserRepository, cfg *config.Config, logger *zap.Logger) *service.AuthService {
	return service.NewAuthService(repo, cfg.AuthSecret, logger)
}

func NewAuthHandler(authService *service.AuthService, logger *zap.Logger) *handler.AuthHandler {
	return handler.NewAuthHandler(authService, logger)
}

func NewOrdersService(logger *zap.Logger, orderRepo *postgres.OrderRepository) *service.OrdersService {
	return service.NewOrdersService(logger, orderRepo)
}

func NewOrdersHandler(ordersService *service.OrdersService, logger *zap.Logger) *handler.OrdersHandler {
	return handler.NewOrdersHandler(ordersService, logger)
}

func NewBalanceService(repo *postgres.WithdrawalRepository, logger *zap.Logger) *service.BalanceService {
	return service.NewBalanceService(repo, logger)
}

func NewBalanceHandler(s *service.BalanceService, logger *zap.Logger) *handler.BalanceHandler {
	return handler.NewBalanceHandler(s, logger)
}

// ------------------------ Accrual ------------------------

func NewAccrualClient(cfg *config.Config) *accrual.Client {
	return service.NewAccrualClient(cfg)
}

func NewAccrualWorker(orderRepo *postgres.OrderRepository, client *accrual.Client, logger *zap.Logger) *service.AccrualWorker {
	return service.NewAccrualWorker(orderRepo, client, logger)
}

// ------------------------ Router ------------------------

func newRouter(cfg *config.Config, logger *zap.Logger, authHandler *handler.AuthHandler, ordersHandler *handler.OrdersHandler, balanceHandler *handler.BalanceHandler) chi.Router {
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
		r.Post("/api/user/orders", ordersHandler.UploadOrder)
		r.Get("/api/user/orders", ordersHandler.ListOrders)
		r.Get("/api/user/balance", balanceHandler.GetBalance)
		r.Post("/api/user/balance/withdraw", balanceHandler.Withdraw)
		r.Get("/api/user/withdrawals", balanceHandler.ListWithdrawals)
	})

	return r
}

// ------------------------ Server & Migrations ------------------------

func startServer(lc fx.Lifecycle, cfg *config.Config, router chi.Router, logger *zap.Logger) {
	srv := &http.Server{
		Addr:    cfg.RunAddress,
		Handler: router,
	}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("starting HTTP server", zap.String("addr", cfg.RunAddress))
			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Error("HTTP server failed", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("shutting down HTTP server")
			return srv.Shutdown(ctx)
		},
	})
}

func startMigrations(cfg *config.Config, logger *zap.Logger) error {
	if err := postgres.RunMigrations(cfg.DatabaseURI, logger); err != nil {
		return fmt.Errorf("migrations failed: %w", err)
	}
	return nil
}

// ------------------------ Accrual Worker Lifecycle ------------------------

func StartAccrualWorker(lc fx.Lifecycle, worker *service.AccrualWorker, logger *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("starting accrual worker")
			go func() {
				for {
					worker.Process(ctx)
					time.Sleep(5 * time.Second)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("stopping accrual worker")
			return nil
		},
	})
}
