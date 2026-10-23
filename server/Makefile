# Global settings
SHELL       := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

.DEFAULT_GOAL := help

# Project variables
APP_NAME  := api
BIN_DIR   := bin
BIN       := $(BIN_DIR)/$(APP_NAME)
CMD_PATH  := ./cmd/api

# Runtime environments
DEV_ENV   ?= development
PROD_ENV  ?= production

# Docker variables
COMPOSE_FILE := docker-compose.dev.yml

# Tool variables
GOLANGCI_LINT_VERSION ?= v2.14.0
GOLANGCI_LINT         := golangci-lint
AIR                   := air
GOFMT                 := gofmt
GO                    := go

# Phony targets
.PHONY: help setup clean \
        build run run-build dev \
        test \
        fmt fmt-check \
        lint lint-check \
        up down \
        install-lint install-air

# Help
help: ## Show available targets
	@printf "Usage:\n  make <target>\n\nTargets:\n"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

# Setup & tooling
setup: install-lint install-air ## Install all development tooling

install-lint: ## Install golangci-lint (version set by GOLANGCI_LINT_VERSION)
	$(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

install-air: ## Install air for live reloading
	$(GO) install github.com/air-verse/air@latest

# Build & run
build: ## Compile the binary into ./bin
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN) $(CMD_PATH)

run: ## Run the app in development mode
	APP_ENV=$(DEV_ENV) $(GO) run $(CMD_PATH)

run-build: build ## Build, then run the compiled binary in production mode
	APP_ENV=$(PROD_ENV) $(BIN)

dev: ## Start the app with live reload (air)
	$(AIR)

clean: ## Remove build artifacts
	rm -rf $(BIN_DIR)

# Test
test: ## Run all tests with race detector and coverage
	$(GO) test ./... -race -cover

# Formatting
fmt: ## Format Go source files in-place
	$(GOFMT) -w .

fmt-check: ## Fail if any Go files are not formatted
	@unformatted=$$($(GOFMT) -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "Unformatted files:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

# Linting
lint: ## Run golangci-lint with automatic fixes
	$(GOLANGCI_LINT) run --fix

lint-check: ## Run golangci-lint without applying fixes
	$(GOLANGCI_LINT) run

# Docker
up: ## Start docker-compose development stack (detached)
	docker compose -f $(COMPOSE_FILE) up -d

down: ## Stop docker-compose development stack
	docker compose -f $(COMPOSE_FILE) down