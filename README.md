<div align="center">

# 🚀 Sprintly

Task management REST API similar to Kanban

[![CI](https://github.com/mykytakuzminov/sprintly-api/actions/workflows/ci.yml/badge.svg)](https://github.com/mykytakuzminov/sprintly-api/actions/workflows/ci.yml)
[![CD](https://github.com/mykytakuzminov/sprintly-api/actions/workflows/cd.yml/badge.svg)](https://github.com/mykytakuzminov/sprintly-api/actions/workflows/cd.yml)
[![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)

[API Docs](http://204.168.173.88:8080/swagger/index.html) · [Health Check](http://204.168.173.88:8080/api/v1/health)

</div>

---

## Features

- **Admin** - get all, delete and change users roles
- **Auth** - register, authorize and logout user
- **Users** - get current user or change its password
- **Boards** - create, get, update and delete boards
- **Columns** - create, get, delete and update columns
- **Tasks** - create, get, delete and update tasks
- **Health** - check application and databases health

## Tech Highlights

- **Clean Architecture** - four layers built on interfaces with dependency injection
- **JWT Authentication** - access and refresh token rotation
- **Role-Based Access Control** - admin and member roles to separate features access
- **Rate Limiting** - Token Bucket algorithm to prevent abuse
- **Pagination** - to prevent big data responses and server load
- **Timeouts** - to prevent frozen or long requests
- **Graceful Shutdown** - to prevent data loss during unexpected shutdowns
- **Logging** - to check app behaviour and errors
- **Migrations** - run automatically on app startup
- **CI/CD** - automated linting, testing and deploy

## Architecture

<img alt="Architecutre" src="https://github.com/user-attachments/assets/3265cbe7-b555-4107-8ac8-b2108d4a6a7d" />


## Getting Started

### Prerequisites

- `docker` and `docker compose`

### Installation

```bash
git clone https://github.com/mykytakuzminov/sprintly-api.git
cd sprintly-api
cp .env.example .env
docker compose up -d
```

- **API:** `http://localhost:8080/api/v1`
- **API Docs:** `http://localhost:8080/swagger/index.html`

### Development

```bash
make check   # run linter
make test    # run unit and integration tests
```

## Tech Stack

### Core

- `Go`, `PostgreSQL`, `Redis`, `Docker`, `Docker Compose`, `GitHub Actions`

### Libraries

- `chi`, `zap`, `testcontainers`, `swag`, `golang-migrate`, `jwt`, `validator`
