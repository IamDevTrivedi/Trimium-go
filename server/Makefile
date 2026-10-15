.PHONY: build test run up down fmt fmt-check

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

fmt-check:
	@out=$$(gofmt -l .); \
	if [ -n "$$out" ]; then \
		echo "Unformatted files:"; echo "$$out"; exit 1; \
	fi

fmt:
	gofmt -w .