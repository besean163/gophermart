package entities

import "time"

type Order struct {
	Number    string    `json:"number" gorm:"primarykey;autoIncrement:false"`
	UserID    int       `json:"-" gorm:"index"`
	Status    string    `json:"status"`
	Accrual   float64   `json:"accrual,omitempty"`
	UpdatedAt time.Time `json:"uploaded_at"`
}

func NewOrder(number string, userID int) *Order {
	return &Order{
		Number:    number,
		UserID:    userID,
		Status:    OrderStatusNew,
		UpdatedAt: time.Now(),
	}
}
