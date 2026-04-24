# =========================
# Variables
# =========================

COMPOSE_FILE = docker-compose.yml
ENV_FILE = ./deployments/envs/prod.env
GOLANGCI_LINT = $(shell go env GOPATH)/bin/golangci-lint
LINT_PACKAGES = ./cmd/app ./cmd/migrator ./docs ./internal/app ./internal/closer ./internal/config ./internal/domain ./internal/http/errors ./internal/http/handler ./internal/http/handler/mocks ./internal/http/middleware ./internal/migrator ./internal/repository/postgres ./internal/service ./internal/service/mocks ./utils/logg

# =========================
# Phony targets
# =========================

.PHONY: up build-up down swagger test test-integration lint
        
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

lint:
	XDG_CACHE_HOME=/tmp GOLANGCI_LINT_CACHE=/tmp/golangci-lint $(GOLANGCI_LINT) run $(LINT_PACKAGES)
