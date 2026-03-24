.PHONY: build run test clean tidy db-up db-down migrate-up migrate-down

# Build the application
build:
	go build -o bin/api ./cmd/api

# Run the application
run:
	go run ./cmd/api

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Download and verify dependencies
tidy:
	go mod tidy
	go mod verify

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Run with hot reload (requires air)
dev:
	air

# Database commands
# Requires golang-migrate: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

db-up:
	docker-compose up -d

db-down:
	docker-compose down

DB_URL := postgres://postgres:postgres@localhost:5432/finance_tracker?sslmode=disable

migrate-up:
	migrate -path migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path migrations -database "$(DB_URL)" down

migrate-create:
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name
