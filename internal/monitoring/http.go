package monitoring

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// HTTPHandler provides HTTP endpoints for metrics
type HTTPHandler struct {
	monitor *Monitor
}

// NewHTTPHandler creates a new HTTP handler for metrics
func NewHTTPHandler(monitor *Monitor) *HTTPHandler {
	return &HTTPHandler{monitor: monitor}
}

// MetricsHandler serves current metrics as JSON
func (h *HTTPHandler) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	metrics := h.monitor.GetCurrentMetrics()
	jsonData, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		http.Error(w, "Failed to marshal metrics", http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
}

// HealthHandler serves a simple health check with basic metrics
func (h *HTTPHandler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	metrics := h.monitor.GetCurrentMetrics()
	health := map[string]interface{}{
		"status":     "healthy",
		"memory_mb":  metrics.Memory.AllocMB,
		"goroutines": metrics.Goroutines,
		"timestamp":  metrics.Timestamp,
	}

	json.NewEncoder(w).Encode(health)
}

// GCHandler triggers garbage collection and returns before/after metrics
func (h *HTTPHandler) GCHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	before, after := h.monitor.ForceGC()
	result := map[string]*ResourceMetrics{
		"before_gc": before,
		"after_gc":  after,
	}

	json.NewEncoder(w).Encode(result)
}

// PrometheusHandler serves metrics in Prometheus text format
func (h *HTTPHandler) PrometheusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	metrics := h.monitor.GetCurrentMetrics()

	// Write Prometheus format metrics
	fmt.Fprintf(w, "# HELP mcp_memory_alloc_bytes Currently allocated memory in bytes\n")
	fmt.Fprintf(w, "# TYPE mcp_memory_alloc_bytes gauge\n")
	fmt.Fprintf(w, "mcp_memory_alloc_bytes %.0f\n", metrics.Memory.AllocMB*1024*1024)

	fmt.Fprintf(w, "# HELP mcp_memory_sys_bytes System memory in bytes\n")
	fmt.Fprintf(w, "# TYPE mcp_memory_sys_bytes gauge\n")
	fmt.Fprintf(w, "mcp_memory_sys_bytes %.0f\n", metrics.Memory.SysMB*1024*1024)

	fmt.Fprintf(w, "# HELP mcp_memory_heap_alloc_bytes Heap allocated memory in bytes\n")
	fmt.Fprintf(w, "# TYPE mcp_memory_heap_alloc_bytes gauge\n")
	fmt.Fprintf(w, "mcp_memory_heap_alloc_bytes %.0f\n", metrics.Memory.HeapAllocMB*1024*1024)

	fmt.Fprintf(w, "# HELP mcp_memory_heap_sys_bytes Heap system memory in bytes\n")
	fmt.Fprintf(w, "# TYPE mcp_memory_heap_sys_bytes gauge\n")
	fmt.Fprintf(w, "mcp_memory_heap_sys_bytes %.0f\n", metrics.Memory.HeapSysMB*1024*1024)

	fmt.Fprintf(w, "# HELP mcp_memory_stack_inuse_bytes Stack memory in use in bytes\n")
	fmt.Fprintf(w, "# TYPE mcp_memory_stack_inuse_bytes gauge\n")
	fmt.Fprintf(w, "mcp_memory_stack_inuse_bytes %.0f\n", metrics.Memory.StackInuseMB*1024*1024)

	fmt.Fprintf(w, "# HELP mcp_goroutines_total Number of goroutines\n")
	fmt.Fprintf(w, "# TYPE mcp_goroutines_total gauge\n")
	fmt.Fprintf(w, "mcp_goroutines_total %d\n", metrics.Goroutines)

	fmt.Fprintf(w, "# HELP mcp_gc_runs_total Total number of garbage collections\n")
	fmt.Fprintf(w, "# TYPE mcp_gc_runs_total counter\n")
	fmt.Fprintf(w, "mcp_gc_runs_total %d\n", metrics.Memory.NumGC)

	fmt.Fprintf(w, "# HELP mcp_cpu_cores Number of CPU cores\n")
	fmt.Fprintf(w, "# TYPE mcp_cpu_cores gauge\n")
	fmt.Fprintf(w, "mcp_cpu_cores %d\n", metrics.CPU.NumCPU)

	fmt.Fprintf(w, "# HELP mcp_cgo_calls_total Total number of CGO calls\n")
	fmt.Fprintf(w, "# TYPE mcp_cgo_calls_total counter\n")
	fmt.Fprintf(w, "mcp_cgo_calls_total %d\n", metrics.CPU.NumCgoCall)
}
