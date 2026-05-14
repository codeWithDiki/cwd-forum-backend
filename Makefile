# Variables
APP_NAME=cwd-forum
MIGRATE_CMD=go run cmd/migrate/main.go
SERVER_CMD=go run cmd/server/main.go

# Default target
.PHONY: all
all: run

# Run the server
.PHONY: run
run:
	$(SERVER_CMD)

# Run migrations and seeders
.PHONY: migrate
migrate:
	$(MIGRATE_CMD)

# Build the application
.PHONY: build
build:
	go build -o bin/$(APP_NAME) cmd/server/main.go

# Clean build artifacts
.PHONY: clean
clean:
	rm -rf bin/
