# =========================
# Variables
# =========================

COMPOSE_FILE = docker-compose.yml
ENV_FILE = ./deployments/envs/prod.env

# =========================
# Phony targets
# =========================

.PHONY: up build-up down swagger test test-integration
        
# =========================
# Docker: production
# =========================

up:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) up -d

build-up:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) down -v
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) up -d --build

down:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) down -v

swagger:
	go run github.com/swaggo/swag/cmd/swag@v1.16.2 init -g ./cmd/app/main.go -o ./docs --parseInternal

test:
	go test ./...

test-integration:
	go test -tags=integration ./...
