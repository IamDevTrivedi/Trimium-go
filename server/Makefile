.SHELLFLAGS   := -eu -o pipefail -c
.DEFAULT_GOAL := help

DEV_ENV  ?= development
PROD_ENV ?= production

GEOIP_DB ?= ./data/geoip/GeoLite2-City.mmdb

GOLANGCI_LINT_VERSION ?= v2.14.0
GOVULNCHECK_VERSION   ?= v1.9.0
AIR_VERSION           ?= v1.67.4

COMPOSE_FILE := docker-compose.dev.yml

.PHONY: help tools build dev run-worker test fmt lint lint-check govulncheck \
	    tidy tidy-check ci check up down geoip-update geoip-check

help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  help          Show available targets"
	@echo "  tools         Install pinned dev tooling (golangci-lint, govulncheck, air)"
	@echo "  build         Compile all packages"
	@echo "  dev           Run the API with live reload (air)"
	@echo "  run-worker    Run the worker in development mode"
	@echo "  test          Run tests with race detector and coverage"
	@echo "  fmt           Format Go source files"
	@echo "  lint          Run golangci-lint with automatic fixes"
	@echo "  lint-check    Run golangci-lint without fixes"
	@echo "  govulncheck   Scan for known vulnerabilities"
	@echo "  tidy          Tidy go.mod and go.sum"
	@echo "  tidy-check    Fail if go.mod or go.sum need tidying"
	@echo "  ci            Run all CI checks (stops at first failure)"
	@echo "  check         Alias for ci"
	@echo "  up            Start the docker compose development stack"
	@echo "  down          Stop the docker compose development stack"
	@echo "  geoip-update  Download GeoLite2 database (needs MAXMIND_LICENSE_KEY)"
	@echo "  geoip-check   Fail if the GeoIP database is missing"
	@echo ""

tools:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)
	go install github.com/air-verse/air@$(AIR_VERSION)

build:
	go build ./...

dev:
	air

run-worker:
	APP_ENV=$(DEV_ENV) go run ./cmd/worker

test:
	go test ./... -race -cover

fmt:
	gofmt -w .

lint:
	golangci-lint run --fix

lint-check:
	golangci-lint run
	golangci-lint fmt --diff

govulncheck:
	govulncheck ./...

tidy:
	go mod tidy

tidy-check:
	go mod tidy -diff

ci: tidy-check lint-check govulncheck test build
	@echo "All CI checks passed successfully!"

check: ci

up:
	docker compose -f $(COMPOSE_FILE) up -d

down:
	docker compose -f $(COMPOSE_FILE) down

geoip-update:
	@test -n "$${MAXMIND_LICENSE_KEY:-}" || { echo "MAXMIND_LICENSE_KEY is not set"; exit 1; }
	GEOIP_DB_PATH=$(GEOIP_DB) go run ./cmd/geoipupdate

geoip-check:
	@test -f $(GEOIP_DB) || { echo "Missing $(GEOIP_DB). Run: make geoip-update"; exit 1; }