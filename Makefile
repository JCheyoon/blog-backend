# .env는 선택이다(필수 환경 변수 없음). 있으면 읽는다
-include .env
export

.PHONY: run build docker-up docker-down

run:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api

docker-up:
	docker compose up --build

docker-down:
	docker compose down
