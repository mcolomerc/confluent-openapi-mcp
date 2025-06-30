# External LLM-Based Prompt Injection Detection

The MCP server supports optional external LLM-based prompt injection detection in addition to the built-in regex patterns. This provides more sophisticated analysis of potentially malicious inputs.

## Overview

The system combines two detection methods:
1. **Regex-based detection** (fast, built-in patterns)
2. **LLM-based detection** (optional, more sophisticated analysis)

When both are enabled, the system uses either detection method to flag malicious content, providing enhanced security coverage.

## Setup with Docker (Ollama)

### 1. Run Ollama with Docker

```bash
# Pull and run Ollama container
docker run -d \
  --name ollama-security \
  -p 11434:11434 \
  -v ollama-data:/root/.ollama \
  --restart unless-stopped \
  ollama/ollama:latest

# Pull a lightweight model for security analysis
docker exec ollama-security ollama pull llama3.2:1b

# Optional: Pull a larger model for better accuracy
docker exec ollama-security ollama pull llama3.2:3b
```

### 2. Test the Ollama Installation

```bash
# Test that Ollama is running
curl -X POST http://localhost:11434/api/chat \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama3.2:1b",
    "messages": [
      {"role": "user", "content": "Hello, are you working?"}
    ],
    "stream": false
  }'
```

### 3. Configure the MCP Server

Add environment variables to your `.env` file:

```bash
# Enable LLM-based detection
LLM_DETECTION_ENABLED=true

# Ollama endpoint (default)
LLM_DETECTION_URL=http://localhost:11434/api/chat

# Model to use for detection
LLM_DETECTION_MODEL=llama3.2:1b

# Timeout for LLM requests (seconds)
LLM_DETECTION_TIMEOUT=10
```

### 4. Alternative: Using OpenAI-Compatible APIs

You can also use other OpenAI-compatible APIs:

```bash
# For OpenAI
LLM_DETECTION_URL=https://api.openai.com/v1/chat/completions
LLM_DETECTION_MODEL=gpt-3.5-turbo
LLM_DETECTION_API_KEY=your-openai-api-key

# For local models via text-generation-webui
LLM_DETECTION_URL=http://localhost:5000/v1/chat/completions
LLM_DETECTION_MODEL=your-local-model
```

## Configuration Options

The LLM detection can be configured programmatically:

```go
detector := NewInjectionDetection()

// Enable LLM detection with custom settings
detector.ConfigureLLM(ExternalLLMConfig{
    Enabled:    true,
    URL:        "http://localhost:11434/api/chat",
    Model:      "llama3.2:1b",
    TimeoutSec: 10,
    APIKey:     "", // Optional, for APIs that require authentication
})

// Or use the simple enable method
detector.EnableLLMDetection("http://localhost:11434/api/chat", "llama3.2:1b")
```

## Model Recommendations

### For Development/Testing
- **llama3.2:1b**: Fastest, smallest model (~1.3GB)
- **llama3.2:3b**: Better accuracy, still fast (~2.0GB)

### For Production
- **llama3.1:8b**: High accuracy for security tasks (~4.7GB)
- **mistral:7b**: Good alternative with strong security understanding (~4.1GB)

### Cloud Options
- **OpenAI GPT-3.5-turbo**: Fast and accurate, requires API key
- **OpenAI GPT-4**: Highest accuracy, more expensive

## Docker Compose Setup

Add to your `docker-compose.yml`:

```yaml
version: '3.8'
services:
  mcp-server:
    # your existing mcp-server config
    environment:
      - LLM_DETECTION_ENABLED=true
      - LLM_DETECTION_URL=http://ollama:11434/api/chat
      - LLM_DETECTION_MODEL=llama3.2:1b
    depends_on:
      - ollama

  ollama:
    image: ollama/ollama:latest
    container_name: ollama-security
    ports:
      - "11434:11434"
    volumes:
      - ollama-data:/root/.ollama
    restart: unless-stopped

volumes:
  ollama-data:
```

Initialize the model after starting:

```bash
docker-compose up -d
docker-compose exec ollama ollama pull llama3.2:1b
```

## Performance Considerations

1. **Latency**: LLM detection adds 100ms-2s per request depending on model size
2. **Memory**: Models require 1-8GB RAM depending on size
3. **Fallback**: If LLM detection fails, regex detection still works
4. **Caching**: Consider implementing response caching for common inputs

## Security Benefits

LLM-based detection can catch:
- **Sophisticated social engineering**: Complex manipulation attempts
- **Context-aware attacks**: Attacks that require understanding context
- **Novel injection patterns**: New attack vectors not covered by regex
- **Semantic analysis**: Understanding intent rather than just pattern matching

## Monitoring and Logging

The system provides detailed detection results:

```go
result := detector.DetectInjection(userInput)
if result.Detected {
    log.Printf("Malicious input detected:")
    log.Printf("  Regex patterns: %d", len(result.Patterns))
    
    if result.LLMResult != nil {
        log.Printf("  LLM confidence: %.2f", result.LLMResult.Confidence)
        log.Printf("  LLM category: %s", result.LLMResult.Category)
        log.Printf("  LLM explanation: %s", result.LLMResult.Explanation)
    }
}
```

## Troubleshooting

### Ollama Connection Issues
```bash
# Check if Ollama is running
curl http://localhost:11434/api/tags

# Check Docker logs
docker logs ollama-security

# Restart Ollama
docker restart ollama-security
```

### Model Download Issues
```bash
# Manual model download
docker exec -it ollama-security ollama pull llama3.2:1b

# Check available models
docker exec ollama-security ollama list
```

### Performance Issues
- Use smaller models (1b-3b parameters) for faster responses
- Implement request caching
- Consider async detection for non-critical paths
- Monitor resource usage

## Cost Considerations

### Local Models (Ollama)
- **Pros**: No API costs, private, customizable
- **Cons**: Requires local compute resources

### Cloud APIs
- **Pros**: No local resources needed, highest accuracy
- **Cons**: API costs, data sent to third party

Choose based on your security requirements, budget, and infrastructure preferences.
