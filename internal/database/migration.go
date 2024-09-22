package database

import (
	"gorm.io/gorm"
)

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
	}

	return nil
}
