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
