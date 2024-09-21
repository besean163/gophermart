package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func NewDB(dsn string) (*gorm.DB, error) {
	if db == nil {
		conn, err := gorm.Open(postgres.Open(dsn))
		if err != nil {
			return nil, err
		}
		parentDB, err := conn.DB()
		if err != nil {
			return nil, err
		}
		parentDB.SetMaxIdleConns(10)
		db = conn
	}

	return db, nil
}
