# Makefile for Internal Transfers System

# Variables
DOCKER_COMPOSE := docker compose
GO := go

# Default target
.PHONY: help
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "%-20s %s\n", $$1, $$2}'

# Main commands
.PHONY: test
test: ## Run tests with cache cleaning and coverage
	$(GO) clean -testcache
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out
	$(GO) tool cover -html=coverage.out -o coverage.html

.PHONY: docker-up
docker-up: ## Start the application via Docker
	@if [ ! -f .env ]; then \
		echo ".env file not found. Creating from template..."; \
		cp env.example .env; \
		echo "Created .env file from template"; \
	fi
	$(DOCKER_COMPOSE) up -d
	@echo "Application started successfully"
	@echo "API available at: http://localhost:8080"
	@echo "Database available at: localhost:5432"

# Utility commands
.PHONY: docker-down
docker-down: ## Stop the Docker containers
	$(DOCKER_COMPOSE) down

.PHONY: docker-logs
docker-logs: ## Show Docker container logs
	$(DOCKER_COMPOSE) logs -f

.PHONY: docker-status
docker-status: ## Show Docker container status
	$(DOCKER_COMPOSE) ps

.PHONY: clean
clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html
	$(GO) clean -cache -testcache 