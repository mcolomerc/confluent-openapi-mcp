#!/bin/bash

# LLM Detection Setup Script
# This script helps you set up external LLM-based prompt injection detection

set -e

echo "🔒 Setting up LLM-based Prompt Injection Detection"
echo

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    exit 1
fi

# Check if Docker is running
if ! docker info &> /dev/null; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

echo "✅ Docker is available"

# Function to check if port is available
check_port() {
    local port=$1
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        return 1
    else
        return 0
    fi
}

# Check if port 11434 is available
if ! check_port 11434; then
    echo "⚠️  Port 11434 is already in use. Stopping any existing Ollama container..."
    docker stop ollama-security 2>/dev/null || true
    docker rm ollama-security 2>/dev/null || true
fi

echo "🚀 Starting Ollama container..."
docker run -d \
    --name ollama-security \
    -p 11434:11434 \
    -v ollama-data:/root/.ollama \
    --restart unless-stopped \
    ollama/ollama:latest

echo "⏳ Waiting for Ollama to start..."
sleep 10

# Wait for Ollama to be ready
max_attempts=30
attempt=0
while [ $attempt -lt $max_attempts ]; do
    if curl -s http://localhost:11434/api/tags >/dev/null 2>&1; then
        echo "✅ Ollama is ready!"
        break
    fi
    echo "   Waiting... (attempt $((attempt + 1))/$max_attempts)"
    sleep 2
    attempt=$((attempt + 1))
done

if [ $attempt -eq $max_attempts ]; then
    echo "❌ Ollama failed to start after 60 seconds"
    exit 1
fi

echo "📥 Downloading security detection model..."
docker exec ollama-security ollama pull llama3.2:1b
#ollama run llama-guard3:1b

echo "🧪 Testing the setup..."
response=$(curl -s -X POST http://localhost:11434/api/chat \
    -H "Content-Type: application/json" \
    -d '{
        "model": "llama3.2:1b",
        "messages": [
            {"role": "user", "content": "Hello, are you working?"}
        ],
        "stream": false
    }')

if echo "$response" | grep -q "choices"; then
    echo "✅ LLM detection is working!"
else
    echo "⚠️  LLM might not be working properly. Response: $response"
fi

echo
echo "🎉 Setup complete!"
echo
echo "To enable LLM detection in your MCP server, add these to your .env file:"
echo
echo "LLM_DETECTION_ENABLED=true"
echo "LLM_DETECTION_URL=http://localhost:11434/api/chat"
echo "LLM_DETECTION_MODEL=llama3.2:1b"
echo "LLM_DETECTION_TIMEOUT=10"
echo
echo "Available commands:"
echo "  📊 Check status:     curl http://localhost:11434/api/tags"
echo "  🔄 Restart:          docker restart ollama-security"
echo "  🛑 Stop:             docker stop ollama-security"
echo "  📋 View logs:        docker logs ollama-security"
echo "  🗑️  Remove:          docker stop ollama-security && docker rm ollama-security"
echo
echo "For more models, see: https://ollama.ai/library"
echo "Example: docker exec ollama-security ollama pull llama3.2:3b"
