# ComputeHive Makefile

.PHONY: help all build test clean docker-build docker-push setup lint fmt

# Variables
GO := go
DOCKER := docker
NPM := npm
GOPATH := $(shell go env GOPATH)
GOBIN := $(GOPATH)/bin
VERSION := $(shell git describe --tags --always --dirty)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -s -w"

# Colors for output
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m # No Color

## help: Display this help message
help:
	@echo "ComputeHive Development Commands:"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "Usage: make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

## all: Build all components
all: build-agent build-services build-frontend

## setup: Install development dependencies
setup:
	@echo "$(GREEN)Installing development dependencies...$(NC)"
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@$(GO) install github.com/securego/gosec/v2/cmd/gosec@latest
	@$(GO) install golang.org/x/tools/cmd/goimports@latest
	@cd web/dashboard && $(NPM) install
	@cd contracts && $(NPM) install
	@echo "$(GREEN)Setup complete!$(NC)"

## build-agent: Build the compute agent
build-agent:
	@echo "$(GREEN)Building agent...$(NC)"
	@cd agent && $(GO) build $(LDFLAGS) -o ../bin/computehive-agent ./cmd/agent

## build-services: Build all core services
build-services:
	@echo "$(GREEN)Building core services...$(NC)"
	@cd core-services && $(GO) build $(LDFLAGS) -o ../bin/auth-service ./auth-service

## build-frontend: Build the web dashboard
build-frontend:
	@echo "$(GREEN)Building frontend...$(NC)"
	@cd web/dashboard && $(NPM) run build

##@ Testing

## test: Run all tests
test: test-go test-frontend test-contracts

## test-go: Run Go tests
test-go:
	@echo "$(GREEN)Running Go tests...$(NC)"
	@cd agent && $(GO) test -v -race -cover ./...
	@cd core-services && $(GO) test -v -race -cover ./...

## test-frontend: Run frontend tests
test-frontend:
	@echo "$(GREEN)Running frontend tests...$(NC)"
	@cd web/dashboard && $(NPM) test

## test-contracts: Run smart contract tests
test-contracts:
	@echo "$(GREEN)Running contract tests...$(NC)"
	@cd contracts && $(NPM) test

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	@cd agent && $(GO) test -v -race -coverprofile=coverage.out ./...
	@cd agent && $(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: agent/coverage.html$(NC)"

##@ Code Quality

## lint: Run linters
lint: lint-go lint-frontend

## lint-go: Run Go linters
lint-go:
	@echo "$(GREEN)Running Go linters...$(NC)"
	@cd agent && golangci-lint run
	@cd core-services && golangci-lint run

## lint-frontend: Run frontend linters
lint-frontend:
	@echo "$(GREEN)Running frontend linters...$(NC)"
	@cd web/dashboard && $(NPM) run lint

## fmt: Format code
fmt:
	@echo "$(GREEN)Formatting code...$(NC)"
	@cd agent && $(GO) fmt ./...
	@cd core-services && $(GO) fmt ./...
	@cd agent && goimports -w .
	@cd core-services && goimports -w .

## security: Run security scans
security:
	@echo "$(GREEN)Running security scans...$(NC)"
	@cd agent && gosec -severity medium ./...
	@cd core-services && gosec -severity medium ./...
	@cd web/dashboard && $(NPM) audit
	@cd contracts && $(NPM) audit

##@ Docker

## docker-build: Build all Docker images
docker-build: docker-build-agent docker-build-auth docker-build-dashboard

## docker-build-agent: Build agent Docker image
docker-build-agent:
	@echo "$(GREEN)Building agent Docker image...$(NC)"
	@$(DOCKER) build -t computehive/agent:$(VERSION) -t computehive/agent:latest ./agent

## docker-build-auth: Build auth service Docker image
docker-build-auth:
	@echo "$(GREEN)Building auth service Docker image...$(NC)"
	@$(DOCKER) build -t computehive/auth-service:$(VERSION) -t computehive/auth-service:latest -f core-services/auth-service/Dockerfile ./core-services

## docker-build-dashboard: Build dashboard Docker image
docker-build-dashboard:
	@echo "$(GREEN)Building dashboard Docker image...$(NC)"
	@$(DOCKER) build -t computehive/dashboard:$(VERSION) -t computehive/dashboard:latest ./web/dashboard

## docker-push: Push Docker images to registry
docker-push:
	@echo "$(GREEN)Pushing Docker images...$(NC)"
	@$(DOCKER) push computehive/agent:$(VERSION)
	@$(DOCKER) push computehive/agent:latest
	@$(DOCKER) push computehive/auth-service:$(VERSION)
	@$(DOCKER) push computehive/auth-service:latest
	@$(DOCKER) push computehive/dashboard:$(VERSION)
	@$(DOCKER) push computehive/dashboard:latest

##@ Local Development

## run-agent: Run the agent locally
run-agent:
	@echo "$(GREEN)Running agent locally...$(NC)"
	@cd agent && $(GO) run ./cmd/agent --control-plane http://localhost:8000

## run-auth: Run auth service locally
run-auth:
	@echo "$(GREEN)Running auth service locally...$(NC)"
	@cd core-services && $(GO) run ./auth-service/main.go

## run-dashboard: Run dashboard locally
run-dashboard:
	@echo "$(GREEN)Running dashboard locally...$(NC)"
	@cd web/dashboard && $(NPM) start

## dev: Start development environment with docker-compose
dev:
	@echo "$(GREEN)Starting development environment...$(NC)"
	@docker-compose up -d
	@echo "$(GREEN)Development environment started!$(NC)"
	@echo "Dashboard: http://localhost:3000"
	@echo "API: http://localhost:8000"

## dev-down: Stop development environment
dev-down:
	@echo "$(YELLOW)Stopping development environment...$(NC)"
	@docker-compose down

##@ Database

## db-migrate: Run database migrations
db-migrate:
	@echo "$(GREEN)Running database migrations...$(NC)"
	@migrate -path ./migrations -database "postgresql://localhost/computehive?sslmode=disable" up

## db-rollback: Rollback database migration
db-rollback:
	@echo "$(YELLOW)Rolling back database migration...$(NC)"
	@migrate -path ./migrations -database "postgresql://localhost/computehive?sslmode=disable" down 1

##@ Utilities

## clean: Clean build artifacts
clean:
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	@rm -rf bin/
	@rm -rf agent/coverage.*
	@rm -rf core-services/coverage.*
	@cd web/dashboard && rm -rf build/
	@echo "$(GREEN)Clean complete!$(NC)"

## deps: Download and tidy dependencies
deps:
	@echo "$(GREEN)Updating dependencies...$(NC)"
	@cd agent && $(GO) mod download && $(GO) mod tidy
	@cd core-services && $(GO) mod download && $(GO) mod tidy

## version: Display version information
version:
	@echo "ComputeHive Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Go Version: $(shell go version)"
	@echo "Node Version: $(shell node --version)"

## logs: Tail logs from development environment
logs:
	@docker-compose logs -f

## generate: Generate code (mocks, protobuf, etc.)
generate:
	@echo "$(GREEN)Generating code...$(NC)"
	@cd agent && $(GO) generate ./...
	@cd core-services && $(GO) generate ./...

# Default target
.DEFAULT_GOAL := help 