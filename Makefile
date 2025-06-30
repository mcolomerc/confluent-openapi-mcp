# Makefile for building the MCP server Go binary

BINARY_NAME = mcp-server
CMD_PATH = ./cmd/main.go
BUILD_DIR = ./bin
PID_FILE = $(BUILD_DIR)/$(BINARY_NAME).pid

.PHONY: all build clean dev watch run stop restart test install-tools

all: build

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@if [ -d "./prompts" ]; then \
		echo "Copying prompts to build directory..."; \
		cp -r ./prompts $(BUILD_DIR)/; \
	fi
	@if [ -d "./directives" ]; then \
		echo "Copying directives to build directory..."; \
		cp -r ./directives $(BUILD_DIR)/; \
	fi

# Build and restart MCP servers for VS Code
build-mcp: build
	@echo "Restarting VS Code MCP servers..."
	@./scripts/restart-mcp.sh

clean:
	rm -rf $(BUILD_DIR)

# Install development tools
install-tools:
	@echo "Installing development tools..."
	@which air > /dev/null || go install github.com/air-verse/air@latest
	@which entr > /dev/null || (echo "Please install entr: brew install entr (macOS) or apt-get install entr (Linux)")
	@echo "Development tools installed!"

# Development mode with automatic reload using air
dev:
	@if which air > /dev/null 2>&1; then \
		echo "Starting development server with air..."; \
		air; \
	else \
		echo "Air not found, falling back to watch mode..."; \
		$(MAKE) watch; \
	fi

# Alternative development mode using entr (if air doesn't work)
watch:
	@which entr > /dev/null || (echo "Please install entr: brew install entr" && exit 1)
	@echo "Watching for changes... Press Ctrl+C to stop"
	find . -name "*.go" | entr -r make run-dev

# Simple development mode without external dependencies
dev-simple:
	@echo "Starting simple development mode..."
	@echo "Note: You'll need to manually restart after changes"
	$(MAKE) run

# Run the server directly (for development)
run:
	go run $(CMD_PATH)

# Run the server in development mode (with build step)
run-dev: build
	@$(MAKE) stop > /dev/null 2>&1 || true
	@echo "Starting server..."
	./$(BUILD_DIR)/$(BINARY_NAME) &
	@echo $$! > $(PID_FILE)
	@echo "Server started with PID $$(cat $(PID_FILE))"

# Stop the running server
stop:
	@if [ -f $(PID_FILE) ]; then \
		PID=$$(cat $(PID_FILE)); \
		if ps -p $$PID > /dev/null 2>&1; then \
			echo "Stopping server (PID: $$PID)..."; \
			kill $$PID; \
			rm -f $(PID_FILE); \
			echo "Server stopped."; \
		else \
			echo "Server not running (stale PID file removed)"; \
			rm -f $(PID_FILE); \
		fi \
	else \
		echo "No PID file found. Server may not be running."; \
	fi

# Restart the server
restart: stop run-dev

# Run tests
test:
	go test ./...

# Run tests with coverage
test-coverage:
	go test -cover ./...

# Run tests in watch mode
test-watch:
	@which entr > /dev/null || (echo "Please install entr: brew install entr" && exit 1)
	find . -name "*.go" | entr -c go test ./...

# Docker targets
.PHONY: docker-build docker-run docker-stop docker-logs docker-clean docker-dev

# Build Docker image
docker-build:
	docker build -t confluent-mcp-server .

# Run Docker container
docker-run:
	docker-compose up -d

# Stop Docker container
docker-stop:
	docker-compose down

# View Docker logs
docker-logs:
	docker-compose logs -f

# Clean Docker resources
docker-clean:
	docker-compose down -v --remove-orphans
	docker rmi confluent-mcp-server 2>/dev/null || true

# Development with Docker (build and run)
docker-dev: docker-build docker-run

# Docker healthcheck
docker-health:
	docker-compose ps
