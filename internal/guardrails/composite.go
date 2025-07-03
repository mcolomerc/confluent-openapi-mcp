package guardrails

import (
	"mcolomerc/mcp-server/internal/config"
	"mcolomerc/mcp-server/internal/logger"
	"os"
	"strconv"
)

// CompositeGuardrails combines multiple guardrail mechanisms
type CompositeGuardrails struct {
	injectionDetector *InjectionDetection
	loopDetector      *LoopDetection
	enabled           bool
}

// GuardrailsResult represents the combined result of all guardrail checks
type GuardrailsResult struct {
	Blocked          bool
	InjectionResult  DetectionResult
	LoopResult       LoopDetectionResult
	BlockingReason   string
	AllowedToExecute bool
}

// NewCompositeGuardrails creates a new composite guardrails instance
func NewCompositeGuardrails(cfg *config.Config) *CompositeGuardrails {
	// Create injection detector
	injectionDetector := NewInjectionDetection()

	// Configure LLM detection if enabled
	if cfg.LLMDetectionEnabled {
		logger.Debug("Configuring LLM detection with URL: %s, Model: %s, Timeout: %ds\n",
			cfg.LLMDetectionURL, cfg.LLMDetectionModel, cfg.LLMDetectionTimeoutSec)

		llmConfig := ExternalLLMConfig{
			Enabled:    cfg.LLMDetectionEnabled,
			URL:        cfg.LLMDetectionURL,
			Model:      cfg.LLMDetectionModel,
			TimeoutSec: cfg.LLMDetectionTimeoutSec,
			APIKey:     cfg.LLMDetectionAPIKey,
		}
		injectionDetector.ConfigureLLM(llmConfig)
		logger.Debug("LLM detection configuration completed successfully\n")
	}

	// Create loop detection with configuration from environment
	loopConfig := LoopDetectionConfig{
		Enabled:                getEnvBool("LOOP_DETECTION_ENABLED", true),
		MaxConsecutiveCalls:    getEnvInt("LOOP_DETECTION_MAX_CONSECUTIVE", 3),
		TimeWindowSeconds:      getEnvInt("LOOP_DETECTION_TIME_WINDOW", 60),
		CooldownSeconds:        getEnvInt("LOOP_DETECTION_COOLDOWN", 30),
		EnableGlobalProtection: getEnvBool("LOOP_DETECTION_GLOBAL", true),
	}

	loopDetector := NewLoopDetection(loopConfig)

	logger.Debug("Loop detection configured: enabled=%v, max_consecutive=%d, time_window=%ds, cooldown=%ds",
		loopConfig.Enabled, loopConfig.MaxConsecutiveCalls, loopConfig.TimeWindowSeconds, loopConfig.CooldownSeconds)

	return &CompositeGuardrails{
		injectionDetector: injectionDetector,
		loopDetector:      loopDetector,
		enabled:           true,
	}
}

// ValidateToolInput validates tool parameters against all guardrails
func (cg *CompositeGuardrails) ValidateToolInput(toolName string, args map[string]interface{}) GuardrailsResult {
	result := GuardrailsResult{
		Blocked:          false,
		AllowedToExecute: true,
	}

	if !cg.enabled {
		return result
	}

	// 1. Check for injection attempts
	injectionResult := cg.injectionDetector.ValidateToolInput(toolName, args)
	result.InjectionResult = injectionResult

	if injectionResult.Detected {
		result.Blocked = true
		result.AllowedToExecute = false
		result.BlockingReason = "Prompt injection detected"
		if injectionResult.HighSeverity {
			result.BlockingReason = "High-risk prompt injection detected"
		}
		return result
	}

	// 2. Check for loop patterns
	loopResult := cg.loopDetector.CheckForLoop(toolName, args)
	result.LoopResult = loopResult

	if loopResult.IsLoop {
		result.Blocked = true
		result.AllowedToExecute = false
		result.BlockingReason = loopResult.Message
		return result
	}

	return result
}

// GetInjectionDetector returns the injection detector for direct access
func (cg *CompositeGuardrails) GetInjectionDetector() *InjectionDetection {
	return cg.injectionDetector
}

// GetLoopDetector returns the loop detector for direct access
func (cg *CompositeGuardrails) GetLoopDetector() *LoopDetection {
	return cg.loopDetector
}

// GetStats returns statistics about all guardrails
func (cg *CompositeGuardrails) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"enabled": cg.enabled,
		"injection_stats": map[string]interface{}{
			"enabled": cg.injectionDetector.enabled,
		},
		"loop_stats": cg.loopDetector.GetStats(),
	}
}

// ClearAllCooldowns clears all cooldowns (for testing or manual intervention)
func (cg *CompositeGuardrails) ClearAllCooldowns() {
	cg.loopDetector.ClearCooldowns()
	cg.loopDetector.ClearCallHistory()
	logger.Debug("All guardrail cooldowns and call history cleared")
}

// Enable or disable all guardrails
func (cg *CompositeGuardrails) SetEnabled(enabled bool) {
	cg.enabled = enabled
	logger.Debug("Guardrails enabled: %v", enabled)
}

// Helper functions for environment variable parsing
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
