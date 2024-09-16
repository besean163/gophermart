package orderrepository

import (
	"slices"

	"github.com/besean163/gophermart/internal/entities"
)

type Repository struct {
	orders      []*entities.Order
	withdrawals []*entities.Withdrawn
}

func New() *Repository {
	return &Repository{
		orders:      make([]*entities.Order, 0),
		withdrawals: make([]*entities.Withdrawn, 0),
	}
}

func (repository *Repository) GetOrder(orderID string) *entities.Order {
	for _, order := range repository.orders {
		if order.Number == orderID {
			return order
		}
	}
	return nil
}

func (repository *Repository) GetUserOrders(userID int) []*entities.Order {
	var orders []*entities.Order
	for _, order := range repository.orders {
		if order.UserID == userID {
			orders = append(orders, order)
		}
	}
	return orders
}

func (repository *Repository) GetUserWithdrawals(userID int) []*entities.Withdrawn {
	var withdrawals []*entities.Withdrawn
	for _, withdrawn := range repository.withdrawals {
		if withdrawn.UserID == userID {
			withdrawals = append(withdrawals, withdrawn)
		}
	}
	return withdrawals
}

func (repository *Repository) SaveOrder(inOrder entities.Order) error {
	var exist *entities.Order
	for _, order := range repository.orders {
		if order.Number == inOrder.Number {
			exist = order
			break
		}
	}

	if exist == nil {
		repository.orders = append(repository.orders, &inOrder)
	} else {
		exist.Accrual = inOrder.Accrual
		exist.Number = inOrder.Number
		exist.Status = inOrder.Status
		exist.UserID = inOrder.UserID
	}

	return nil
}

func (repository *Repository) SaveWithdrawn(inWithdrawn entities.Withdrawn) error {
	var exist *entities.Withdrawn
	for _, withdrawn := range repository.withdrawals {
		if withdrawn.OrderNumber == inWithdrawn.OrderNumber {
			exist = withdrawn
			break
		}
	}

	if exist == nil {
		repository.withdrawals = append(repository.withdrawals, &inWithdrawn)
	} else {
		exist.ID = inWithdrawn.ID
		exist.UserID = inWithdrawn.UserID
		exist.OrderNumber = inWithdrawn.OrderNumber
		exist.Sum = inWithdrawn.Sum
		exist.ProccesedAt = inWithdrawn.ProccesedAt
	}

	return nil
}

func (repository Repository) GetWaitProcessOrders() []*entities.Order {
	var orders []*entities.Order
	for _, order := range repository.orders {
		if slices.Contains([]string{
			entities.OrderStatusNew,
			entities.OrderStatusProcessing,
		}, order.Status) {
			orders = append(orders, order)
		}
	}
	return orders
}
