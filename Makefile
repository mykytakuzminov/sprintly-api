include .env
export

# ── Docker ────────────────────────────────────────────────
env-up:
	docker-compose up -d

env-down:
	docker-compose down

env-rebuild:
	docker-compose up --build -d

env-reset:
	docker-compose down -v

env-logs:
	docker-compose logs -f

# ── App ───────────────────────────────────────────────────
run:
	go run ./cmd/api/...

build:
	go build -o bin/api ./cmd/api/...

# ── Tests ─────────────────────────────────────────────────
test:
	go test ./... -race -count=1

test-verbose:
	go test ./... -race -count=1 -v

test-cover:
	go test ./... -race -count=1 -coverprofile=coverage.out
	go tool cover -html=coverage.out

# ── Code quality ──────────────────────────────────────────
check:
	go mod tidy
	gofmt -w .
	go vet ./...
	golangci-lint run
	go build ./...
