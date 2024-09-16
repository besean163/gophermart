package entities

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"time"
)

type User struct {
	ID       int `gorm:"primarykey"`
	Login    string
	Password string
}

func (user User) Validate() error {
	if user.Password == "" {
		return errors.New("empty password")
	}
	return nil
}

func (user *User) HashingPassword() error {

	h := md5.New()
	_, err := h.Write([]byte(user.Password))
	if err != nil {
		return err
	}

	user.Password = hex.EncodeToString(h.Sum(nil))
	return nil
}

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

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

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

type Accural struct {
	Order   string
	Status  string
	Accural float64
}

type AccrualOrder struct {
	Order string
	Goods []Good
}

type Good struct {
	Description string
	Price       float64
}

type AccuralRule struct {
	Match      string
	Reward     float64
	RewardType string
}
