# Confluent OpenAPI MCP Server

A Model Context Protocol (MCP) server that dynamically generates semantic tools from the Confluent Cloud OpenAPI specification. This server provides a bridge between MCP clients and the Confluent Cloud API, enabling AI agents to interact with Kafka clusters, Flink compute pools, Schema Registry, TableFlow, and other Confluent services through natural language interfaces.

## 📖 Quick Navigation

- 🚀 **[Building and Running](#building-and-running)** - Get started quickly
- 🔧 **[Configuration](#configuration)** - Environment setup
- 📚 **[Documentation](#-documentation)** - Complete guides and references
- 🐳 **[Docker Deployment](docs/DOCKER.md)** - Production deployment
- 📊 **[Monitoring](docs/MONITORING_STACK.md)** - Observability stack

## How It Works

### 1. OpenAPI Specification Loading

The server loads the Confluent Cloud OpenAPI specification from either:

- A local file (`api-spec/confluent-apispec.json` by default)
- A remote URL (specified via `OPENAPI_SPEC_URL` environment variable)

The OpenAPI spec is parsed to extract:

- API endpoints and their HTTP methods
- Request/response schemas
- Parameter definitions
- Security requirements

### 2. Semantic Tool Generation

The server transforms raw OpenAPI endpoints into semantic tools using intelligent mapping:

**Resource Extraction**: Analyzes API paths to identify resources (e.g., `topics`, `clusters`, `connectors`)

**Action Mapping**: Maps HTTP methods and paths to semantic actions:

- `POST` → `create` (for collection endpoints)
- `GET` → `list` (for collections) or `get` (for individual resources)
- `PUT/PATCH` → `update`
- `DELETE` → `delete`

**Tool Creation**: Generates MCP tools with names like:

- `create` - Create resources
- `list` - List resources
- `get` - Get individual resources
- `update` - Update resources
- `delete` - Delete resources

### 3. Request Processing

When a client invokes a tool, the server:

1. **Validates Parameters**: Checks for required parameters and applies defaults from configuration
2. **Auto-resolution**: Automatically resolves common parameters like `clusterId`, `environmentId` from configuration
3. **Schema Building**: Constructs request bodies according to OpenAPI schemas
4. **API Authentication**: Determines appropriate credentials (Cloud API keys vs Resource API keys)
5. **HTTP Request**: Executes the actual API call to Confluent Cloud
6. **Response Handling**: Returns formatted responses or error messages

### 4. Dual Server Architecture

The server runs both:

- **HTTP Server** (port 8080): For HTTP-based MCP clients
- **STDIO Server**: For standard input/output MCP communication

## Building and Running

### Prerequisites

- Go 1.19 or later
- Access to Confluent Cloud with API credentials

### Development Setup (Recommended)

For the best development experience with automatic rebuilding and restarting:

#### Option 1: Using Air (Recommended)

```bash
# Install development tools
make install-tools

# Start development server with auto-reload
make dev
```

This will:

- Watch for changes in `.go`, `.json`, and `.env` files
- Automatically rebuild and restart the server
- Display build errors and runtime logs
- Keep the server running until you stop it with `Ctrl+C`

#### Option 2: Using VS Code Tasks

1. Open the project in VS Code
2. Use `Cmd+Shift+P` (macOS) or `Ctrl+Shift+P` (Windows/Linux)
3. Select "Tasks: Run Task"
4. Choose "Dev: Start Auto-Reload Server"

The server will automatically start and reload on any code changes. You can also use:

- "Dev: Stop Server" - Stop the running server
- "Dev: Restart Server" - Manually restart the server
- "Build Server" - Build without running
- "Run Tests" - Execute all tests

#### Option 3: Manual File Watching

```bash
# Alternative using entr (requires: brew install entr)
make watch
```

### Build

```bash
# Using Makefile
make build

# Or directly with Go
go build -o bin/mcp-server cmd/main.go
```

### Run

```bash
# Development mode (auto-reload)
make dev

# Production mode (using the binary)
./bin/mcp-server

# Or directly with Go
go run cmd/main.go

# With custom environment file
go run cmd/main.go -env /path/to/your/.env
```

### Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run tests in watch mode (auto-rerun on changes)
make test-watch

# Or directly with Go
go test ./...
```

### Debugging in VS Code

1. Set breakpoints in your code
2. Press `F5` or use "Run and Debug" panel
3. Select "Debug MCP Server" configuration
4. The debugger will start with automatic building

## Configuration

The server requires multiple environment variables for proper operation. Create a `.env` file in the project root with the following parameters:

### Required Configuration

#### Confluent Cloud Control Plane

- **`CONFLUENT_CLOUD_API_KEY`**: Your Confluent Cloud API key for control plane operations
- **`CONFLUENT_CLOUD_API_SECRET`**: Your Confluent Cloud API secret
- **`CONFLUENT_ENV_ID`**: Environment ID (must start with `env-`)
  - Example: `env-12345`

#### Kafka Cluster

- **`BOOTSTRAP_SERVERS`**: Kafka bootstrap servers
  - Example: `pkc-abc123.us-west-2.aws.confluent.cloud:9092`
- **`KAFKA_API_KEY`**: Kafka cluster API key
- **`KAFKA_API_SECRET`**: Kafka cluster API secret
- **`KAFKA_REST_ENDPOINT`**: Kafka REST proxy endpoint
- **`KAFKA_CLUSTER_ID`**: Kafka cluster identifier
  - Example: `lkc-abc123`

#### Flink Compute Pool

- **`FLINK_ORG_ID`**: Flink organization ID
- **`FLINK_REST_ENDPOINT`**: Flink REST API endpoint
- **`FLINK_ENV_NAME`**: Flink environment name
- **`FLINK_DATABASE_NAME`**: Flink database name
- **`FLINK_API_KEY`**: Flink API key
- **`FLINK_API_SECRET`**: Flink API secret
- **`FLINK_COMPUTE_POOL_ID`**: Flink compute pool ID

#### Schema Registry

- **`SCHEMA_REGISTRY_API_KEY`**: Schema Registry API key
- **`SCHEMA_REGISTRY_API_SECRET`**: Schema Registry API secret
- **`SCHEMA_REGISTRY_ENDPOINT`**: Schema Registry endpoint
  - Example: `https://psrc-abc123.us-west-2.aws.confluent.cloud`

#### TableFlow

- **`TABLEFLOW_API_KEY`**: TableFlow API key
- **`TABLEFLOW_API_SECRET`**: TableFlow API secret

### Optional Configuration

- **`LOG`**: Log level (`DEBUG`, `INFO`, `WARN`, `ERROR`)
  - Default: `INFO`
- **`PROMPTS_FOLDER`**: Custom path to prompts folder (see [Prompt Support](#prompt-support) for details)
  - Default: Automatically uses `<executable-directory>/prompts` or `./prompts`
  - Example: `/path/to/custom/prompts`
- **`OPENAPI_SPEC_URL`**: Custom OpenAPI specification URL or path
  - Default: Uses local `api-spec/confluent-apispec.json`
  - Example: `https://api.confluent.cloud/openapi.json`
- **`DISABLE_RESOURCE_DISCOVERY`**: Disable automatic resource instance discovery (`true` or `false`)
  - Default: `false` (resource discovery enabled)
  - When `true`: Skips enumeration of individual resource instances for faster startup
  - When `false`: Discovers and registers all available resource instances as individual tools
  - Use `true` for development or when you only need basic CRUD operations

## Security Model

The server uses different credential types based on the API endpoint:

- **Cloud API Keys**: For control plane operations (creating clusters, environments)
- **Resource API Keys**: For data plane operations (topics, schemas, Flink queries)

Authentication is automatically selected based on the API path being accessed.

## Prompt Support

The server includes built-in support for external prompts per the MCP specification. This allows clients to access predefined prompt templates for common Confluent Cloud operations.

### Prompt Configuration

#### Environment Variables

- **`PROMPTS_FOLDER`**: Path to the folder containing prompt files
  - **Optional**: If not specified, the server automatically uses a default location
  - **Default Behavior**:
    - First tries: `<executable-directory>/prompts`
    - Falls back to: `./prompts` (current working directory)
  - **Custom Path Example**: `/path/to/custom/prompts`
  - **Note**: The server will gracefully handle missing prompt folders by returning an empty prompt list

#### Prompt File Format

Prompts are stored as `.txt` files in the prompts folder. Each file becomes a prompt with the filename (without extension) as the prompt name.

**Default Location**: When no `PROMPTS_FOLDER` is configured, the server automatically looks for prompts in:

1. `<executable-directory>/prompts` (where the `mcp-server` binary is located)
2. `./prompts` (current working directory as fallback)

Example prompt file structure:

```text
prompts/
├── kafka-topic-analysis.txt
├── environment-setup.txt
├── schema-registry-guide.txt
└── troubleshooting.txt
```

**Getting Started**: The repository includes example prompts in the `prompts/` folder that work out of the box without any configuration.

### Built-in Prompts

The server comes with several example prompts:

- **`kafka-topic-analysis`**: Guide for analyzing Kafka topic configurations and performance
- **`environment-setup`**: Step-by-step environment configuration instructions
- **`schema-registry-guide`**: Best practices for Schema Registry management

### HTTP Endpoints

#### List Available Prompts

```http
GET http://localhost:8080/prompts
```

**Response:**

```json
{
  "prompts": [
    {
      "name": "kafka-topic-analysis",
      "description": "Guide for analyzing Kafka topic configurations"
    },
    {
      "name": "environment-setup", 
      "description": "Environment configuration instructions"
    }
  ]
}
```

#### Get Prompt Content

```http
GET http://localhost:8080/prompts/{name}
```

**Example:**

```http
GET http://localhost:8080/prompts/kafka-topic-analysis
```

**Response:**

```json
{
  "name": "kafka-topic-analysis",
  "description": "Guide for analyzing Kafka topic configurations",
  "content": "You are a Kafka expert helping analyze topic configurations..."
}
```

### MCP Client Integration

For MCP clients, prompts are exposed through the standard MCP prompt methods:

- **`prompts/list`**: List available prompts
- **`prompts/get`**: Get specific prompt content

The server automatically reloads prompts when files are added, modified, or removed from the prompts folder.

### Creating Custom Prompts

1. Create a `.txt` file in your prompts folder
2. Add your prompt content to the file
3. The server will automatically detect and load the new prompt
4. Access via HTTP API or MCP client

**Example custom prompt file** (`prompts/my-custom-prompt.txt`):

```text
You are an expert Confluent Cloud administrator. Help the user with the following task:

{task_description}

Consider the following environment context:
- Environment: {environment_id}
- Cluster: {cluster_id}
- Region: {region}

Provide step-by-step guidance with specific API calls and configurations.
```

### Summary

**Zero Configuration**: The server works out of the box with included example prompts—no `PROMPTS_FOLDER` configuration required.

**Flexible Setup**: Easily customize prompt locations for advanced use cases while maintaining backward compatibility.

**Automatic Discovery**: The server intelligently locates prompts relative to the executable or working directory, making deployment simple across different environments.

## Docker Deployment

The MCP server can be easily deployed using Docker for containerized environments.

### Quick Start with Docker

1. **Build and run with Docker Compose**:

   ```bash
   # Clone the repository and navigate to the project directory
   cd mcp-server
   
   # Copy the example environment file
   cp .env.docker.example .env
   
   # Edit .env with your actual Confluent Cloud credentials
   nano .env
   
   # Build and start the container
   make docker-dev
   ```

2. **Check the status**:

   ```bash
   make docker-health
   ```

3. **View logs**:

   ```bash
   make docker-logs
   ```

### Docker Configuration

#### Environment File Setup

Create a `.env` file based on `.env.docker.example`:

```bash
# Copy the example file
cp .env.docker.example .env

# Edit with your credentials
vim .env
```

The Docker setup uses the same environment variables as the native deployment. See the [Configuration](#configuration) section for detailed parameter descriptions.

#### Docker Compose Services

The `docker-compose.yml` file defines:

- **Resource limits**: Memory (256M max) and CPU (0.5 cores max)
- **Health checks**: Automatic health monitoring
- **Volume mounts**: For prompts and API specifications
- **Port mapping**: Exposes port 8080 for HTTP access
- **Restart policy**: Automatically restarts on failure

#### Available Make Targets

```bash
# Build Docker image
make docker-build

# Start container (with docker-compose)
make docker-run

# Stop container
make docker-stop

# View logs
make docker-logs

# Clean up Docker resources
make docker-clean

# Full deployment (build + run + status)
make docker-dev

# Check container health
make docker-health
```

### Advanced Docker Usage

#### Using the Deployment Script

For more control, use the provided deployment script:

```bash
# Make the script executable (if not already)
chmod +x scripts/docker-deploy.sh

# Full deployment
./scripts/docker-deploy.sh deploy

# Or individual commands
./scripts/docker-deploy.sh build
./scripts/docker-deploy.sh run
./scripts/docker-deploy.sh status
./scripts/docker-deploy.sh logs
```

#### Manual Docker Commands

```bash
# Build the image
docker build -t confluent-mcp-server .

# Run the container
docker run -d \
  --name confluent-mcp-server \
  --env-file .env \
  -p 8080:8080 \
  -v $(pwd)/prompts:/app/prompts:ro \
  confluent-mcp-server

# View logs
docker logs -f confluent-mcp-server

# Stop and remove
docker stop confluent-mcp-server
docker rm confluent-mcp-server
```

#### Custom Configuration

To override specific configurations in Docker:

```bash
# Run with custom environment variables
docker run -d \
  --name confluent-mcp-server \
  -e CONFLUENT_ENV_ID=env-12345 \
  -e LOG=DEBUG \
  -p 8080:8080 \
  confluent-mcp-server
```

### Docker Image Details

The Docker image is built using a multi-stage approach:

- **Builder stage**: Uses Go 1.23 Alpine image to compile the binary
- **Runtime stage**: Uses scratch image for minimal size
- **Security**: Runs as non-root user (`appuser`)
- **Size**: Optimized for minimal image size (~15-20MB)
- **Dependencies**: Includes CA certificates for HTTPS requests

### Production Considerations

For production deployments:

1. **Resource Limits**: Adjust memory and CPU limits in `docker-compose.yml`
2. **Health Checks**: Configure appropriate health check intervals
3. **Logging**: Consider using centralized logging solutions
4. **Secrets Management**: Use Docker secrets or external secret managers
5. **Networking**: Configure appropriate network policies
6. **Monitoring**: Set up monitoring and alerting for the container

## API Tool Usage

Once running, the server exposes semantic tools that can be used by MCP clients:

## 📚 Documentation

### Core Documentation

- **[Development Guide](docs/DEVELOPMENT.md)** - Development setup, debugging, and workflow
- **[Docker Guide](docs/DOCKER.md)** - Docker setup and deployment instructions
- **[Monitoring](docs/MONITORING.md)** - Basic monitoring setup and resource tracking

### Monitoring & Observability

- **[Prometheus Integration](docs/PROMETHEUS.md)** - Prometheus metrics export and configuration
- **[Monitoring Stack](docs/MONITORING_STACK.md)** - Complete monitoring stack with Grafana dashboards 

### Quick Links

- 🚀 **[Quick Start](#building-and-running)** - Get up and running quickly
- 🔧 **[Configuration](#configuration)** - Environment setup and API credentials
- 🐳 **[Docker Deployment](docs/DOCKER.md)** - Production Docker setup
- 📊 **[Monitoring Setup](docs/MONITORING_STACK.md)** - Full monitoring stack in one command

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run `go test ./...` to ensure tests pass
6. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
