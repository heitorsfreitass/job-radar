.PHONY: up down build test test-integration lint migrate-up migrate-down

up:
	docker compose -f deployments/docker-compose.yml up --build

down:
	docker compose -f deployments/docker-compose.yml down

build:
	go build ./...

test:
	go test ./... -race -cover

# Requires Docker: spins up real Postgres/Redis containers via
# testcontainers-go for the duration of the run.
test-integration:
	go test -tags=integration ./... -race -cover

lint:
	gofmt -l .
	go vet ./...

MIGRATE_DB_URL ?= postgres://job_radar:job_radar@localhost:5433/job_radar?sslmode=disable

migrate-up:
	migrate -path migrations -database "$(MIGRATE_DB_URL)" up

migrate-down:
	migrate -path migrations -database "$(MIGRATE_DB_URL)" down 1
