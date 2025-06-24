package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"time"
)

// ResourceMetrics holds various system and runtime metrics
type ResourceMetrics struct {
	Memory     MemoryMetrics `json:"memory"`
	CPU        CPUMetrics    `json:"cpu"`
	Goroutines int           `json:"goroutines"`
	Timestamp  time.Time     `json:"timestamp"`
}

// MemoryMetrics holds memory-related metrics
type MemoryMetrics struct {
	AllocMB        float64 `json:"alloc_mb"`
	TotalAllocMB   float64 `json:"total_alloc_mb"`
	SysMB          float64 `json:"sys_mb"`
	NumGC          uint32  `json:"num_gc"`
	LastGC         string  `json:"last_gc"`
	HeapAllocMB    float64 `json:"heap_alloc_mb"`
	HeapSysMB      float64 `json:"heap_sys_mb"`
	HeapIdleMB     float64 `json:"heap_idle_mb"`
	HeapInuseMB    float64 `json:"heap_inuse_mb"`
	HeapReleasedMB float64 `json:"heap_released_mb"`
	StackInuseMB   float64 `json:"stack_inuse_mb"`
	StackSysMB     float64 `json:"stack_sys_mb"`
}

// CPUMetrics holds CPU-related metrics
type CPUMetrics struct {
	NumCPU     int   `json:"num_cpu"`
	NumCgoCall int64 `json:"num_cgo_call"`
}

// Monitor represents a resource monitor
type Monitor struct {
	interval time.Duration
	stopCh   chan struct{}
}

// NewMonitor creates a new resource monitor
func NewMonitor(interval time.Duration) *Monitor {
	return &Monitor{
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// GetCurrentMetrics returns current resource metrics
func (m *Monitor) GetCurrentMetrics() *ResourceMetrics {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	metrics := &ResourceMetrics{
		Memory: MemoryMetrics{
			AllocMB:        bytesToMB(memStats.Alloc),
			TotalAllocMB:   bytesToMB(memStats.TotalAlloc),
			SysMB:          bytesToMB(memStats.Sys),
			NumGC:          memStats.NumGC,
			LastGC:         time.Unix(0, int64(memStats.LastGC)).Format(time.RFC3339),
			HeapAllocMB:    bytesToMB(memStats.HeapAlloc),
			HeapSysMB:      bytesToMB(memStats.HeapSys),
			HeapIdleMB:     bytesToMB(memStats.HeapIdle),
			HeapInuseMB:    bytesToMB(memStats.HeapInuse),
			HeapReleasedMB: bytesToMB(memStats.HeapReleased),
			StackInuseMB:   bytesToMB(memStats.StackInuse),
			StackSysMB:     bytesToMB(memStats.StackSys),
		},
		CPU: CPUMetrics{
			NumCPU:     runtime.NumCPU(),
			NumCgoCall: runtime.NumCgoCall(),
		},
		Goroutines: runtime.NumGoroutine(),
		Timestamp:  time.Now(),
	}

	return metrics
}

// StartPeriodicLogging starts periodic logging of metrics
func (m *Monitor) StartPeriodicLogging(ctx context.Context) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopCh:
			return
		case <-ticker.C:
			metrics := m.GetCurrentMetrics()
			m.logMetrics(metrics)
		}
	}
}

// Stop stops the monitor
func (m *Monitor) Stop() {
	close(m.stopCh)
}

// logMetrics logs the metrics in a formatted way
func (m *Monitor) logMetrics(metrics *ResourceMetrics) {
	fmt.Printf("\n=== Resource Metrics [%s] ===\n", metrics.Timestamp.Format("15:04:05"))
	fmt.Printf("Memory Usage:\n")
	fmt.Printf("  Allocated: %.2f MB\n", metrics.Memory.AllocMB)
	fmt.Printf("  System:    %.2f MB\n", metrics.Memory.SysMB)
	fmt.Printf("  Heap:      %.2f MB (in use: %.2f MB)\n", metrics.Memory.HeapSysMB, metrics.Memory.HeapInuseMB)
	fmt.Printf("  Stack:     %.2f MB\n", metrics.Memory.StackInuseMB)
	fmt.Printf("  GC Runs:   %d\n", metrics.Memory.NumGC)
	fmt.Printf("CPU & Concurrency:\n")
	fmt.Printf("  CPUs:      %d\n", metrics.CPU.NumCPU)
	fmt.Printf("  Goroutines: %d\n", metrics.Goroutines)
	fmt.Printf("  CGO Calls: %d\n", metrics.CPU.NumCgoCall)
	fmt.Println("=======================================")
}

// GetMetricsJSON returns metrics as JSON string
func (m *Monitor) GetMetricsJSON() (string, error) {
	metrics := m.GetCurrentMetrics()
	jsonData, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// bytesToMB converts bytes to megabytes
func bytesToMB(bytes uint64) float64 {
	return float64(bytes) / 1024 / 1024
}

// ForceGC triggers garbage collection and returns metrics before and after
func (m *Monitor) ForceGC() (before, after *ResourceMetrics) {
	before = m.GetCurrentMetrics()
	runtime.GC()
	after = m.GetCurrentMetrics()
	return before, after
}
