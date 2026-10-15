.PHONY: build test run

build:
	go build -o bin/api ./cmd/api

test:
	go test ./... -race -cover

run:
	APP_ENV=development go run ./cmd/api
