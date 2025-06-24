#!/bin/bash

# Docker deployment script for Confluent MCP Server
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
IMAGE_NAME="confluent-mcp-server"
CONTAINER_NAME="confluent-mcp-server"
ENV_FILE=".env"

# Functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_requirements() {
    log_info "Checking requirements..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed or not in PATH"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose is not installed or not in PATH"
        exit 1
    fi
    
    log_info "Requirements check passed"
}

check_env_file() {
    if [ ! -f "$ENV_FILE" ]; then
        log_warn "Environment file $ENV_FILE not found"
        log_info "Creating .env file from .env.docker.example..."
        if [ -f ".env.docker.example" ]; then
            cp .env.docker.example .env
            log_warn "Please edit .env file with your actual configuration values"
        else
            log_error ".env.docker.example not found. Cannot create .env file."
            exit 1
        fi
    fi
}

build_image() {
    log_info "Building Docker image..."
    docker build -t $IMAGE_NAME .
    log_info "Docker image built successfully"
}

run_container() {
    log_info "Starting container with Docker Compose..."
    docker-compose up -d
    log_info "Container started successfully"
}

stop_container() {
    log_info "Stopping container..."
    docker-compose down
    log_info "Container stopped"
}

show_logs() {
    log_info "Showing container logs..."
    docker-compose logs -f
}

show_status() {
    log_info "Container status:"
    docker-compose ps
    
    echo ""
    log_info "Container health:"
    docker inspect --format='{{.State.Health.Status}}' $CONTAINER_NAME 2>/dev/null || echo "Health check not available"
}

cleanup() {
    log_info "Cleaning up Docker resources..."
    docker-compose down -v --remove-orphans
    docker rmi $IMAGE_NAME 2>/dev/null || true
    log_info "Cleanup completed"
}

# Command handling
case "${1:-}" in
    "build")
        check_requirements
        build_image
        ;;
    "run")
        check_requirements
        check_env_file
        run_container
        ;;
    "stop")
        stop_container
        ;;
    "restart")
        stop_container
        sleep 2
        run_container
        ;;
    "logs")
        show_logs
        ;;
    "status")
        show_status
        ;;
    "clean")
        cleanup
        ;;
    "deploy")
        check_requirements
        check_env_file
        build_image
        run_container
        show_status
        ;;
    *)
        echo "Usage: $0 {build|run|stop|restart|logs|status|clean|deploy}"
        echo ""
        echo "Commands:"
        echo "  build     - Build the Docker image"
        echo "  run       - Start the container"
        echo "  stop      - Stop the container"
        echo "  restart   - Restart the container"
        echo "  logs      - Show container logs"
        echo "  status    - Show container status"
        echo "  clean     - Clean up Docker resources"
        echo "  deploy    - Full deployment (build + run + status)"
        exit 1
        ;;
esac
