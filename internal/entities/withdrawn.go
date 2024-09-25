package entities

import "time"

type Withdrawn struct {
	ID          int       `json:"-" gorm:"primarykey"`
	UserID      int       `json:"-" gorm:"index"`
	OrderNumber string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProccesedAt time.Time `json:"processed_at"`
}

func NewWithdrawn(userID int, orderID string, sum float64) *Withdrawn {
	return &Withdrawn{
		UserID:      userID,
		OrderNumber: orderID,
		Sum:         sum,
	}
}
