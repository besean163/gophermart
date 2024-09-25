package main

import (
	"log"

	"github.com/besean163/gophermart/internal/logger"
	"github.com/besean163/gophermart/internal/migration"
)

func main() {
	logger.NewLogger()
	config := migration.NewConfig()
	if err := migration.Run(config.DatabaseDSN); err != nil {
		log.Fatal(err)
	}
}
