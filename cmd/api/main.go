package main

import (
	"github.com/mykytakuzminov/sprintly-api/internal/app"
)

// @title           Sprintly API
// @version         1.0
// @description     REST API for managing boards, columns and tasks with JWT authentication.
// @description     Supports token-based auth, rate limiting, pagination, sorting and filtering.
//
// @contact.name    Mykyta Kuzminov
// @contact.url     https://github.com/mykytakuzminov
//
// @license.name    MIT
// @license.url     https://opensource.org/licenses/MIT
//
// @host            204.168.173.88:8080
// @BasePath        /api/v1
//
// @securityDefinitions.apikey BearerAuth
// @in                         header
// @name                       Authorization
// @description                Enter your Bearer token in the format: Bearer <token>
func main() {
	app := app.New()
	app.Run()
}
