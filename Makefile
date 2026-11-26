include .env
export

.PHONY: run build docker-up docker-down migrate-up migrate-down

run:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api

docker-up:
	docker compose up --build

docker-down:
	docker compose down

# requires golang-migrate CLI: https://github.com/golang-migrate/migrate
migrate-up:
	migrate -path migrations -database "$$DATABASE_URL" up

migrate-down:
	migrate -path migrations -database "$$DATABASE_URL" down 1
