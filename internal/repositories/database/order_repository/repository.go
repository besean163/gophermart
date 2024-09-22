package orderrepository

import (
	"errors"

	"github.com/besean163/gophermart/internal/database"
	"github.com/besean163/gophermart/internal/entities"
	"github.com/besean163/gophermart/internal/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ErrEmptyBDConnection = errors.New("empty db connect")
)

type Repository struct {
	DB *gorm.DB
}

func NewRepository(db *gorm.DB) (Repository, error) {
	if db == nil {
		return Repository{}, ErrEmptyBDConnection
	}

	err := database.Migration(db, entities.Order{}, entities.Withdrawn{})

	if err != nil {
		logger.Get().Warn("migration error", zap.String("error", err.Error()))
		return Repository{}, err
	}

	return Repository{
		DB: db,
	}, nil
}

func (repository Repository) GetOrder(id string) *entities.Order {
	var order entities.Order
	repository.DB.Take(&order, "number = ?", id)
	if order.Number == "" {
		return nil
	}
	return &order
}

func (repository Repository) SaveOrder(order entities.Order) error {
	repository.DB.Save(&order)
	return nil
}

func (repository Repository) GetUserOrders(userID int) []*entities.Order {
	var orders []*entities.Order
	repository.DB.Find(&orders, "user_id = ?", userID)
	return orders
}

func (repository Repository) GetUserWithdrawals(userID int) []*entities.Withdrawn {
	var withdrawals []*entities.Withdrawn
	repository.DB.Find(&withdrawals, "user_id = ?", userID)
	return withdrawals
}

func (repository Repository) SaveWithdrawn(withdrawn entities.Withdrawn) error {
	repository.DB.Save(&withdrawn)
	return nil
}

func (repository Repository) GetWaitProcessOrders() []*entities.Order {
	var orders []*entities.Order
	repository.DB.Find(&orders, "status IN ?", []string{entities.OrderStatusNew, entities.OrderStatusProcessing})
	return orders
}
