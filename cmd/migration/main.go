package main

import (
	"log"

	"github.com/besean163/gophermart/internal/migration"
)

func main() {
	if err := migration.Run(); err != nil {
		log.Fatal(err)
	}
}
