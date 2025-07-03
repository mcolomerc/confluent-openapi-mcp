package guardrails

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"mcolomerc/mcp-server/internal/logger"
	"sync"
	"time"
)

// LoopDetectionConfig holds configuration for loop detection
type LoopDetectionConfig struct {
	Enabled                bool
	MaxConsecutiveCalls    int
	TimeWindowSeconds      int
	CooldownSeconds        int
	EnableGlobalProtection bool
}

// ToolCall represents a single tool call with its parameters
type ToolCall struct {
	ToolName  string
	Args      map[string]interface{}
	Timestamp time.Time
	Hash      string
}

// LoopDetection provides protection against infinite loops in tool calls
type LoopDetection struct {
	config     LoopDetectionConfig
	callQueue  []ToolCall
	mu         sync.RWMutex
	cooldowns  map[string]time.Time // Hash -> cooldown end time
	cooldownMu sync.RWMutex
}

// LoopDetectionResult represents the result of loop detection
type LoopDetectionResult struct {
	IsLoop           bool
	ConsecutiveCalls int
	MaxAllowed       int
	CooldownUntil    *time.Time
	Message          string
}

// NewLoopDetection creates a new loop detection instance
func NewLoopDetection(config LoopDetectionConfig) *LoopDetection {
	if config.MaxConsecutiveCalls == 0 {
		config.MaxConsecutiveCalls = 3 // Default: max 3 consecutive identical calls
	}
	if config.TimeWindowSeconds == 0 {
		config.TimeWindowSeconds = 60 // Default: 1 minute time window
	}
	if config.CooldownSeconds == 0 {
		config.CooldownSeconds = 30 // Default: 30 second cooldown
	}

	return &LoopDetection{
		config:    config,
		callQueue: make([]ToolCall, 0),
		cooldowns: make(map[string]time.Time),
	}
}

// generateCallHash creates a hash for a tool call based on tool name and arguments
func (ld *LoopDetection) generateCallHash(toolName string, args map[string]interface{}) string {
	// Create a consistent representation of the call
	callData := map[string]interface{}{
		"tool": toolName,
		"args": args,
	}

	// Convert to JSON for consistent hashing
	jsonData, err := json.Marshal(callData)
	if err != nil {
		logger.Error("Failed to marshal call data for hashing: %v", err)
		// Fallback to simple string concatenation
		return fmt.Sprintf("%s-%v", toolName, args)
	}

	// Generate SHA256 hash
	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:])
}

// CheckForLoop checks if the current tool call would create a loop
func (ld *LoopDetection) CheckForLoop(toolName string, args map[string]interface{}) LoopDetectionResult {
	if !ld.config.Enabled {
		return LoopDetectionResult{IsLoop: false}
	}

	ld.mu.Lock()
	defer ld.mu.Unlock()

	now := time.Now()
	callHash := ld.generateCallHash(toolName, args)

	// Check if this call is in cooldown
	ld.cooldownMu.RLock()
	if cooldownEnd, exists := ld.cooldowns[callHash]; exists {
		if now.Before(cooldownEnd) {
			ld.cooldownMu.RUnlock()
			return LoopDetectionResult{
				IsLoop:        true,
				CooldownUntil: &cooldownEnd,
				Message:       fmt.Sprintf("Tool call is in cooldown until %s to prevent loops", cooldownEnd.Format("15:04:05")),
			}
		}
		// Cooldown expired, remove it
		ld.cooldownMu.RUnlock()
		ld.cooldownMu.Lock()
		delete(ld.cooldowns, callHash)
		ld.cooldownMu.Unlock()
	} else {
		ld.cooldownMu.RUnlock()
	}

	// Create current call
	currentCall := ToolCall{
		ToolName:  toolName,
		Args:      args,
		Timestamp: now,
		Hash:      callHash,
	}

	// Clean up old calls outside time window
	timeWindow := time.Duration(ld.config.TimeWindowSeconds) * time.Second
	var recentCalls []ToolCall
	for _, call := range ld.callQueue {
		if now.Sub(call.Timestamp) <= timeWindow {
			recentCalls = append(recentCalls, call)
		}
	}
	ld.callQueue = recentCalls

	// Count consecutive identical calls
	consecutiveCount := 1 // Current call counts as 1
	for i := len(ld.callQueue) - 1; i >= 0; i-- {
		if ld.callQueue[i].Hash == callHash {
			consecutiveCount++
		} else {
			break // Break on first non-matching call
		}
	}

	// Check if we've exceeded the limit
	if consecutiveCount > ld.config.MaxConsecutiveCalls {
		// Set cooldown
		cooldownEnd := now.Add(time.Duration(ld.config.CooldownSeconds) * time.Second)
		ld.cooldownMu.Lock()
		ld.cooldowns[callHash] = cooldownEnd
		ld.cooldownMu.Unlock()

		logger.Debug("Loop detected: %s called %d times consecutively (max: %d). Cooldown until %s",
			toolName, consecutiveCount, ld.config.MaxConsecutiveCalls, cooldownEnd.Format("15:04:05"))

		return LoopDetectionResult{
			IsLoop:           true,
			ConsecutiveCalls: consecutiveCount,
			MaxAllowed:       ld.config.MaxConsecutiveCalls,
			CooldownUntil:    &cooldownEnd,
			Message: fmt.Sprintf("Loop detected: %s called %d times consecutively (max: %d). Cooldown applied until %s",
				toolName, consecutiveCount, ld.config.MaxConsecutiveCalls, cooldownEnd.Format("15:04:05")),
		}
	}

	// Add current call to queue
	ld.callQueue = append(ld.callQueue, currentCall)

	// Log for monitoring
	if consecutiveCount > 1 {
		logger.Debug("Consecutive call detected: %s called %d times (max: %d)",
			toolName, consecutiveCount, ld.config.MaxConsecutiveCalls)
	}

	return LoopDetectionResult{
		IsLoop:           false,
		ConsecutiveCalls: consecutiveCount,
		MaxAllowed:       ld.config.MaxConsecutiveCalls,
	}
}

// GetCallHistory returns recent call history for debugging
func (ld *LoopDetection) GetCallHistory(limit int) []ToolCall {
	ld.mu.RLock()
	defer ld.mu.RUnlock()

	if limit <= 0 || limit > len(ld.callQueue) {
		limit = len(ld.callQueue)
	}

	// Return the last 'limit' calls
	start := len(ld.callQueue) - limit
	if start < 0 {
		start = 0
	}

	history := make([]ToolCall, limit)
	copy(history, ld.callQueue[start:])
	return history
}

// ClearCooldowns removes all cooldowns (for testing or manual intervention)
func (ld *LoopDetection) ClearCooldowns() {
	ld.cooldownMu.Lock()
	defer ld.cooldownMu.Unlock()
	ld.cooldowns = make(map[string]time.Time)
	logger.Debug("All loop detection cooldowns cleared")
}

// ClearCallHistory removes all call history (for testing or manual intervention)
func (ld *LoopDetection) ClearCallHistory() {
	ld.mu.Lock()
	defer ld.mu.Unlock()
	ld.callQueue = make([]ToolCall, 0)
	logger.Debug("All loop detection call history cleared")
}

// GetStats returns statistics about loop detection
func (ld *LoopDetection) GetStats() map[string]interface{} {
	ld.mu.RLock()
	ld.cooldownMu.RLock()
	defer ld.mu.RUnlock()
	defer ld.cooldownMu.RUnlock()

	stats := map[string]interface{}{
		"enabled":             ld.config.Enabled,
		"max_consecutive":     ld.config.MaxConsecutiveCalls,
		"time_window_seconds": ld.config.TimeWindowSeconds,
		"cooldown_seconds":    ld.config.CooldownSeconds,
		"recent_calls_count":  len(ld.callQueue),
		"active_cooldowns":    len(ld.cooldowns),
		"global_protection":   ld.config.EnableGlobalProtection,
	}

	return stats
}
