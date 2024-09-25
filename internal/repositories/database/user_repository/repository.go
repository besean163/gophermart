package userrepository

import (
	"errors"

	"github.com/besean163/gophermart/internal/entities"
	"gorm.io/gorm"
)

var (
	ErrEmptyBDConnection = errors.New("empty db connect")
)

type Repository struct {
	DB *gorm.DB
}

func New(db *gorm.DB) (Repository, error) {
	if db == nil {
		return Repository{}, ErrEmptyBDConnection
	}

	return Repository{
		DB: db,
	}, nil
}

func (repository Repository) SaveUser(user entities.User) error {
	repository.DB.Save(&user)
	return nil
}

func (repository Repository) GetUser(login string) *entities.User {
	var user entities.User
	repository.DB.Take(&user, "login = ?", login)
	if user.ID == 0 {
		return nil
	}
	return &user
}
