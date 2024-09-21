package main

import (
	"log"
)

func main() {
	app := NewApp()
	if err := app.run(); err != nil {
		log.Fatal(err)
	}
}
