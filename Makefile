.PHONY: build run dev test clean deps services-up services-down services-logs migrate

APP_NAME=goapi
MAIN_FILE=cmd/api/main.go

# Build application
build:
	@go build -o bin/$(APP_NAME) $(MAIN_FILE)

# Run application (local)
run:
	@go run $(MAIN_FILE)

# Development dengan hot reload
dev:
	@which air > /dev/null || (echo "Installing air..." && go install github.com/air-verse/air@latest)
	@$(shell go env GOPATH)/bin/air

# Download dependencies
deps:
	@go mod download
	@go mod tidy

# Run tests
test:
	@go test -v ./...

# Clean
clean:
	@rm -rf bin/

# Docker Services Only
up:
	@echo "🚀 Starting PostgreSQL and Redis..."
	@docker compose up -d
	@echo "✅ Services running!"
	@echo "   PostgreSQL: localhost:5433"
	@echo "   Redis: localhost:6380"
	@echo "   Adminer: http://localhost:8081"

down:
	@echo "🛑 Stopping services..."
	@docker compose down


services-logs:
	@docker compose logs -f

# Check services health
services-status:
	@docker compose ps

# Database commands
migrate-up:
	@which migrate > /dev/null || (echo "Install migrate: https://github.com/golang-migrate/migrate" && exit 1)
	@migrate -path migrations -database "postgres://postgres:postgres@localhost:5433/goapi?sslmode=disable" up

migrate-down:
	@migrate -path migrations -database "postgres://postgres:postgres@localhost:5433/goapi?sslmode=disable" down

# Full setup untuk development baru
setup: deps services-up
	@echo "⏳ Waiting for PostgreSQL..."
	@sleep 3
	@echo "✅ Setup complete! Run 'make dev' to start development server"