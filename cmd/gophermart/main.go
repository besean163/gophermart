package main

import (
	"log"

	"github.com/besean163/gophermart/internal/app"
)

func main() {
	a := app.NewApp()
	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}
