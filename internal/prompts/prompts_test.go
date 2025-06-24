package prompts

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPromptManager(t *testing.T) {
	// Create a temporary directory for test prompts
	tempDir := t.TempDir()

	// Create a test prompt file
	testContent := `# Test Prompt
This is a test prompt for testing purposes.

Please help with testing:
1. Check functionality
2. Verify parsing
3. Ensure content loads correctly
`

	testFile := filepath.Join(tempDir, "test-prompt.txt")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create prompt manager
	pm := NewPromptManager(tempDir)

	// Load prompts
	if err := pm.LoadPrompts(); err != nil {
		t.Fatalf("Failed to load prompts: %v", err)
	}

	// Test getting prompts list
	prompts := pm.GetPrompts()
	if len(prompts) != 1 {
		t.Fatalf("Expected 1 prompt, got %d", len(prompts))
	}

	// Test prompt details
	prompt := prompts[0]
	if prompt.Name != "test-prompt" {
		t.Errorf("Expected prompt name 'test-prompt', got '%s'", prompt.Name)
	}

	if prompt.Description != "Test Prompt" {
		t.Errorf("Expected description 'Test Prompt', got '%s'", prompt.Description)
	}

	// Test getting specific prompt
	foundPrompt, exists := pm.GetPrompt("test-prompt")
	if !exists {
		t.Error("Expected prompt to exist")
	}

	if foundPrompt.Name != "test-prompt" {
		t.Errorf("Expected found prompt name 'test-prompt', got '%s'", foundPrompt.Name)
	}

	// Test getting prompt content
	content, err := pm.GetPromptContent("test-prompt")
	if err != nil {
		t.Fatalf("Failed to get prompt content: %v", err)
	}

	expectedContent := `This is a test prompt for testing purposes.

Please help with testing:
1. Check functionality
2. Verify parsing
3. Ensure content loads correctly`

	if content != expectedContent {
		t.Errorf("Content mismatch.\nExpected:\n%s\n\nGot:\n%s", expectedContent, content)
	}
}

func TestPromptManagerEmptyFolder(t *testing.T) {
	// Test with a nonexistent default folder path to simulate empty folder behavior
	pm := NewPromptManager("/definitely/nonexistent/path/that/should/not/exist")

	// Should not error on empty folder
	if err := pm.LoadPrompts(); err != nil {
		t.Errorf("Expected no error for empty folder, got: %v", err)
	}

	// Should return empty list
	prompts := pm.GetPrompts()
	if len(prompts) != 0 {
		t.Errorf("Expected 0 prompts for empty folder, got %d", len(prompts))
	}
}

func TestPromptManagerNonexistentFolder(t *testing.T) {
	pm := NewPromptManager("/nonexistent/folder")

	// Should not error on nonexistent folder
	if err := pm.LoadPrompts(); err != nil {
		t.Errorf("Expected no error for nonexistent folder, got: %v", err)
	}

	// Should return empty list
	prompts := pm.GetPrompts()
	if len(prompts) != 0 {
		t.Errorf("Expected 0 prompts for nonexistent folder, got %d", len(prompts))
	}
}
