# 📊 MCP Server Monitoring Stack

Complete monitoring solution with Prometheus and Grafana for your MCP Server.

## 🚀 Quick Start

```bash
# Start the entire monitoring stack
./scripts/monitoring.sh start

# Check status
./scripts/monitoring.sh status

# View logs
./scripts/monitoring.sh logs

# Stop everything
./scripts/monitoring.sh stop
```

## 🎯 What You Get

### Services Running

- **MCP Server** (`:8080`) - Your server with monitoring enabled
- **Prometheus** (`:9090`) - Metrics collection
- **Grafana** (`:3000`) - Visualization dashboards
- **Node Exporter** (`:9100`) - System metrics

### 📈 Monitoring Endpoints

- `http://localhost:8080/health` - Health check
- `http://localhost:8080/metrics` - JSON metrics
- `http://localhost:8080/metrics/prometheus` - Prometheus format
- `http://localhost:8080/gc` - Force garbage collection

### 🔐 Access

- **Grafana**: <http://localhost:3000>
  - Username: `admin`
  - Password: `admin123`
- **Prometheus**: <http://localhost:9090>

## 📊 Available Metrics

| Metric | Description |
|--------|-------------|
| `mcp_memory_alloc_bytes` | Currently allocated memory |
| `mcp_memory_sys_bytes` | System memory from OS |
| `mcp_memory_heap_alloc_bytes` | Heap allocated memory |
| `mcp_goroutines_total` | Number of goroutines |
| `mcp_gc_runs_total` | Total garbage collections |
| `mcp_cpu_cores` | Number of CPU cores |

## 🎨 Pre-built Dashboard

The stack includes a ready-to-use Grafana dashboard with:

- Real-time memory usage
- Goroutine count tracking
- Garbage collection rate
- Historical trends

## 💡 Usage Examples

### Test Monitoring

```bash
# Health check
curl http://localhost:8080/health | jq

# Get all metrics in Prometheus format
curl http://localhost:8080/metrics/prometheus

# Trigger garbage collection
curl -X POST http://localhost:8080/gc | jq
```

### Prometheus Queries

```promql
# Memory usage in MB
mcp_memory_alloc_bytes / 1024 / 1024

# GC rate per minute
rate(mcp_gc_runs_total[5m]) * 60

# Memory efficiency
mcp_memory_alloc_bytes / mcp_memory_sys_bytes
```

Ready to monitor your MCP server! 🎉
