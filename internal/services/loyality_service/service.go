package loyalityservice

import (
	"context"
	"time"

	"github.com/besean163/gophermart/internal/entities"
	"github.com/besean163/gophermart/internal/logger"
	"go.uber.org/zap"
)

const (
	tickSec            = 1
	accrualWorkerCount = 10
)

type Service struct {
	accrualServiceURL string
	repository        OrderRepository
}

type OrderRepository interface {
	GetOrder(orderID string) *entities.Order
	GetUserOrders(userID int) []*entities.Order
	GetUserWithdrawals(userID int) []*entities.Withdrawn
	SaveOrder(entities.Order) error
	SaveWithdrawn(entities.Withdrawn) error
	GetWaitProcessOrders() []*entities.Order
}

func New(ctx context.Context, repository OrderRepository, accrualServiceURL string) Service {

	service := Service{
		accrualServiceURL: accrualServiceURL,
		repository:        repository,
	}
	service.runAccrualJobService(ctx)

	return service
}

func (service Service) GetOrder(orderNumber string) *entities.Order {
	return service.repository.GetOrder(orderNumber)
}

func (service Service) SaveOrder(order entities.Order) error {
	return service.repository.SaveOrder(order)
}

func (service Service) GetUserOrders(userID int) []*entities.Order {
	return service.repository.GetUserOrders(userID)
}

func (service Service) GetUserWithdrawals(userID int) []*entities.Withdrawn {
	return service.repository.GetUserWithdrawals(userID)
}

func (service Service) GetUserBalance(userID int) entities.Balance {
	orders := service.GetUserOrders(userID)
	withdrawals := service.GetUserWithdrawals(userID)

	totalSum := 0.0
	for _, order := range orders {
		totalSum += order.Accrual
	}

	totalWithdrawn := 0.0
	for _, withdrawn := range withdrawals {
		totalWithdrawn += withdrawn.Sum
	}

	return entities.Balance{
		Current:   totalSum - totalWithdrawn,
		Withdrawn: totalWithdrawn,
	}
}

func (service Service) SaveWithdrawn(withdrawn entities.Withdrawn) error {
	return service.repository.SaveWithdrawn(withdrawn)
}

type AccrualOrder struct {
	Order   string
	Status  string
	Accrual float64
}

func (service Service) runAccrualJobService(ctx context.Context) {
	orderIn := make(chan entities.Order, 1)
	savingOrders := make(chan entities.Order, 1)
	errorChan := make(chan error)

	for workerID := 1; workerID <= accrualWorkerCount; workerID++ {
		go worker(ctx, workerID, orderIn, savingOrders, service.accrualServiceURL, errorChan)
	}

	go service.saver(ctx, savingOrders)
	go log(ctx, errorChan)

	go func() {
		ticker := time.NewTicker(time.Second * tickSec)
		for {
			select {
			case <-ticker.C:
				orders := service.repository.GetWaitProcessOrders()
				for _, order := range orders {
					orderIn <- *order
				}
			case <-ctx.Done():
				logger.Get().Info("close update")
				return
			}
		}
	}()
}

func (service Service) saver(ctx context.Context, savingOrders chan entities.Order) {
	for {
		select {
		case order := <-savingOrders:
			err := service.repository.SaveOrder(order)
			if err != nil {
				logger.Get().Warn("save order error", zap.String("error", err.Error()))
			}
		case <-ctx.Done():
			return
		}
	}
}

func log(ctx context.Context, errorChan chan error) {
	for {
		select {
		case err := <-errorChan:
			logger.Get().Warn("Service error.", zap.String("error", err.Error()))
		case <-ctx.Done():
			return
		}
	}
}
