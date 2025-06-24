package server

import (
	"mcolomerc/mcp-server/internal/config"
	"mcolomerc/mcp-server/internal/openapi"
	"mcolomerc/mcp-server/internal/tools"
	"os"
	"path/filepath"
	"testing"
)

func TestPromptsFunctionality(t *testing.T) {
	// Create a temporary directory for test prompts
	tempDir := t.TempDir()

	// Create test prompt files
	testPrompts := map[string]string{
		"test-prompt-1.txt": `# Test Prompt 1
This is the first test prompt.`,
		"test-prompt-2.txt": `# Test Prompt 2
This is the second test prompt with more content.

It has multiple lines.`,
	}

	for filename, content := range testPrompts {
		testFile := filepath.Join(tempDir, filename)
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Create a test config with the prompts folder
	cfg := &config.Config{
		PromptsFolder:           tempDir,
		OpenAPISpecURL:          "http://test.com/spec.json",
		ConfluentEnvID:          "env-test",
		ConfluentCloudAPIKey:    "test-key",
		ConfluentCloudAPISecret: "test-secret",
		BootstrapServers:        "test-servers",
		KafkaAPIKey:             "test-key",
		KafkaAPISecret:          "test-secret",
		KafkaRestEndpoint:       "http://test.com",
		KafkaClusterID:          "lkc-test",
		FlinkOrgID:              "test-org",
		FlinkRestEndpoint:       "http://test.com",
		FlinkEnvName:            "test",
		FlinkDatabaseName:       "test",
		FlinkAPIKey:             "test-key",
		FlinkAPISecret:          "test-secret",
		FlinkComputePoolID:      "lfcp-test",
		SchemaRegistryAPIKey:    "test-key",
		SchemaRegistryAPISecret: "test-secret",
		SchemaRegistryEndpoint:  "http://test.com",
	}

	// Create a server
	spec := &openapi.OpenAPISpec{}
	server := NewCompositeServer(cfg, spec, []tools.Tool{})

	t.Run("GetPrompts returns all prompts", func(t *testing.T) {
		prompts := server.GetPrompts()

		if len(prompts) != 2 {
			t.Errorf("Expected 2 prompts, got %d", len(prompts))
		}

		// Check that we have the expected prompts
		promptNames := make(map[string]bool)
		for _, prompt := range prompts {
			promptNames[prompt.Name] = true
		}

		if !promptNames["test-prompt-1"] {
			t.Error("Expected to find test-prompt-1")
		}
		if !promptNames["test-prompt-2"] {
			t.Error("Expected to find test-prompt-2")
		}
	})

	t.Run("GetPrompt returns specific prompt", func(t *testing.T) {
		prompt, found := server.GetPrompt("test-prompt-1")
		if !found {
			t.Fatal("Expected to find test-prompt-1")
		}

		if prompt.Name != "test-prompt-1" {
			t.Errorf("Expected name 'test-prompt-1', got '%s'", prompt.Name)
		}

		if prompt.Description != "Test Prompt 1" {
			t.Errorf("Expected description 'Test Prompt 1', got '%s'", prompt.Description)
		}
	})

	t.Run("GetPromptContent returns content", func(t *testing.T) {
		content, err := server.GetPromptContent("test-prompt-1")
		if err != nil {
			t.Fatalf("Failed to get prompt content: %v", err)
		}

		expected := "This is the first test prompt."
		if content != expected {
			t.Errorf("Expected content '%s', got '%s'", expected, content)
		}
	})

	t.Run("GetPrompt returns false for nonexistent prompt", func(t *testing.T) {
		_, found := server.GetPrompt("nonexistent")
		if found {
			t.Error("Expected not to find nonexistent prompt")
		}
	})

	t.Run("GetPromptContent returns error for nonexistent prompt", func(t *testing.T) {
		_, err := server.GetPromptContent("nonexistent")
		if err == nil {
			t.Error("Expected error for nonexistent prompt")
		}
	})
}
