# =========================
# Variables
# =========================

COMPOSE_FILE = docker-compose.yml
ENV_FILE = ./deployments/envs/dev.env #TODO: Поменять на prod.env

# =========================
# Phony targets
# =========================

.PHONY: up build-up down swagger
        
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
