package prompts

import (
	"mcolomerc/mcp-server/internal/config"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVariableSubstitution(t *testing.T) {
	tempDir := t.TempDir()

	testContent := `# Test
Environment: {CONFLUENT_ENV_ID}
Cluster: {KAFKA_CLUSTER_ID}`

	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		ConfluentEnvID: "env-123",
		KafkaClusterID: "lkc-456",
	}

	pm := NewPromptManager(tempDir, cfg)
	if err := pm.LoadPrompts(); err != nil {
		t.Fatal(err)
	}

	content, err := pm.GetPromptContentWithSubstitution("test")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(content, "env-123") {
		t.Error("Expected substituted env ID")
	}
	if !strings.Contains(content, "lkc-456") {
		t.Error("Expected substituted cluster ID")
	}
}

func TestBothVariableFormats(t *testing.T) {
	tempDir := t.TempDir()

	// Test content with both environment variable format and parameter format
	testContent := `# Test Both Formats
Environment Variable Format:
- Environment: {CONFLUENT_ENV_ID}
- Cluster: {KAFKA_CLUSTER_ID}
- Compute Pool: {FLINK_COMPUTE_POOL_ID}
- Organization: {FLINK_ORG_ID}

Parameter Format (same as tools):
- Environment: {environment_id}
- Cluster: {cluster_id}
- Compute Pool: {compute_pool_id}
- Organization: {org_id}

Mixed Usage:
- Use environment {environment} with cluster {KAFKA_CLUSTER_ID}
- Deploy to pool {pool_id} in org {FLINK_ORG_ID}`

	testFile := filepath.Join(tempDir, "test-formats.txt")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		ConfluentEnvID:     "env-test123",
		KafkaClusterID:     "lkc-test456",
		FlinkComputePoolID: "lfcp-test789",
		FlinkOrgID:         "org-test000",
	}

	pm := NewPromptManager(tempDir, cfg)
	if err := pm.LoadPrompts(); err != nil {
		t.Fatal(err)
	}

	content, err := pm.GetPromptContentWithSubstitution("test-formats")
	if err != nil {
		t.Fatal(err)
	}

	// Verify all environment variable format substitutions
	if !strings.Contains(content, "env-test123") {
		t.Error("Expected substituted env ID from CONFLUENT_ENV_ID")
	}
	if !strings.Contains(content, "lkc-test456") {
		t.Error("Expected substituted cluster ID from KAFKA_CLUSTER_ID")
	}
	if !strings.Contains(content, "lfcp-test789") {
		t.Error("Expected substituted compute pool ID from FLINK_COMPUTE_POOL_ID")
	}
	if !strings.Contains(content, "org-test000") {
		t.Error("Expected substituted org ID from FLINK_ORG_ID")
	}

	// Verify no unsubstituted placeholders remain
	if strings.Contains(content, "{CONFLUENT_ENV_ID}") {
		t.Error("Found unsubstituted CONFLUENT_ENV_ID placeholder")
	}
	if strings.Contains(content, "{environment_id}") {
		t.Error("Found unsubstituted environment_id placeholder")
	}
	if strings.Contains(content, "{cluster_id}") {
		t.Error("Found unsubstituted cluster_id placeholder")
	}
	if strings.Contains(content, "{compute_pool_id}") {
		t.Error("Found unsubstituted compute_pool_id placeholder")
	}
	if strings.Contains(content, "{org_id}") {
		t.Error("Found unsubstituted org_id placeholder")
	}

	t.Logf("Substituted content:\n%s", content)
}

func TestArgumentOverridesWithBothFormats(t *testing.T) {
	tempDir := t.TempDir()

	// Test content with both formats that should be overridden by arguments
	testContent := `# Test Argument Overrides
Original env: {CONFLUENT_ENV_ID}
Original cluster: {cluster_id}
Original pool: {compute_pool_id}
Original org: {org_id}`

	testFile := filepath.Join(tempDir, "test-overrides.txt")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		ConfluentEnvID:     "env-default",
		KafkaClusterID:     "lkc-default",
		FlinkComputePoolID: "lfcp-default",
		FlinkOrgID:         "org-default",
	}

	pm := NewPromptManager(tempDir, cfg)
	if err := pm.LoadPrompts(); err != nil {
		t.Fatal(err)
	}

	// Test with argument overrides
	args := map[string]interface{}{
		"environment_id":  "env-override",
		"cluster_id":      "lkc-override",
		"compute_pool_id": "lfcp-override",
		"organization_id": "org-override",
	}

	content, err := pm.GetPromptContentWithArguments("test-overrides", args)
	if err != nil {
		t.Fatal(err)
	}

	// Verify overrides took effect
	if !strings.Contains(content, "env-override") {
		t.Error("Expected overridden environment ID")
	}
	if !strings.Contains(content, "lkc-override") {
		t.Error("Expected overridden cluster ID")
	}
	if !strings.Contains(content, "lfcp-override") {
		t.Error("Expected overridden compute pool ID")
	}
	if !strings.Contains(content, "org-override") {
		t.Error("Expected overridden org ID")
	}

	// Verify defaults were not used
	if strings.Contains(content, "env-default") {
		t.Error("Found default env ID instead of override")
	}
	if strings.Contains(content, "lkc-default") {
		t.Error("Found default cluster ID instead of override")
	}
	if strings.Contains(content, "lfcp-default") {
		t.Error("Found default compute pool ID instead of override")
	}
	if strings.Contains(content, "org-default") {
		t.Error("Found default org ID instead of override")
	}

	t.Logf("Content with overrides:\n%s", content)
}
