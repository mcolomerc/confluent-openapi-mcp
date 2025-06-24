# Docker Deployment Guide

This guide provides detailed instructions for deploying the Confluent MCP Server using Docker.

## Quick Start

1. **Setup environment**:
   ```bash
   cp .env.docker.example .env
   # Edit .env with your Confluent Cloud credentials
   ```

2. **Deploy**:
   ```bash
   make docker-dev
   ```

3. **Test**:
   ```bash
   curl http://localhost:8080/tools
   ```

## Environment Configuration

### Required Variables

Copy `.env.docker.example` to `.env` and configure:

```bash
# Confluent Cloud
CONFLUENT_ENV_ID=env-xxxxxx
CONFLUENT_CLOUD_API_KEY=your-cloud-api-key
CONFLUENT_CLOUD_API_SECRET=your-cloud-api-secret

# Kafka Cluster
BOOTSTRAP_SERVERS=pkc-xxxxxx.region.provider.confluent.cloud:9092
KAFKA_API_KEY=your-kafka-api-key
KAFKA_API_SECRET=your-kafka-api-secret
KAFKA_REST_ENDPOINT=https://pkc-xxxxxx.region.provider.confluent.cloud
KAFKA_CLUSTER_ID=lkc-xxxxxx

# Schema Registry
SCHEMA_REGISTRY_API_KEY=your-sr-api-key
SCHEMA_REGISTRY_API_SECRET=your-sr-api-secret
SCHEMA_REGISTRY_ENDPOINT=https://psrc-xxxxxx.region.provider.confluent.cloud

# ... (see .env.docker.example for complete list)
```

## Deployment Options

### Option 1: Make Commands (Recommended)

```bash
# Build and start
make docker-dev

# Individual operations
make docker-build    # Build image
make docker-run      # Start container
make docker-stop     # Stop container
make docker-logs     # View logs
make docker-health   # Check status
make docker-clean    # Clean up
```

### Option 2: Docker Compose

```bash
# Start services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

### Option 3: Deployment Script

```bash
# Full deployment
./scripts/docker-deploy.sh deploy

# Individual commands
./scripts/docker-deploy.sh build
./scripts/docker-deploy.sh run
./scripts/docker-deploy.sh status
```

### Option 4: Manual Docker

```bash
# Build
docker build -t confluent-mcp-server .

# Run
docker run -d \
  --name confluent-mcp-server \
  --env-file .env \
  -p 8080:8080 \
  confluent-mcp-server
```

## Testing the Deployment

### Health Check

```bash
# Check if container is running
docker ps | grep confluent-mcp-server

# Test HTTP endpoint
curl http://localhost:8080/tools

# Check health status
docker inspect --format='{{.State.Health.Status}}' confluent-mcp-server
```

### API Testing

```bash
# List available tools
curl http://localhost:8080/tools

# Get prompts
curl http://localhost:8080/prompts

# Test MCP protocol
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/list"}' | \
  curl -X POST -H "Content-Type: application/json" -d @- \
  http://localhost:8080/
```

## Container Management

### Viewing Logs

```bash
# Real-time logs
docker logs -f confluent-mcp-server

# Last 100 lines
docker logs --tail 100 confluent-mcp-server

# With docker-compose
docker-compose logs -f
```

### Resource Monitoring

```bash
# Container stats
docker stats confluent-mcp-server

# Detailed inspection
docker inspect confluent-mcp-server
```

### Updating the Container

```bash
# Pull latest changes and rebuild
git pull
make docker-clean
make docker-dev
```

## Development with Docker

### Development Compose

For development, use the dev compose file:

```bash
# Start development environment
docker-compose -f docker-compose.dev.yml up -d

# View logs with debug level
docker-compose -f docker-compose.dev.yml logs -f
```

### Volume Mounts

The container mounts:
- `./prompts:/app/prompts:ro` - Prompt templates
- `./api-spec:/app/api-spec:ro` - OpenAPI specification
- `./logs:/app/logs` - Log files (dev only)

### Debugging

```bash
# Execute shell in running container
docker exec -it confluent-mcp-server sh

# Run with different entrypoint for debugging
docker run -it --rm --entrypoint sh confluent-mcp-server
```

## Production Considerations

### Security

- Container runs as non-root user (`appuser`)
- Minimal attack surface (Alpine Linux base)
- Secrets should be managed externally in production

### Performance

- Resource limits configured in docker-compose.yml
- Memory limit: 256MB (adjust based on usage)
- CPU limit: 0.5 cores (adjust based on load)

### Scaling

```bash
# Scale with docker-compose
docker-compose up -d --scale mcp-server=3

# Use load balancer in front of multiple instances
```

### Monitoring

Set up monitoring for:
- Container health status
- Memory and CPU usage
- API response times
- Error rates

## Troubleshooting

### Common Issues

1. **Container won't start**:
   ```bash
   # Check logs
   docker logs confluent-mcp-server
   
   # Verify environment variables
   docker inspect confluent-mcp-server | grep -A 20 '"Env"'
   ```

2. **Health check failing**:
   ```bash
   # Test endpoint manually
   curl http://localhost:8080/tools
   
   # Check if port is accessible
   docker port confluent-mcp-server
   ```

3. **API credentials issues**:
   ```bash
   # Verify environment file
   cat .env
   
   # Test with debug logging
   docker run --rm --env-file .env -e LOG=DEBUG confluent-mcp-server
   ```

### Getting Help

- Check container logs: `docker logs confluent-mcp-server`
- Verify configuration: `docker inspect confluent-mcp-server`
- Test endpoints: `curl http://localhost:8080/tools`
- Review documentation: See main README.md

## Cleanup

```bash
# Stop and remove container
make docker-clean

# Or manually
docker-compose down -v --remove-orphans
docker rmi confluent-mcp-server
```
