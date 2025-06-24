#!/bin/bash

# Script to stop all development servers
echo "🛑 Stopping all development servers..."

# Kill any running mcp-server processes
pkill -f "mcp-server" 2>/dev/null || true

# Kill any running air processes
pkill -f "air" 2>/dev/null || true

# Wait a moment for processes to stop
sleep 1

# Check if port 8080 is still in use and force kill if needed
if lsof -i :8080 >/dev/null 2>&1; then
    echo "🔨 Force killing processes on port 8080..."
    lsof -ti :8080 | xargs kill -9 2>/dev/null || true
fi

echo "✅ All servers stopped"
