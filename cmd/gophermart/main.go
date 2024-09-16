package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/besean163/gophermart/internal/handlers"
	"github.com/besean163/gophermart/internal/logger"
	orderrepository "github.com/besean163/gophermart/internal/repositories/database/order_repository"
	userrepository "github.com/besean163/gophermart/internal/repositories/database/user_repository"
	inmemorders "github.com/besean163/gophermart/internal/repositories/inmem/order_repository"
	inmem "github.com/besean163/gophermart/internal/repositories/inmem/user_repository"
	authservice "github.com/besean163/gophermart/internal/services/auth_service"
	loyalityservice "github.com/besean163/gophermart/internal/services/loyality_service"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}

}

func run() error {
	err := logger.Set()
	if err != nil {
		return err
	}
	config := NewConfig()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errGroup, _ := errgroup.WithContext(ctx)
	// errGroup, errGroupCtx := errgroup.WithContext(ctx)
	runGracefulStopRutine(cancel)

	authService, err := getAuthService(config)
	if err != nil {
		return err
	}
	loyalityService, err := getLoyaltyService(ctx, config)
	if err != nil {
		return err
	}

	handler := handlers.New(authService, loyalityService, config.HashSecret)
	handler.Mount()

	s := http.Server{
		Addr:    config.RunAddress,
		Handler: handler,
	}

	errGroup.Go(func() error {
		logger.Get().Info("run server", zap.String("addres", s.Addr), zap.String("accrual addres", config.RunAccrualAddress))
		err := s.ListenAndServe()
		if err != nil {
			fmt.Println(err)
		}
		return err
	})

	errGroup.Go(func() error {
		<-ctx.Done()
		return s.Shutdown(context.Background())
	})

	if err := errGroup.Wait(); err != nil {
		return err
	}

	return nil
}

func runGracefulStopRutine(cancel context.CancelFunc) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	go func() {
		<-sig
		cancel()
	}()
}

func getDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn))
	if err != nil {
		return nil, err
	}
	return db, nil
}

func getLoyaltyService(ctx context.Context, config ServerConfig) (handlers.LoyaltyService, error) {

	var repository loyalityservice.OrderRepository
	if config.DatabaseDSN == "" {
		repository = inmemorders.New()
	} else {
		db, err := getDB(config.DatabaseDSN)
		if err != nil {
			return nil, err
		}
		repository, err = orderrepository.New(db)
		if err != nil {
			return nil, err
		}
	}

	return loyalityservice.New(ctx, repository, config.RunAccrualAddress), nil
}

func getAuthService(config ServerConfig) (handlers.AuthService, error) {
	var repository authservice.UserRepository
	if config.DatabaseDSN == "" {
		repository = inmem.New()
	} else {
		db, err := getDB(config.DatabaseDSN)
		if err != nil {
			return nil, err
		}
		repository, err = userrepository.New(db)
		if err != nil {
			return nil, err
		}
	}

	return authservice.New(repository, config.HashSecret, time.Hour*3), nil
}
