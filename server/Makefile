# Global settings
SHELL       := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

.DEFAULT_GOAL := help

# Project variables
APP_NAME  := api
BIN_DIR   := bin
BIN       := $(BIN_DIR)/$(APP_NAME)
CMD_PATH  := ./cmd/api
GEOIP_CMD := ./cmd/geoipupdate
GEOIP_DB  ?= ./data/geoip/GeoLite2-City.mmdb

# Runtime environments
DEV_ENV   ?= development
PROD_ENV  ?= production

# Docker variables
COMPOSE_FILE := docker-compose.dev.yml

# Tool variables
GOLANGCI_LINT_VERSION ?= v2.14.0
GOVULNCHECK_VERSION   ?= v1.9.0
GOLANGCI_LINT         := golangci-lint
GOVULNCHECK           := govulncheck
AIR                   := air
GOFMT                 := gofmt
GO                    := go

# Phony targets
.PHONY: help setup clean \
        build build-check run run-build dev \
        test vet govulncheck \
        fmt fmt-check \
        lint lint-check \
        tidy tidy-check \
        ci check \
        up down \
        install-lint install-air install-govulncheck \
		geoip-update geoip-check

# Help
help: ## Show available targets
	@printf "Usage:\n  make <target>\n\nTargets:\n"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

# Setup & tooling
setup: install-lint install-air install-govulncheck ## Install all development tooling

install-lint: ## Install golangci-lint (version set by GOLANGCI_LINT_VERSION)
	$(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

install-air: ## Install air for live reloading
	$(GO) install github.com/air-verse/air@latest

install-govulncheck: ## Install govulncheck (version set by GOVULNCHECK_VERSION)
	$(GO) install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)

# Build & run
build: ## Compile the binary into ./bin
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN) $(CMD_PATH)

build-check: ## Verify that all packages and commands compile
	$(GO) build ./...

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

vet: ## Run go vet on all packages
	$(GO) vet ./...

govulncheck: ## Scan code, dependencies and toolchain for known vulnerabilities
	$(GOVULNCHECK) ./...

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
	$(GOLANGCI_LINT) fmt --diff

# Modules
tidy: ## Download and clean up go.mod and go.sum
	$(GO) mod tidy

tidy-check: ## Verify module files and fail if go.mod/go.sum need tidying
	$(GO) mod download
	$(GO) mod verify
	$(GO) mod tidy -diff

# All CI checks
ci: ## Run all CI checks silently and report final result (mod, fmt, lint, vet, vuln, race test, build)
	@tmp=$$(mktemp); \
	trap 'rm -f "$$tmp"' EXIT; \
	run_step() { \
		if ! "$$@" > "$$tmp" 2>&1; then \
			echo "[FAILED] $$*"; \
			cat "$$tmp"; \
			exit 1; \
		fi; \
	}; \
	run_step make tidy-check; \
	run_step make fmt-check; \
	run_step make lint-check; \
	run_step make vet; \
	run_step make govulncheck; \
	run_step make test; \
	run_step make build-check; \
	echo "All CI checks passed successfully!"

check: ci ## Alias for make ci

# Docker
up: ## Start docker-compose development stack (detached)
	docker compose -f $(COMPOSE_FILE) up -d

down: ## Stop docker-compose development stack
	docker compose -f $(COMPOSE_FILE) down

# GeoIP
geoip-update: ## Download/refresh the GeoLite2-City database (needs MAXMIND_LICENSE_KEY)
	@test -n "$${MAXMIND_LICENSE_KEY:-}" || { echo "MAXMIND_LICENSE_KEY is not set"; exit 1; }
	GEOIP_DB_PATH=$(GEOIP_DB) $(GO) run $(GEOIP_CMD)

geoip-check: ## Fail if the GeoIP database is missing
	@test -f $(GEOIP_DB) || { echo "Missing $(GEOIP_DB). Run: make geoip-update"; exit 1; }