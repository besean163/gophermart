package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/besean163/gophermart/internal/database"
	"github.com/besean163/gophermart/internal/handlers"
	"github.com/besean163/gophermart/internal/logger"
	"github.com/besean163/gophermart/internal/migration"
	databaseorders "github.com/besean163/gophermart/internal/repositories/database/order_repository"
	databaseusers "github.com/besean163/gophermart/internal/repositories/database/user_repository"
	inmemorders "github.com/besean163/gophermart/internal/repositories/inmem/order_repository"
	inmemusers "github.com/besean163/gophermart/internal/repositories/inmem/user_repository"
	authservice "github.com/besean163/gophermart/internal/services/auth_service"
	loyalityservice "github.com/besean163/gophermart/internal/services/loyality_service"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type Server interface {
	Start() error
	Shutdown(ctx context.Context) error
}

type App struct {
	ctx    context.Context
	config AppConfig
}

func NewApp() App {
	config := NewConfig()
	ctx := context.Background()

	return App{
		ctx:    ctx,
		config: config,
	}
}

func (app App) Run() error {
	err := logger.NewLogger()
	if err != nil {
		return err
	}
	err = migration.Run(app.config.DatabaseDSN)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(app.ctx)
	defer cancel()
	runGracefulStopRoutine(cancel)

	handler, err := NewHandler(ctx, app.config)
	if err != nil {
		return err
	}
	server := NewServer(app.config, handler)

	errGroup, _ := errgroup.WithContext(ctx)
	errGroup.Go(func() error {
		logger.Get().Info("run app", zap.String("address", app.config.RunAddress), zap.String("accrual address", app.config.RunAccrualAddress))
		return server.Start()
	})

	errGroup.Go(func() error {
		<-ctx.Done()
		return server.Shutdown(context.Background())
	})

	if err := errGroup.Wait(); err != nil {
		return err
	}

	return nil
}

func runGracefulStopRoutine(cancel context.CancelFunc) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	go func() {
		<-sig
		cancel()
	}()
}

func NewHandler(ctx context.Context, config AppConfig) (handlers.Handler, error) {
	var handler handlers.Handler
	authService, err := NewAuthService(config)
	if err != nil {
		return handler, err
	}
	loyalityService, err := NewLoyaltyService(ctx, config)
	if err != nil {
		return handler, err
	}

	handler = handlers.NewHandlers(authService, loyalityService, config.HashSecret)
	return handler, nil
}

type CustomServer struct {
	handler handlers.Handler
	config  AppConfig
	server  *http.Server
}

func NewServer(config AppConfig, handler handlers.Handler) Server {
	server := http.Server{
		Addr:    config.RunAddress,
		Handler: handler,
	}
	return &CustomServer{
		handler: handler,
		config:  config,
		server:  &server,
	}
}

func (server *CustomServer) Start() error {
	return server.server.ListenAndServe()
}

func (server *CustomServer) Shutdown(ctx context.Context) error {
	return server.server.Shutdown(ctx)
}

func NewLoyaltyService(ctx context.Context, config AppConfig) (handlers.LoyaltyService, error) {

	var repository loyalityservice.OrderRepository
	if config.DatabaseDSN == "" {
		repository = inmemorders.New()
	} else {
		db, err := database.NewDB(config.DatabaseDSN)
		if err != nil {
			return nil, err
		}
		repository, err = databaseorders.NewRepository(db)
		if err != nil {
			return nil, err
		}
	}

	return loyalityservice.New(ctx, repository, config.RunAccrualAddress), nil
}

func NewAuthService(config AppConfig) (handlers.AuthService, error) {
	var repository authservice.UserRepository
	if config.DatabaseDSN == "" {
		repository = inmemusers.New()
	} else {
		db, err := database.NewDB(config.DatabaseDSN)
		if err != nil {
			return nil, err
		}
		repository, err = databaseusers.New(db)
		if err != nil {
			return nil, err
		}
	}

	return authservice.New(repository, config.HashSecret, time.Hour*3), nil
}
