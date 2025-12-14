.PHONY: setup up down server test clean mocks

GO_BINARY := go
GO_MOD_DOWNLOAD := $(GO_BINARY) mod download
GO_GET := $(GO_BINARY) get -u
GO_INSTALL := $(GO_BINARY) install

DOCKER_COMPOSE := docker compose -f deploy/docker-compose.yml

# Go build commands
GO_RUN_SERVER := $(GO_BINARY) run cmd/server/main.go
GO_TEST := $(GO_BINARY) test -v -race ./...

# Mock generation
MOCKGEN := $(GO_BINARY) run go.uber.org/mock/mockgen@latest
MOCKS_DIR := internal/mocks
USER_REPO_SRC := internal/repository/user_repo.go
USER_SERVICE_SRC := internal/service/user_service.go
USER_REPO_MOCK_DEST := $(MOCKS_DIR)/user_repo_mock.go
USER_SERVICE_MOCK_DEST := $(MOCKS_DIR)/user_service_mock.go

# setup: Install dependencies and generate mocks
setup: $(MOCKGEN)
	@echo "Installing Go dependencies..."
	$(GO_MOD_DOWNLOAD)
	$(GO_GET) gorm.io/gorm gorm.io/driver/postgres github.com/spf13/viper github.com/stretchr/testify golang.org/x/crypto/bcrypt github.com/golang-jwt/jwt/v5 github.com/gin-gonic/gin go.uber.org/mock/mockgen
	@echo "Generating mocks..."
	mkdir -p $(MOCKS_DIR)
	$(MOCKGEN) -source=$(USER_REPO_SRC) -destination=$(USER_REPO_MOCK_DEST) -package=mocks
	$(MOCKGEN) -source=$(USER_SERVICE_SRC) -destination=$(USER_SERVICE_MOCK_DEST) -package=mocks

# up: Start Docker Compose services
up:
	@echo "Starting Docker services..."
	$(DOCKER_COMPOSE) up -d --wait

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

# clean: Clean up generated files
clean:
	@echo "Cleaning up generated files..."
	rm -f $(USER_REPO_MOCK_DEST)
	rm -f $(USER_SERVICE_MOCK_DEST)
	rm -rf $(MOCKS_DIR)
	$(GO_BINARY) clean -testcache
