.PHONY: up down build test lint migrate-up migrate-down

up:
	docker compose -f deployments/docker-compose.yml up --build

down:
	docker compose -f deployments/docker-compose.yml down

build:
	go build ./...

test:
	go test ./... -race -cover

lint:
	gofmt -l .
	go vet ./...

MIGRATE_DB_URL ?= postgres://job_radar:job_radar@localhost:5432/job_radar?sslmode=disable

migrate-up:
	migrate -path migrations -database "$(MIGRATE_DB_URL)" up

migrate-down:
	migrate -path migrations -database "$(MIGRATE_DB_URL)" down 1
