.PHONY: build test run up down

build:
	go build -o bin/api ./cmd/api

test:
	go test ./... -race -cover

run:
	APP_ENV=development go run ./cmd/api

up: 
	docker compose -f docker-compose.dev.yml up -d

down: 
	docker compose -f docker-compose.dev.yml down