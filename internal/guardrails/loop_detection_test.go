package guardrails

import (
	"testing"
	"time"
)

func TestLoopDetection(t *testing.T) {
	config := LoopDetectionConfig{
		Enabled:                true,
		MaxConsecutiveCalls:    3,
		TimeWindowSeconds:      60,
		CooldownSeconds:        30,
		EnableGlobalProtection: true,
	}

	detector := NewLoopDetection(config)

	// Test normal calls - should not trigger loop detection
	t.Run("Normal calls", func(t *testing.T) {
		result1 := detector.CheckForLoop("list", map[string]interface{}{"resource": "environments"})
		if result1.IsLoop {
			t.Error("First call should not be detected as loop")
		}

		result2 := detector.CheckForLoop("list", map[string]interface{}{"resource": "clusters"})
		if result2.IsLoop {
			t.Error("Different tool call should not be detected as loop")
		}
	})

	// Test consecutive identical calls - should trigger loop detection
	t.Run("Consecutive identical calls", func(t *testing.T) {
		detector := NewLoopDetection(config) // Fresh instance

		args := map[string]interface{}{"resource": "costs", "start_date": "2025-06-01"}

		// First 3 calls should be allowed
		for i := 1; i <= 3; i++ {
			result := detector.CheckForLoop("list", args)
			if result.IsLoop {
				t.Errorf("Call %d should not be detected as loop", i)
			}
			if result.ConsecutiveCalls != i {
				t.Errorf("Expected consecutive calls to be %d, got %d", i, result.ConsecutiveCalls)
			}
		}

		// 4th call should trigger loop detection
		result := detector.CheckForLoop("list", args)
		if !result.IsLoop {
			t.Error("4th consecutive call should be detected as loop")
		}
		if result.ConsecutiveCalls != 4 {
			t.Errorf("Expected consecutive calls to be 4, got %d", result.ConsecutiveCalls)
		}
		if result.CooldownUntil == nil {
			t.Error("Loop detection should set cooldown")
		}
	})

	// Test cooldown behavior
	t.Run("Cooldown behavior", func(t *testing.T) {
		detector := NewLoopDetection(config) // Fresh instance

		args := map[string]interface{}{"resource": "costs", "start_date": "2025-06-01"}

		// Trigger loop detection
		for i := 1; i <= 4; i++ {
			detector.CheckForLoop("list", args)
		}

		// Immediate retry should be blocked
		result := detector.CheckForLoop("list", args)
		if !result.IsLoop {
			t.Error("Call during cooldown should be blocked")
		}
	})

	// Test different parameters don't trigger loop detection
	t.Run("Different parameters", func(t *testing.T) {
		detector := NewLoopDetection(config) // Fresh instance

		// Same tool, different parameters
		result1 := detector.CheckForLoop("list", map[string]interface{}{"resource": "costs", "start_date": "2025-06-01"})
		result2 := detector.CheckForLoop("list", map[string]interface{}{"resource": "costs", "start_date": "2025-05-01"})
		result3 := detector.CheckForLoop("list", map[string]interface{}{"resource": "environments"})

		if result1.IsLoop || result2.IsLoop || result3.IsLoop {
			t.Error("Different parameters should not trigger loop detection")
		}
	})

	// Test time window cleanup
	t.Run("Time window cleanup", func(t *testing.T) {
		shortConfig := LoopDetectionConfig{
			Enabled:                true,
			MaxConsecutiveCalls:    3,
			TimeWindowSeconds:      1, // Very short window
			CooldownSeconds:        30,
			EnableGlobalProtection: true,
		}

		detector := NewLoopDetection(shortConfig)
		args := map[string]interface{}{"resource": "costs"}

		// Make 2 calls
		detector.CheckForLoop("list", args)
		detector.CheckForLoop("list", args)

		// Wait for time window to expire
		time.Sleep(2 * time.Second)

		// Next call should start fresh count
		result := detector.CheckForLoop("list", args)
		if result.IsLoop {
			t.Error("Call after time window should not be detected as loop")
		}
		if result.ConsecutiveCalls != 1 {
			t.Errorf("Expected consecutive calls to be 1 after time window, got %d", result.ConsecutiveCalls)
		}
	})
}

func TestLoopDetectionDisabled(t *testing.T) {
	config := LoopDetectionConfig{
		Enabled: false,
	}

	detector := NewLoopDetection(config)
	args := map[string]interface{}{"resource": "costs"}

	// Even many consecutive calls should not trigger loop detection when disabled
	for i := 0; i < 10; i++ {
		result := detector.CheckForLoop("list", args)
		if result.IsLoop {
			t.Error("Loop detection should be disabled")
		}
	}
}

func TestHashGeneration(t *testing.T) {
	detector := NewLoopDetection(LoopDetectionConfig{Enabled: true})

	// Same parameters should generate same hash
	args1 := map[string]interface{}{"resource": "costs", "start_date": "2025-06-01"}
	args2 := map[string]interface{}{"resource": "costs", "start_date": "2025-06-01"}
	hash1 := detector.generateCallHash("list", args1)
	hash2 := detector.generateCallHash("list", args2)

	if hash1 != hash2 {
		t.Error("Same parameters should generate same hash")
	}

	// Different parameters should generate different hash
	args3 := map[string]interface{}{"resource": "costs", "start_date": "2025-05-01"}
	hash3 := detector.generateCallHash("list", args3)

	if hash1 == hash3 {
		t.Error("Different parameters should generate different hash")
	}
}
