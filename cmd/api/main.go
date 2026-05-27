package main

import (
	"github.com/mykytakuzminov/task-manager-api/internal/config"
)

func main() {
	cfg := config.Load()
	_ = cfg
}
