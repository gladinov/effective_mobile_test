# =========================
# Variables
# =========================

COMPOSE_FILE = docker-compose.yml
ENV_FILE = ./deployments/envs/dev.env #TODO: Поменять на prod.env

# =========================
# Phony targets
# =========================

.PHONY: up build-up down
        
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
