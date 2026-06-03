package main

import (
	"github.com/mykytakuzminov/task-manager-api/internal/app"
)

// @title           Task Manager API
// @version         1.0
// @description     REST API for managing boards, columns and tasks with JWT authentication.
//
// @contact.name    Mykyta Kuzminov
//
// @host            localhost:8080
// @BasePath        /api/v1
//
// @securityDefinitions.apikey BearerAuth
// @in                         header
// @name                       Authorization
func main() {
	app := app.New()
	app.Run()
}
