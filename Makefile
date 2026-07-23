.PHONY: dev build compose-up compose-down

dev:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

compose-up:
	docker compose up -d

compose-down:
	docker compose down
