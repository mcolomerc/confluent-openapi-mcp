package prompts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"mcolomerc/mcp-server/internal/config"

	"github.com/mark3labs/mcp-go/mcp"
)

// PromptManager handles loading and managing prompts from external files
type PromptManager struct {
	prompts          map[string]mcp.Prompt
	promptContent    map[string]string // Store prompt content separately
	folder           string
	config           *config.Config // Add config for variable substitution
	directives       string         // Combined directives content
	directivesFolder string         // Path to directives folder
}

// NewPromptManager creates a new prompt manager
// If folder is empty, it will default to "./prompts" relative to the executable
func NewPromptManager(folder string, cfg *config.Config) *PromptManager {
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

	// Determine directives folder path
	directivesFolder := ""
	if cfg != nil && cfg.DirectivesFolder != "" {
		// Use config-specified directives folder
		directivesFolder = cfg.DirectivesFolder
	} else {
		// Try to find directives folder relative to executable
		execPath, err := os.Executable()
		if err == nil {
			execDir := filepath.Dir(execPath)
			directivesFolder = filepath.Join(execDir, "directives")
		} else {
			// Fallback to current working directory + directives
			directivesFolder = "directives"
		}
	}

	return &PromptManager{
		prompts:          make(map[string]mcp.Prompt),
		promptContent:    make(map[string]string),
		folder:           folder,
		config:           cfg,
		directivesFolder: directivesFolder,
	}
}

// LoadPrompts loads all .txt files from the configured prompts folder
func (pm *PromptManager) LoadPrompts() error {
	// First, load directives
	if err := pm.loadDirectives(); err != nil {
		return fmt.Errorf("failed to load directives: %w", err)
	}

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

	// Store the original prompt content without substitution for potential argument-based substitution later
	pm.promptContent[promptName] = promptText

	// Perform variable substitution only for default content retrieval
	promptText, err = pm.substituteVariables(promptText)
	if err != nil {
		return fmt.Errorf("failed to substitute variables in prompt text: %w", err)
	}

	// Define common arguments that can override default config values
	arguments := []mcp.PromptArgument{}

	// Check for environment ID references (both formats)
	if strings.Contains(promptText, "CONFLUENT_ENV_ID") || strings.Contains(pm.promptContent[promptName], "{CONFLUENT_ENV_ID}") ||
		strings.Contains(promptText, "environment") || strings.Contains(pm.promptContent[promptName], "{environment}") ||
		strings.Contains(promptText, "environment_id") || strings.Contains(pm.promptContent[promptName], "{environment_id}") {
		arguments = append(arguments, mcp.PromptArgument{
			Name:        "environment_id",
			Description: "Override the default Confluent environment ID",
			Required:    false,
		})
	}

	// Check for cluster ID references (both formats)
	if strings.Contains(promptText, "KAFKA_CLUSTER_ID") || strings.Contains(pm.promptContent[promptName], "{KAFKA_CLUSTER_ID}") ||
		strings.Contains(promptText, "cluster_id") || strings.Contains(pm.promptContent[promptName], "{cluster_id}") ||
		strings.Contains(promptText, "kafka_cluster_id") || strings.Contains(pm.promptContent[promptName], "{kafka_cluster_id}") {
		arguments = append(arguments, mcp.PromptArgument{
			Name:        "cluster_id",
			Description: "Override the default Kafka cluster ID",
			Required:    false,
		})
	}

	// Check for compute pool ID references (both formats)
	if strings.Contains(promptText, "FLINK_COMPUTE_POOL_ID") || strings.Contains(pm.promptContent[promptName], "{FLINK_COMPUTE_POOL_ID}") ||
		strings.Contains(promptText, "compute_pool_id") || strings.Contains(pm.promptContent[promptName], "{compute_pool_id}") ||
		strings.Contains(promptText, "pool_id") || strings.Contains(pm.promptContent[promptName], "{pool_id}") {
		arguments = append(arguments, mcp.PromptArgument{
			Name:        "compute_pool_id",
			Description: "Override the default Flink compute pool ID",
			Required:    false,
		})
	}

	// Check for organization ID references (both formats)
	if strings.Contains(promptText, "FLINK_ORG_ID") || strings.Contains(pm.promptContent[promptName], "{FLINK_ORG_ID}") ||
		strings.Contains(promptText, "organization_id") || strings.Contains(pm.promptContent[promptName], "{organization_id}") ||
		strings.Contains(promptText, "org_id") || strings.Contains(pm.promptContent[promptName], "{org_id}") ||
		strings.Contains(promptText, "org") || strings.Contains(pm.promptContent[promptName], "{org}") {
		arguments = append(arguments, mcp.PromptArgument{
			Name:        "organization_id",
			Description: "Override the default Flink organization ID",
			Required:    false,
		})
	}

	// Create the prompt
	prompt := mcp.Prompt{
		Name:        promptName,
		Description: description,
		Arguments:   arguments,
	}

	// Store the prompt with its content
	pm.prompts[promptName] = prompt

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

// ReloadPrompts reloads all prompts and directives from their respective folders
func (pm *PromptManager) ReloadPrompts() error {
	// Clear existing prompts and directives
	pm.prompts = make(map[string]mcp.Prompt)
	pm.promptContent = make(map[string]string)
	pm.directives = ""

	// Reload all prompts (which will also reload directives)
	return pm.LoadPrompts()
}

// substituteVariables performs variable substitution in the prompt text
func (pm *PromptManager) substituteVariables(promptText string) (string, error) {
	if pm.config == nil {
		return promptText, nil
	}

	// Define mappings from placeholder names to config values
	// Support both environment variable format and parameter format
	replacements := map[string]string{
		// Environment variable format (original)
		"{CONFLUENT_ENV_ID}":           pm.config.ConfluentEnvID,
		"{KAFKA_CLUSTER_ID}":           pm.config.KafkaClusterID,
		"{CONFLUENT_CLOUD_API_KEY}":    pm.config.ConfluentCloudAPIKey,
		"{CONFLUENT_CLOUD_API_SECRET}": pm.config.ConfluentCloudAPISecret,
		"{BOOTSTRAP_SERVERS}":          pm.config.BootstrapServers,
		"{KAFKA_API_KEY}":              pm.config.KafkaAPIKey,
		"{KAFKA_API_SECRET}":           pm.config.KafkaAPISecret,
		"{KAFKA_REST_ENDPOINT}":        pm.config.KafkaRestEndpoint,
		"{FLINK_ORG_ID}":               pm.config.FlinkOrgID,
		"{FLINK_REST_ENDPOINT}":        pm.config.FlinkRestEndpoint,
		"{FLINK_ENV_NAME}":             pm.config.FlinkEnvName,
		"{FLINK_DATABASE_NAME}":        pm.config.FlinkDatabaseName,
		"{FLINK_API_KEY}":              pm.config.FlinkAPIKey,
		"{FLINK_API_SECRET}":           pm.config.FlinkAPISecret,
		"{FLINK_COMPUTE_POOL_ID}":      pm.config.FlinkComputePoolID,
		"{SCHEMA_REGISTRY_API_KEY}":    pm.config.SchemaRegistryAPIKey,
		"{SCHEMA_REGISTRY_API_SECRET}": pm.config.SchemaRegistryAPISecret,
		"{SCHEMA_REGISTRY_ENDPOINT}":   pm.config.SchemaRegistryEndpoint,
		"{TABLEFLOW_API_KEY}":          pm.config.TableflowAPIKey,
		"{TABLEFLOW_API_SECRET}":       pm.config.TableflowAPISecret,

		// Parameter format (same as tools use) - more user-friendly
		"{environment}":              pm.config.ConfluentEnvID,
		"{environment_id}":           pm.config.ConfluentEnvID,
		"{cluster_id}":               pm.config.KafkaClusterID,
		"{kafka_cluster_id}":         pm.config.KafkaClusterID,
		"{compute_pool_id}":          pm.config.FlinkComputePoolID,
		"{pool_id}":                  pm.config.FlinkComputePoolID,
		"{organization_id}":          pm.config.FlinkOrgID,
		"{org_id}":                   pm.config.FlinkOrgID,
		"{org}":                      pm.config.FlinkOrgID,
		"{schema_registry_endpoint}": pm.config.SchemaRegistryEndpoint,
		"{bootstrap_servers}":        pm.config.BootstrapServers,
		"{kafka_rest_endpoint}":      pm.config.KafkaRestEndpoint,
		"{flink_rest_endpoint}":      pm.config.FlinkRestEndpoint,
		"{flink_env_name}":           pm.config.FlinkEnvName,
		"{flink_database_name}":      pm.config.FlinkDatabaseName,
	}

	// Perform replacements
	result := promptText
	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result, nil
}

// GetPromptContentWithSubstitution returns the content of a specific prompt with variable substitution and directives
func (pm *PromptManager) GetPromptContentWithSubstitution(name string) (string, error) {
	content, err := pm.GetPromptContent(name)
	if err != nil {
		return "", err
	}

	substituted, err := pm.substituteVariables(content)
	if err != nil {
		return "", err
	}

	// Compose with directives
	return pm.composePromptWithDirectives(substituted), nil
}

// GetPromptContentWithArguments returns the content of a specific prompt with variable substitution, argument overrides, and directives
func (pm *PromptManager) GetPromptContentWithArguments(name string, args map[string]interface{}) (string, error) {
	content, exists := pm.promptContent[name]
	if !exists {
		return "", fmt.Errorf("prompt '%s' not found", name)
	}

	// Start with original content
	result := content

	// Apply argument overrides first (support both formats)
	if environmentID, ok := args["environment_id"].(string); ok && environmentID != "" {
		// Replace both environment variable format and parameter format
		result = strings.ReplaceAll(result, "{CONFLUENT_ENV_ID}", environmentID)
		result = strings.ReplaceAll(result, "{environment}", environmentID)
		result = strings.ReplaceAll(result, "{environment_id}", environmentID)
	}
	if clusterID, ok := args["cluster_id"].(string); ok && clusterID != "" {
		// Replace both environment variable format and parameter format
		result = strings.ReplaceAll(result, "{KAFKA_CLUSTER_ID}", clusterID)
		result = strings.ReplaceAll(result, "{cluster_id}", clusterID)
		result = strings.ReplaceAll(result, "{kafka_cluster_id}", clusterID)
	}
	if computePoolID, ok := args["compute_pool_id"].(string); ok && computePoolID != "" {
		// Replace both environment variable format and parameter format
		result = strings.ReplaceAll(result, "{FLINK_COMPUTE_POOL_ID}", computePoolID)
		result = strings.ReplaceAll(result, "{compute_pool_id}", computePoolID)
		result = strings.ReplaceAll(result, "{pool_id}", computePoolID)
	}
	if orgID, ok := args["organization_id"].(string); ok && orgID != "" {
		// Replace both environment variable format and parameter format
		result = strings.ReplaceAll(result, "{FLINK_ORG_ID}", orgID)
		result = strings.ReplaceAll(result, "{organization_id}", orgID)
		result = strings.ReplaceAll(result, "{org_id}", orgID)
		result = strings.ReplaceAll(result, "{org}", orgID)
	}

	// Then apply default config substitutions for any remaining placeholders
	substituted, err := pm.substituteVariables(result)
	if err != nil {
		return "", err
	}

	// Compose with directives
	return pm.composePromptWithDirectives(substituted), nil
}

// loadDirectives loads all .txt files from the directives folder and combines them
func (pm *PromptManager) loadDirectives() error {
	// Check if directives are enabled
	if pm.config != nil && !pm.config.EnableDirectives {
		// Directives are disabled
		return nil
	}

	if pm.directivesFolder == "" {
		// No directives folder configured
		return nil
	}

	// Check if directives folder exists
	if _, err := os.Stat(pm.directivesFolder); os.IsNotExist(err) {
		// Directives folder doesn't exist, this is not an error
		return nil
	}

	// Read all .txt files in the directives folder
	files, err := filepath.Glob(filepath.Join(pm.directivesFolder, "*.txt"))
	if err != nil {
		return fmt.Errorf("failed to read directives folder: %w", err)
	}

	var allDirectives []string

	// Load each directive file
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read directive file %s: %w", file, err)
		}

		// Trim whitespace and add to collection
		directive := strings.TrimSpace(string(content))
		if directive != "" {
			allDirectives = append(allDirectives, directive)
		}
	}

	// Combine all directives with double newlines for separation
	pm.directives = strings.Join(allDirectives, "\n\n")

	return nil
}

// GetDirectives returns the combined directives content
func (pm *PromptManager) GetDirectives() string {
	return pm.directives
}

// composePromptWithDirectives combines directives with a prompt
func (pm *PromptManager) composePromptWithDirectives(promptContent string) string {
	if pm.directives == "" {
		return promptContent
	}
	return pm.directives + "\n\n" + promptContent
}

// SetDirectivesFolder sets the directives folder path (useful for testing)
func (pm *PromptManager) SetDirectivesFolder(folder string) {
	pm.directivesFolder = folder
}
