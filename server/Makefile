GOLANGCI_LINT_VERSION ?= v2.14.0

.PHONY: build test run run-build dev up down fmt fmt-check lint lint-check install-lint setup

build:
	go build -o bin/api ./cmd/api

test:
	go test ./... -race -cover

run:
	APP_ENV=development go run ./cmd/api

run-build:
	APP_ENV=production ./bin/api

dev:
	air

up:
	docker compose -f docker-compose.dev.yml up -d

down:
	docker compose -f docker-compose.dev.yml down

fmt-check:
	@out=$$(gofmt -l .); \
	if [ -n "$$out" ]; then \
		echo "Unformatted files:"; echo "$$out"; exit 1; \
	fi

fmt:
	gofmt -w .

install-lint:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

lint:
	golangci-lint run --fix

lint-check:
	golangci-lint run

setup: install-lint
	go install github.com/cosmtrek/air@latest
