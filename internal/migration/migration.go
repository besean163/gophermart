package migration

import (
	"errors"
	"reflect"

	"github.com/besean163/gophermart/internal/database"
	"github.com/besean163/gophermart/internal/entities"
	"github.com/besean163/gophermart/internal/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ErrEmptyDBConnectRow = errors.New("empty dsn row")
)

func Run(dsn string) error {
	if dsn == "" {
		return ErrEmptyDBConnectRow
	}

	db, err := database.NewDB(dsn)
	if err != nil {
		return err
	}

	e := []interface{}{
		entities.User{},
		entities.Order{},
		entities.Withdrawn{},
	}

	err = Migration(db, e...)

	if err != nil {
		logger.Get().Warn("migration error", zap.String("error", err.Error()))
		return err
	}
	return nil
}

func Migration(db *gorm.DB, items ...interface{}) error {
	for _, item := range items {
		// дропаем таблицы для чистоты каждого запуска
		if db.Migrator().HasTable(&item) {
			err := db.Migrator().DropTable(&item)
			if err != nil {
				return err
			}
		}

		if !db.Migrator().HasTable(&item) {
			err := db.Migrator().CreateTable(&item)
			if err != nil {
				return err
			}
		}

		name := getEntityName(item)
		logger.Get().Info("Entity migration done", zap.String("name", name))
	}

	return nil
}

func getEntityName(entity interface{}) string {
	if t := reflect.TypeOf(entity); t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	} else {
		return t.Name()
	}
}
