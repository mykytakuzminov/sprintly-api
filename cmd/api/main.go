package main

import (
	"log"

	"github.com/mykytakuzminov/task-manager-api/internal/app"
)

func main() {
	app := app.New()
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
	log.Println("server stopped")
}
