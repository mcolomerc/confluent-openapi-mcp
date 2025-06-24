package prompts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// PromptManager handles loading and managing prompts from external files
type PromptManager struct {
	prompts       map[string]mcp.Prompt
	promptContent map[string]string // Store prompt content separately
	folder        string
}

// NewPromptManager creates a new prompt manager
// If folder is empty, it will default to "./prompts" relative to the executable
func NewPromptManager(folder string) *PromptManager {
	// If no folder provided, use default local prompts folder
	if folder == "" {
		// Get the executable directory
		execPath, err := os.Executable()
		if err == nil {
			execDir := filepath.Dir(execPath)
			folder = filepath.Join(execDir, "prompts")
		} else {
			// Fallback to current working directory + prompts
			folder = "prompts"
		}
	}

	return &PromptManager{
		prompts:       make(map[string]mcp.Prompt),
		promptContent: make(map[string]string),
		folder:        folder,
	}
}

// LoadPrompts loads all .txt files from the configured prompts folder
func (pm *PromptManager) LoadPrompts() error {
	if pm.folder == "" {
		// This shouldn't happen with the new default logic, but keep as safety
		return nil
	}

	// Check if folder exists
	if _, err := os.Stat(pm.folder); os.IsNotExist(err) {
		// Folder doesn't exist, try to find prompts relative to current working directory
		cwd, err := os.Getwd()
		if err == nil {
			altFolder := filepath.Join(cwd, "prompts")
			if _, err := os.Stat(altFolder); err == nil {
				pm.folder = altFolder
			} else {
				// No prompts folder found, return empty list (not an error)
				return nil
			}
		} else {
			// No prompts folder found, return empty list (not an error)
			return nil
		}
	}

	// Read all .txt files in the folder
	files, err := filepath.Glob(filepath.Join(pm.folder, "*.txt"))
	if err != nil {
		return fmt.Errorf("failed to read prompts folder: %w", err)
	}

	// Load each prompt file
	for _, file := range files {
		if err := pm.loadPromptFile(file); err != nil {
			return fmt.Errorf("failed to load prompt file %s: %w", file, err)
		}
	}

	return nil
}

// loadPromptFile loads a single prompt file
func (pm *PromptManager) loadPromptFile(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Extract prompt name from filename (without .txt extension)
	fileName := filepath.Base(filePath)
	promptName := strings.TrimSuffix(fileName, filepath.Ext(fileName))

	// Parse the content to extract description and prompt text
	description, promptText := parsePromptContent(string(content))

	// Create the prompt
	prompt := mcp.Prompt{
		Name:        promptName,
		Description: description,
		Arguments:   []mcp.PromptArgument{}, // No arguments for now, can be extended
	}

	// Store the prompt with its content
	pm.prompts[promptName] = prompt
	pm.promptContent[promptName] = promptText

	return nil
}

// parsePromptContent parses prompt file content
// Format: First line starting with # is description, rest is prompt content
func parsePromptContent(content string) (description, promptText string) {
	lines := strings.Split(content, "\n")

	// Look for description line (starts with #)
	var descriptionFound bool
	var promptLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !descriptionFound && strings.HasPrefix(trimmed, "#") {
			description = strings.TrimSpace(strings.TrimPrefix(trimmed, "#"))
			descriptionFound = true
		} else if trimmed != "" || descriptionFound {
			// Include non-empty lines, or all lines after description is found
			promptLines = append(promptLines, line)
		}
	}

	promptText = strings.Join(promptLines, "\n")

	// Default description if none found
	if description == "" {
		description = fmt.Sprintf("Prompt: %s", promptText[:min(50, len(promptText))])
		if len(promptText) > 50 {
			description += "..."
		}
	}

	return description, strings.TrimSpace(promptText)
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetPrompts returns all loaded prompts
func (pm *PromptManager) GetPrompts() []mcp.Prompt {
	prompts := make([]mcp.Prompt, 0, len(pm.prompts))
	for _, prompt := range pm.prompts {
		prompts = append(prompts, prompt)
	}
	return prompts
}

// GetPrompt returns a specific prompt by name
func (pm *PromptManager) GetPrompt(name string) (*mcp.Prompt, bool) {
	prompt, exists := pm.prompts[name]
	return &prompt, exists
}

// GetPromptContent returns the content of a specific prompt
func (pm *PromptManager) GetPromptContent(name string) (string, error) {
	content, exists := pm.promptContent[name]
	if !exists {
		return "", fmt.Errorf("prompt '%s' not found", name)
	}
	return content, nil
}

// ReloadPrompts reloads all prompts from the folder
func (pm *PromptManager) ReloadPrompts() error {
	// Clear existing prompts
	pm.prompts = make(map[string]mcp.Prompt)
	pm.promptContent = make(map[string]string)

	// Reload all prompts
	return pm.LoadPrompts()
}
