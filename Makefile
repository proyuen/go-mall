.PHONY: setup up down server test clean mocks

GO_BINARY := go
GO_MOD_DOWNLOAD := $(GO_BINARY) mod download

DOCKER_COMPOSE := docker compose -f deploy/docker-compose.yml

# Go build commands
GO_RUN_SERVER := $(GO_BINARY) run cmd/server/main.go
GO_TEST := $(GO_BINARY) test -v -race ./...

# setup: Install dependencies and tools
setup:
	@echo "Installing Go dependencies..."
	$(GO_MOD_DOWNLOAD)
	@echo "Installing mockgen tool..."
	$(GO_BINARY) install go.uber.org/mock/mockgen@latest

# mocks: Generate mocks using go generate
mocks:
	@echo "Generating mocks..."
	$(GO_BINARY) generate ./...

# up: Start Docker Compose services
up:
	@echo "Starting Docker services..."
	$(DOCKER_COMPOSE) up -d --wait
	@echo "Waiting for database to be ready..."
	sleep 5

# down: Stop and remove Docker Compose services
down:
	@echo "Stopping Docker services..."
	$(DOCKER_COMPOSE) down

# server: Run the Go API server
server:
	@echo "Running Go API server..."
	$(GO_RUN_SERVER)

# test: Run all Go tests
test: up
	@echo "Running all Go tests..."
	$(GO_TEST)

# clean: Clean up
clean:
	@echo "Cleaning up..."
	$(GO_BINARY) clean -testcache