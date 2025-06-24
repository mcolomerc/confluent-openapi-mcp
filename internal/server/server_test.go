package server

import (
	"mcolomerc/mcp-server/internal/config"
	"mcolomerc/mcp-server/internal/openapi"
	"mcolomerc/mcp-server/internal/tools"
	"os"
	"path/filepath"
	"testing"
)

func TestServerPromptsIntegration(t *testing.T) {
	// Create a temporary directory for test prompts
	tempDir := t.TempDir()

	// Create a test prompt file
	testContent := `# Server Test Prompt
This is a test prompt for server integration testing.
`

	testFile := filepath.Join(tempDir, "server-test.txt")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a test config with the prompts folder
	cfg := &config.Config{
		PromptsFolder: tempDir,
		// Add minimal required fields (you may need to adjust these)
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

	// Create a minimal OpenAPI spec
	spec := &openapi.OpenAPISpec{}

	// Create a server
	server := NewCompositeServer(cfg, spec, []tools.Tool{})

	// Test getting prompts
	prompts := server.GetPrompts()
	if len(prompts) != 1 {
		t.Fatalf("Expected 1 prompt, got %d", len(prompts))
	}

	// Test getting specific prompt
	prompt, exists := server.GetPrompt("server-test")
	if !exists {
		t.Fatal("Expected prompt to exist")
	}

	if prompt.Name != "server-test" {
		t.Errorf("Expected prompt name 'server-test', got '%s'", prompt.Name)
	}

	// Test getting prompt content
	content, err := server.GetPromptContent("server-test")
	if err != nil {
		t.Fatalf("Failed to get prompt content: %v", err)
	}

	expectedContent := "This is a test prompt for server integration testing."
	if content != expectedContent {
		t.Errorf("Expected content '%s', got '%s'", expectedContent, content)
	}
}

func TestServerPromptsNoFolder(t *testing.T) {
	// Create a test config without prompts folder
	cfg := &config.Config{
		// Add minimal required fields
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

	// Create a minimal OpenAPI spec
	spec := &openapi.OpenAPISpec{}

	// Create a server
	server := NewCompositeServer(cfg, spec, []tools.Tool{})

	// Test getting prompts - should return empty list
	prompts := server.GetPrompts()
	if len(prompts) != 0 {
		t.Errorf("Expected 0 prompts, got %d", len(prompts))
	}

	// Test getting non-existent prompt
	_, exists := server.GetPrompt("nonexistent")
	if exists {
		t.Error("Expected prompt to not exist")
	}
}
