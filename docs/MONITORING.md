# Resource Monitoring Guide

This document explains how to monitor CPU, memory, and other resource usage for the MCP server.

## Overview

The MCP server includes built-in resource monitoring capabilities that provide:

- **Real-time metrics** via HTTP endpoints
- **Periodic logging** to console
- **Memory management** tools (garbage collection)
- **System resource** tracking (CPU, memory, goroutines)

## Quick Start

### 1. Start Server with Monitoring

```bash
# Monitor every 30 seconds (default)
./bin/mcp-server -mode=http

# Monitor every 10 seconds
./bin/mcp-server -mode=http -monitor=10s

# Monitor every 1 minute
./bin/mcp-server -mode=http -monitor=1m

# Disable monitoring
./bin/mcp-server -mode=http -monitor=off
```

### 2. Access Metrics via HTTP

```bash
# Health check with basic metrics
curl http://localhost:8080/health

# Comprehensive metrics
curl http://localhost:8080/metrics

# Force garbage collection
curl -X POST http://localhost:8080/gc
```

## Monitoring Modes

### Console Monitoring

When enabled, metrics are logged to stderr at regular intervals:

```shell
=== Resource Metrics [14:30:15] ===
Memory Usage:
  Allocated: 12.34 MB
  System:    45.67 MB
  Heap:      32.10 MB (in use: 12.34 MB)
  Stack:     1.23 MB
  GC Runs:   5
CPU & Concurrency:
  CPUs:      8
  Goroutines: 15
  CGO Calls: 0
=======================================
```

### HTTP Endpoints

#### `/health` - Health Check

Returns basic health status with key metrics:

```json
{
  "status": "healthy",
  "memory_mb": 12.34,
  "goroutines": 15,
  "timestamp": "2025-06-24T14:30:15Z"
}
```

#### `/metrics` - Comprehensive Metrics

Returns detailed resource information:

```json
{
  "memory": {
    "alloc_mb": 12.34,
    "total_alloc_mb": 123.45,
    "sys_mb": 45.67,
    "num_gc": 5,
    "last_gc": "2025-06-24T14:30:10Z",
    "heap_alloc_mb": 12.34,
    "heap_sys_mb": 32.10,
    "heap_idle_mb": 19.76,
    "heap_inuse_mb": 12.34,
    "heap_released_mb": 15.00,
    "stack_inuse_mb": 1.23,
    "stack_sys_mb": 1.50
  },
  "cpu": {
    "num_cpu": 8,
    "num_cgo_call": 0
  },
  "goroutines": 15,
  "timestamp": "2025-06-24T14:30:15Z"
}
```

#### `/gc` - Garbage Collection

Triggers garbage collection and shows before/after comparison:

```json
{
  "before_gc": {
    "memory": { "alloc_mb": 20.45, ... },
    ...
  },
  "after_gc": {
    "memory": { "alloc_mb": 12.34, ... },
    ...
  }
}
```

## Metric Definitions

### Memory Metrics

- **alloc_mb**: Currently allocated memory
- **total_alloc_mb**: Cumulative memory allocated (includes freed memory)
- **sys_mb**: Total memory obtained from OS
- **heap_alloc_mb**: Currently allocated heap memory
- **heap_sys_mb**: Total heap memory from OS
- **heap_idle_mb**: Idle heap memory
- **heap_inuse_mb**: In-use heap memory
- **heap_released_mb**: Released heap memory back to OS
- **stack_inuse_mb**: In-use stack memory
- **stack_sys_mb**: Total stack memory from OS
- **num_gc**: Number of garbage collection cycles
- **last_gc**: Timestamp of last garbage collection

### CPU & Concurrency Metrics

- **num_cpu**: Number of logical CPUs
- **num_cgo_call**: Number of CGO calls made
- **goroutines**: Current number of goroutines

## Integration Examples

### Monitoring with curl

```bash
# Watch metrics every 5 seconds
watch -n 5 'curl -s http://localhost:8080/metrics | jq ".memory.alloc_mb"'

# Monitor health status
while true; do
  curl -s http://localhost:8080/health | jq '.memory_mb'
  sleep 10
done
```

### Integration with Prometheus

You can easily integrate with Prometheus by creating a metrics exporter:

```bash
# Simple metrics exporter
curl -s http://localhost:8080/metrics | jq -r '
  "mcp_memory_allocated_mb \(.memory.alloc_mb)",
  "mcp_goroutines \(.goroutines)",
  "mcp_gc_count \(.memory.num_gc)"
'
```

### Docker Monitoring

When running in Docker, you can expose the metrics port:

```bash
docker run -p 8080:8080 mcp-server -mode=http -monitor=30s
```

## Performance Impact

The monitoring system has minimal performance impact:

- **Console logging**: Very low overhead, only affects stderr output
- **HTTP endpoints**: Only consume resources when accessed
- **Memory overhead**: ~1-2MB for monitoring data structures
- **CPU overhead**: <1% for periodic metrics collection

## Troubleshooting

### High Memory Usage

1. Check metrics to identify memory patterns:

   ```bash
   curl http://localhost:8080/metrics | jq '.memory'
   ```

2. Force garbage collection to see if memory is reclaimable:

   ```bash
   curl -X POST http://localhost:8080/gc
   ```

3. Monitor heap usage over time to detect memory leaks

### High CPU Usage

1. Check goroutine count for potential goroutine leaks:

   ```bash
   curl http://localhost:8080/metrics | jq '.goroutines'
   ```

2. Monitor CGO calls if using C libraries:

   ```bash
   curl http://localhost:8080/metrics | jq '.cpu.num_cgo_call'
   ```

### Monitoring Not Working

1. Ensure server is running in HTTP mode:

   ```bash
   ./bin/mcp-server -mode=http
   ```

2. Check if monitoring is enabled:

   ```bash
   curl http://localhost:8080/health
   ```

3. Verify correct port and address:

   ```bash
   netstat -an | grep 8080
   ```

## Command Line Options

```bash
./bin/mcp-server [options]

Options:
  -mode string
        Server mode: 'stdio', 'http', or 'both' (default "both")
  -monitor string
        Resource monitoring interval (e.g., 30s, 1m, 5m). Set to 'off' to disable (default "30s")
  -env string
        Path to environment file (default ".env")
```

## Demo Script

Use the provided demo script to explore monitoring features:

```bash
# Show all monitoring information
./scripts/monitor-demo.sh

# Test endpoints (server must be running)
./scripts/monitor-demo.sh test

# Show available endpoints
./scripts/monitor-demo.sh endpoints

# Show monitoring tips
./scripts/monitor-demo.sh tips
```
