package userrepository

import (
	"errors"

	"github.com/besean163/gophermart/internal/entities"
	"github.com/besean163/gophermart/internal/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Repository struct {
	DB *gorm.DB
}

func New(db *gorm.DB) (Repository, error) {
	if db == nil {
		return Repository{}, errors.New("empty db connect")
	}

	err := migration(db)
	if err != nil {
		logger.Get().Warn("migration error", zap.String("error", err.Error()))
		return Repository{}, err
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

func migration(db *gorm.DB) error {
	// дропаем таблицы для чистоты каждого запуска
	dropTables(db)

	if !db.Migrator().HasTable(&entities.User{}) {
		err := db.Migrator().CreateTable(&entities.User{})
		if err != nil {
			return err
		}
	}

	return nil
}

func dropTables(db *gorm.DB) error {
	if db.Migrator().HasTable(&entities.User{}) {
		err := db.Migrator().DropTable(&entities.User{})
		if err != nil {
			return err
		}
	}
	return nil
}
