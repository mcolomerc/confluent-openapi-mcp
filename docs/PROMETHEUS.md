# Prometheus Integration Guide

## 🎯 Prometheus Export

MCP server supports Prometheus metrics with **zero external dependencies**!

### 📊 Available Endpoints

```bash
# JSON format (original)
curl http://localhost:8080/metrics

# Prometheus format (new!)
curl http://localhost:8080/metrics/prometheus

# Health check
curl http://localhost:8080/health
```

### 🚀 Quick Start

1. **Start your server:**

```bash
./bin/mcp-server -mode=http -monitor=30s
```

1. **Test Prometheus endpoint:**

```bash
curl http://localhost:8080/metrics/prometheus
```

**Output example:**

```shell
# HELP mcp_memory_alloc_bytes Currently allocated memory in bytes
# TYPE mcp_memory_alloc_bytes gauge
mcp_memory_alloc_bytes 12582912

# HELP mcp_goroutines_total Number of goroutines
# TYPE mcp_goroutines_total gauge
mcp_goroutines_total 15

# HELP mcp_gc_runs_total Total number of garbage collections
# TYPE mcp_gc_runs_total counter
mcp_gc_runs_total 3
```

### 🔧 Prometheus Configuration

Add this to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'mcp-server'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics/prometheus'
    scrape_interval: 30s
```

### 📈 Available Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `mcp_memory_alloc_bytes` | gauge | Currently allocated memory |
| `mcp_memory_sys_bytes` | gauge | System memory from OS |
| `mcp_memory_heap_alloc_bytes` | gauge | Heap allocated memory |
| `mcp_memory_heap_sys_bytes` | gauge | Heap system memory |
| `mcp_memory_stack_inuse_bytes` | gauge | Stack memory in use |
| `mcp_goroutines_total` | gauge | Number of goroutines |
| `mcp_gc_runs_total` | counter | Total garbage collections |
| `mcp_cpu_cores` | gauge | Number of CPU cores |
| `mcp_cgo_calls_total` | counter | Total CGO calls |

### 🎨 Grafana Dashboard

Import these queries for instant dashboards:

```promql
# Memory usage over time
mcp_memory_alloc_bytes

# Goroutine count
mcp_goroutines_total

# GC frequency
rate(mcp_gc_runs_total[5m])

# Memory efficiency (allocated vs system)
mcp_memory_alloc_bytes / mcp_memory_sys_bytes
```

### 🔄 Integration Examples

**Docker Compose with Prometheus:**

```yaml
version: '3'
services:
  mcp-server:
    build: .
    ports:
      - "8080:8080"
    command: ["./bin/mcp-server", "-mode=http", "-monitor=30s"]
  
  prometheus:
    image: prom/prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
```

**Kubernetes Service Monitor:**

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: mcp-server
spec:
  selector:
    matchLabels:
      app: mcp-server
  endpoints:
  - port: http
    path: /metrics/prometheus
```
