# Documentation Index

This folder contains all the documentation for the Confluent OpenAPI MCP Server.

## 📚 Available Documentation

### Core Guides

- **[DEVELOPMENT.md](DEVELOPMENT.md)** - Development setup, debugging, and workflow
- **[DOCKER.md](DOCKER.md)** - Docker setup and deployment instructions

### Monitoring & Observability

- **[MONITORING.md](MONITORING.md)** - Basic monitoring setup and resource tracking
- **[PROMETHEUS.md](PROMETHEUS.md)** - Prometheus metrics export and configuration
- **[MONITORING_STACK.md](MONITORING_STACK.md)** - Complete monitoring stack with Grafana dashboards

## 🚀 Quick Start Links

- **Main README**: [../README.md](../README.md)
- **Building and Running**: [../README.md#building-and-running](../README.md#building-and-running)
- **Configuration**: [../README.md#configuration](../README.md#configuration)

## 📊 Monitoring Quick Start

For a complete monitoring setup with Grafana dashboards:

```bash
# Start the full monitoring stack
./scripts/monitoring.sh start

# Access Grafana at http://localhost:3000
# Username: admin, Password: admin123
```

See [MONITORING_STACK.md](MONITORING_STACK.md) for complete details.

## 🔧 Development Quick Start

For development with auto-reload:

```bash
# Install development tools
make install-tools

# Start development server with auto-reload
make dev
```

See [DEVELOPMENT.md](DEVELOPMENT.md) for complete development guide.
