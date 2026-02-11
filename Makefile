# Makefile
.PHONY: build run dev migrate-up migrate-down generate test clean docker-up docker-down docker-build lint fmt

BINARY_NAME=aogeri-api
BIN_DIR=bin

build:
	@echo "Building..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/api

run: build
	@echo "Running..."
	@./$(BIN_DIR)/$(BINARY_NAME)

dev:
	@echo "Running in development mode..."
	@go run ./cmd/api

migrate-up:
	@echo "Running migrations up..."
	@goose -dir internal/db/migrations postgres "user=postgres dbname=aogeri sslmode=disable" up

migrate-down:
	@echo "Running migrations down..."
	@goose -dir internal/db/migrations postgres "user=postgres dbname=aogeri sslmode=disable" down

generate:
	@echo "Generating SQLC code..."
	@sqlc generate

test:
	@echo "Running tests..."
	@go test ./... -v

fmt:
	@gofmt -w .

clean:
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR)

docker-up:
	@echo "Starting Docker containers..."
	@docker compose up -d --build

docker-down:
	@echo "Stopping Docker containers..."
	@docker compose down

docker-build:
	@echo "Building Docker image..."
	@docker compose build

lint:
	@golangci-lint run