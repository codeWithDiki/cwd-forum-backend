# Variables
APP_NAME=cwd-forum
MIGRATE_CMD=go run cmd/migrate/main.go
SERVER_CMD=go run cmd/server/main.go
GOLANGCI_LINT=golangci-lint

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

# Run lint checks
.PHONY: lint
lint:
	$(GOLANGCI_LINT) run

# Apply automatic lint fixes
.PHONY: fix
fix:
	$(GOLANGCI_LINT) run --fix

# Clean build artifacts
.PHONY: clean
clean:
	rm -rf bin/
