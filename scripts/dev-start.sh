#!/bin/bash

# Script to start development server safely
echo "🚀 Starting development server..."

# Function to kill existing servers
cleanup_servers() {
    echo "🧹 Cleaning up existing servers..."
    
    # Kill any running mcp-server processes
    pkill -f "mcp-server" 2>/dev/null || true
    
    # Wait a moment for processes to stop
    sleep 1
    
    # Check if port 8080 is still in use
    if lsof -i :8080 >/dev/null 2>&1; then
        echo "⚠️  Port 8080 still in use, killing processes..."
        lsof -ti :8080 | xargs kill -9 2>/dev/null || true
        sleep 1
    fi
    
    echo "✅ Cleanup complete"
}

# Function to start air development server
start_air() {
    echo "🔄 Starting air development server..."
    
    # Try different air locations
    AIR_BIN=""
    if command -v air >/dev/null 2>&1; then
        AIR_BIN="air"
    elif [ -f "$HOME/go/bin/air" ]; then
        AIR_BIN="$HOME/go/bin/air"
    elif [ -f "$(go env GOPATH)/bin/air" ]; then
        AIR_BIN="$(go env GOPATH)/bin/air"
    else
        echo "❌ Air not found. Please install: go install github.com/air-verse/air@latest"
        echo "   Or add $HOME/go/bin to your PATH"
        exit 1
    fi
    
    echo "Using air from: $AIR_BIN"
    $AIR_BIN
}

# Main execution
cleanup_servers
start_air
