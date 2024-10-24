# Simple Makefile for a Go project

# Build the application
all: build test

build:
	@echo "Building..."


	@go build -o main cmd/api/main.go

# Run the application
run:
	@go run cmd/api/main.go
# Create DB container
docker-run:
	@if docker compose up -d --build 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose up --build; \
	fi

# Shutdown DB container
docker-down:
	@if docker compose down 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose down; \
	fi

# create DB
db-create:
	@echo "Creating Database..."
	@go run cmd/create_db/main.go

# migration
db-migrate:
	@echo "executing Migration..."
	@go run cmd/migrate/main.go --direction=up

db-rollback:
	@echo "executing Rollback Migration..."
	@go run cmd/migrate/main.go --direction=down --steps=$(or ${s},1)
# make create-migration t=create_users_table
create-migration:
	@echo "executing Creating Migration..."
	@migrate create -ext sql -dir ./migrations -seq $(t)

seed:
	@echo "executing Seed..."
	@go run cmd/seed/main.go

seed-truncate:
	@echo "executing Seed Truncate..."
	@go run cmd/seed/main.go --truncate

# Test the application
test:
	@echo "Testing..."
	@go test ./...

test-v:
	@echo "Testing with v..."
	@go test ./... -v

test-coverage-html:
	@echo "Testing with coverage..."
	@go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out

# Integrations Tests for the application
itest:
	@echo "Running integration tests..."
	@go test ./internal/database -v

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main

# Live Reload
watch:
	@if command -v air > /dev/null; then \
            air; \
            echo "Watching...";\
        else \
            read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
            if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
                go install github.com/air-verse/air@latest; \
                air; \
                echo "Watching...";\
            else \
                echo "You chose not to install air. Exiting..."; \
                exit 1; \
            fi; \
        fi

.PHONY: all build run test clean watch docker-run docker-down itest
