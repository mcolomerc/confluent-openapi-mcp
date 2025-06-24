package server

import (
	"mcolomerc/mcp-server/internal/config"
	"testing"
)

func TestGetAPICredentials(t *testing.T) {
	// Create a test config with all credential types
	cfg := &config.Config{
		ConfluentCloudAPIKey:    "cloud-key",
		ConfluentCloudAPISecret: "cloud-secret",
		KafkaAPIKey:             "kafka-key",
		KafkaAPISecret:          "kafka-secret",
		FlinkAPIKey:             "flink-key",
		FlinkAPISecret:          "flink-secret",
		SchemaRegistryAPIKey:    "sr-key",
		SchemaRegistryAPISecret: "sr-secret",
		TableflowAPIKey:         "tableflow-key",
		TableflowAPISecret:      "tableflow-secret",
	}

	tests := []struct {
		name           string
		securityType   string
		endpoint       string
		expectedKey    string
		expectedSecret string
	}{
		{
			name:           "Cloud API Key",
			securityType:   SecurityTypeCloudAPIKey,
			endpoint:       "/iam/v2/environments",
			expectedKey:    "cloud-key",
			expectedSecret: "cloud-secret",
		},
		{
			name:           "Kafka endpoint",
			securityType:   SecurityTypeResourceAPIKey,
			endpoint:       "/kafka/v3/clusters/lkc-abc123/topics",
			expectedKey:    "kafka-key",
			expectedSecret: "kafka-secret",
		},
		{
			name:           "Flink endpoint",
			securityType:   SecurityTypeResourceAPIKey,
			endpoint:       "/flink/v1/compute-pools",
			expectedKey:    "flink-key",
			expectedSecret: "flink-secret",
		},
		{
			name:           "Schema Registry endpoint (hyphenated)",
			securityType:   SecurityTypeResourceAPIKey,
			endpoint:       "/schema-registry/v1/subjects",
			expectedKey:    "sr-key",
			expectedSecret: "sr-secret",
		},
		{
			name:           "Schema Registry endpoint (no hyphen)",
			securityType:   SecurityTypeResourceAPIKey,
			endpoint:       "/schemaregistry/v1/subjects",
			expectedKey:    "sr-key",
			expectedSecret: "sr-secret",
		},
		{
			name:           "TableFlow endpoint",
			securityType:   SecurityTypeResourceAPIKey,
			endpoint:       "/tableflow/v1/tableflow-topics",
			expectedKey:    "tableflow-key",
			expectedSecret: "tableflow-secret",
		},
		{
			name:           "Unknown resource endpoint",
			securityType:   SecurityTypeResourceAPIKey,
			endpoint:       "/unknown/v1/resources",
			expectedKey:    "",
			expectedSecret: "",
		},
		{
			name:           "Unknown security type",
			securityType:   "unknown-type",
			endpoint:       "/any/endpoint",
			expectedKey:    "",
			expectedSecret: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, secret := getAPICredentials(cfg, tt.securityType, tt.endpoint)

			if key != tt.expectedKey {
				t.Errorf("Expected key %q, got %q", tt.expectedKey, key)
			}

			if secret != tt.expectedSecret {
				t.Errorf("Expected secret %q, got %q", tt.expectedSecret, secret)
			}
		})
	}
}

func TestCredentialsMapBased(t *testing.T) {
	cfg := &config.Config{
		TableflowAPIKey:    "tf-key",
		TableflowAPISecret: "tf-secret",
		KafkaAPIKey:        "k-key",
		KafkaAPISecret:     "k-secret",
	}

	// Test that map-based approach correctly resolves credentials
	t.Run("TableFlow credentials resolution", func(t *testing.T) {
		key, secret := getAPICredentials(cfg, "resource-api-key", "/tableflow/v1/regions")
		if key != "tf-key" || secret != "tf-secret" {
			t.Errorf("TableFlow credentials not resolved correctly. Got key=%q, secret=%q", key, secret)
		}
	})

	t.Run("Kafka credentials resolution", func(t *testing.T) {
		key, secret := getAPICredentials(cfg, "resource-api-key", "/kafka/v3/topics")
		if key != "k-key" || secret != "k-secret" {
			t.Errorf("Kafka credentials not resolved correctly. Got key=%q, secret=%q", key, secret)
		}
	})

	t.Run("Case insensitive endpoint matching", func(t *testing.T) {
		key, secret := getAPICredentials(cfg, "resource-api-key", "/TABLEFLOW/V1/REGIONS")
		if key != "tf-key" || secret != "tf-secret" {
			t.Errorf("Case insensitive matching failed. Got key=%q, secret=%q", key, secret)
		}
	})
}

func TestGetBaseURL(t *testing.T) {
	// Create a test config with all endpoint types
	cfg := &config.Config{
		KafkaRestEndpoint:      "https://pkc-12345.us-west-2.aws.confluent.cloud:443",
		FlinkRestEndpoint:      "https://flink.us-west-2.aws.confluent.cloud",
		SchemaRegistryEndpoint: "https://psrc-abc123.us-west-2.aws.confluent.cloud",
	}

	tests := []struct {
		name        string
		path        string
		expectedURL string
	}{
		// Kafka REST API paths
		{
			name:        "Kafka topics endpoint",
			path:        "/kafka/v3/clusters/lkc-12345/topics",
			expectedURL: "https://pkc-12345.us-west-2.aws.confluent.cloud:443",
		},
		{
			name:        "Kafka consumer groups endpoint",
			path:        "/consumer-groups/my-group",
			expectedURL: "https://pkc-12345.us-west-2.aws.confluent.cloud:443",
		},
		{
			name:        "Kafka ACLs endpoint",
			path:        "/acls",
			expectedURL: "https://pkc-12345.us-west-2.aws.confluent.cloud:443",
		},
		{
			name:        "Direct topics path",
			path:        "/topics/my-topic",
			expectedURL: "https://pkc-12345.us-west-2.aws.confluent.cloud:443",
		},

		// Flink REST API paths
		{
			name:        "Flink compute pools endpoint",
			path:        "/flink/v1/compute-pools",
			expectedURL: "https://flink.us-west-2.aws.confluent.cloud",
		},
		{
			name:        "Flink statements endpoint",
			path:        "/statements/my-statement",
			expectedURL: "https://flink.us-west-2.aws.confluent.cloud",
		},
		{
			name:        "Direct compute pools path",
			path:        "/compute-pools/lfcp-12345",
			expectedURL: "https://flink.us-west-2.aws.confluent.cloud",
		},

		// Schema Registry API paths
		{
			name:        "Schema Registry subjects endpoint",
			path:        "/schemas/ids/123",
			expectedURL: "https://psrc-abc123.us-west-2.aws.confluent.cloud",
		},
		{
			name:        "Schema Registry subjects endpoint",
			path:        "/subjects/my-subject",
			expectedURL: "https://psrc-abc123.us-west-2.aws.confluent.cloud",
		},
		{
			name:        "Schema Registry mode endpoint",
			path:        "/mode/subject",
			expectedURL: "https://psrc-abc123.us-west-2.aws.confluent.cloud",
		},
		{
			name:        "Schema Registry config endpoint",
			path:        "/config",
			expectedURL: "https://psrc-abc123.us-west-2.aws.confluent.cloud",
		},

		// TableFlow API paths
		{
			name:        "TableFlow topics endpoint",
			path:        "/tableflow/v1/tableflow-topics",
			expectedURL: BaseURLConfluentCloud,
		},
		{
			name:        "TableFlow regions endpoint",
			path:        "/tableflow/v1/regions",
			expectedURL: BaseURLConfluentCloud,
		},

		// Default case
		{
			name:        "Confluent Cloud API endpoint",
			path:        "/iam/v2/environments",
			expectedURL: BaseURLConfluentCloud,
		},
		{
			name:        "Unknown endpoint",
			path:        "/unknown/v1/resources",
			expectedURL: BaseURLConfluentCloud,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getBaseURL(cfg, tt.path)
			if result != tt.expectedURL {
				t.Errorf("Expected URL %q, got %q", tt.expectedURL, result)
			}
		})
	}
}

func TestGetBaseURLCaseInsensitive(t *testing.T) {
	cfg := &config.Config{
		KafkaRestEndpoint: "https://kafka.test.com",
		FlinkRestEndpoint: "https://flink.test.com",
	}

	tests := []struct {
		name        string
		path        string
		expectedURL string
	}{
		{
			name:        "Uppercase Kafka path",
			path:        "/KAFKA/V3/TOPICS",
			expectedURL: "https://kafka.test.com",
		},
		{
			name:        "Mixed case Flink path",
			path:        "/Flink/V1/Compute-Pools",
			expectedURL: "https://flink.test.com",
		},
		{
			name:        "Uppercase TableFlow path",
			path:        "/TABLEFLOW/V1/REGIONS",
			expectedURL: BaseURLConfluentCloud,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getBaseURL(cfg, tt.path)
			if result != tt.expectedURL {
				t.Errorf("Expected URL %q, got %q", tt.expectedURL, result)
			}
		})
	}
}

func TestGetBaseURLEmptyEndpoints(t *testing.T) {
	// Test with empty endpoint configurations
	cfg := &config.Config{
		KafkaRestEndpoint:      "", // Empty
		FlinkRestEndpoint:      "", // Empty
		SchemaRegistryEndpoint: "", // Empty
	}

	tests := []struct {
		name        string
		path        string
		expectedURL string
	}{
		{
			name:        "Kafka path with empty endpoint",
			path:        "/kafka/v3/topics",
			expectedURL: BaseURLConfluentCloud, // Should fall back to default
		},
		{
			name:        "Flink path with empty endpoint",
			path:        "/flink/v1/compute-pools",
			expectedURL: BaseURLConfluentCloud, // Should fall back to default
		},
		{
			name:        "Schema Registry path with empty endpoint",
			path:        "/subjects/test",
			expectedURL: BaseURLConfluentCloud, // Should fall back to default
		},
		{
			name:        "TableFlow path (always uses default)",
			path:        "/tableflow/v1/topics",
			expectedURL: BaseURLConfluentCloud,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getBaseURL(cfg, tt.path)
			if result != tt.expectedURL {
				t.Errorf("Expected URL %q, got %q", tt.expectedURL, result)
			}
		})
	}
}

func TestResolveDefaultParam(t *testing.T) {
	// Create a test config with all parameter types
	cfg := &config.Config{
		ConfluentEnvID:         "env-test123",
		KafkaClusterID:         "lkc-test456",
		FlinkComputePoolID:     "lfcp-test789",
		FlinkOrgID:             "org-test000",
		SchemaRegistryEndpoint: "https://psrc-test.aws.confluent.cloud",
	}

	tests := []struct {
		name          string
		paramName     string
		endpoint      string
		expectedValue string
		description   string
	}{
		// Environment ID tests
		{
			name:          "Environment by exact param name",
			paramName:     "environment",
			endpoint:      "/some/endpoint",
			expectedValue: "env-test123",
			description:   "Should match 'environment' parameter name",
		},
		{
			name:          "Environment ID by exact param name",
			paramName:     "environment_id",
			endpoint:      "/some/endpoint",
			expectedValue: "env-test123",
			description:   "Should match 'environment_id' parameter name",
		},
		{
			name:          "Environment by endpoint pattern",
			paramName:     "some_param",
			endpoint:      "/iam/v2/environments/env-123",
			expectedValue: "env-test123",
			description:   "Should match endpoint containing 'environment'",
		},

		// Kafka cluster ID tests
		{
			name:          "Cluster ID by exact param name",
			paramName:     "cluster_id",
			endpoint:      "/some/endpoint",
			expectedValue: "lkc-test456",
			description:   "Should match 'cluster_id' parameter name",
		},
		{
			name:          "Kafka cluster ID by exact param name",
			paramName:     "kafka_cluster_id",
			endpoint:      "/some/endpoint",
			expectedValue: "lkc-test456",
			description:   "Should match 'kafka_cluster_id' parameter name",
		},
		{
			name:          "Cluster ID by endpoint pattern",
			paramName:     "some_param",
			endpoint:      "/kafka/v3/clusters/lkc-123/topics",
			expectedValue: "lkc-test456",
			description:   "Should match endpoint containing 'kafka'",
		},

		// Flink compute pool ID tests
		{
			name:          "Compute pool ID by exact param name",
			paramName:     "compute_pool_id",
			endpoint:      "/some/endpoint",
			expectedValue: "lfcp-test789",
			description:   "Should match 'compute_pool_id' parameter name",
		},
		{
			name:          "Pool ID by exact param name",
			paramName:     "pool_id",
			endpoint:      "/some/endpoint",
			expectedValue: "lfcp-test789",
			description:   "Should match 'pool_id' parameter name",
		},
		{
			name:          "Compute pool ID by endpoint pattern",
			paramName:     "some_param",
			endpoint:      "/flink/v1/compute-pools/lfcp-123",
			expectedValue: "lfcp-test789",
			description:   "Should match endpoint containing 'flink'",
		},

		// Organization ID tests
		{
			name:          "Organization ID by exact param name",
			paramName:     "organization_id",
			endpoint:      "/some/endpoint",
			expectedValue: "org-test000",
			description:   "Should match 'organization_id' parameter name",
		},
		{
			name:          "Org ID by exact param name",
			paramName:     "org_id",
			endpoint:      "/some/endpoint",
			expectedValue: "org-test000",
			description:   "Should match 'org_id' parameter name",
		},
		{
			name:          "Org by param name substring",
			paramName:     "my_org_param",
			endpoint:      "/some/endpoint",
			expectedValue: "org-test000",
			description:   "Should match parameter name containing 'org'",
		},
		{
			name:          "Organization by endpoint pattern",
			paramName:     "some_param",
			endpoint:      "/iam/v2/organizations/org-123",
			expectedValue: "org-test000",
			description:   "Should match endpoint containing 'organization'",
		},

		// Schema Registry endpoint tests
		{
			name:          "Schema Registry endpoint by exact param name",
			paramName:     "schema_registry_endpoint",
			endpoint:      "/some/endpoint",
			expectedValue: "https://psrc-test.aws.confluent.cloud",
			description:   "Should match 'schema_registry_endpoint' parameter name",
		},
		{
			name:          "Schema Registry by endpoint pattern",
			paramName:     "some_param",
			endpoint:      "/schema-registry/v1/subjects",
			expectedValue: "https://psrc-test.aws.confluent.cloud",
			description:   "Should match endpoint containing 'schema'",
		},

		// No match tests
		{
			name:          "Unknown parameter and endpoint",
			paramName:     "unknown_param",
			endpoint:      "/unknown/endpoint",
			expectedValue: "",
			description:   "Should return empty string for unknown parameter and endpoint",
		},
		{
			name:          "Known parameter but empty config",
			paramName:     "environment",
			endpoint:      "/some/endpoint",
			expectedValue: "",
			description:   "Should return empty string when config value is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For the empty config test, use a config with empty values
			testCfg := cfg
			if tt.name == "Known parameter but empty config" {
				testCfg = &config.Config{} // Empty config
			}

			result := resolveDefaultParam(testCfg, tt.paramName, tt.endpoint)
			if result != tt.expectedValue {
				t.Errorf("Expected %q, got %q. %s", tt.expectedValue, result, tt.description)
			}
		})
	}
}

func TestResolveDefaultParamCaseInsensitive(t *testing.T) {
	cfg := &config.Config{
		ConfluentEnvID:     "env-test123",
		KafkaClusterID:     "lkc-test456",
		FlinkComputePoolID: "lfcp-test789",
	}

	tests := []struct {
		name          string
		paramName     string
		endpoint      string
		expectedValue string
	}{
		{
			name:          "Uppercase parameter name",
			paramName:     "ENVIRONMENT_ID",
			endpoint:      "/some/endpoint",
			expectedValue: "env-test123",
		},
		{
			name:          "Mixed case parameter name",
			paramName:     "Cluster_Id",
			endpoint:      "/some/endpoint",
			expectedValue: "lkc-test456",
		},
		{
			name:          "Uppercase endpoint",
			paramName:     "some_param",
			endpoint:      "/KAFKA/V3/TOPICS",
			expectedValue: "lkc-test456",
		},
		{
			name:          "Mixed case endpoint",
			paramName:     "some_param",
			endpoint:      "/Flink/V1/Compute-Pools",
			expectedValue: "lfcp-test789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolveDefaultParam(cfg, tt.paramName, tt.endpoint)
			if result != tt.expectedValue {
				t.Errorf("Expected %q, got %q", tt.expectedValue, result)
			}
		})
	}
}

func TestResolveDefaultParamParameterPriority(t *testing.T) {
	cfg := &config.Config{
		ConfluentEnvID: "env-test123",
		KafkaClusterID: "lkc-test456",
	}

	// Test that parameter name matching works when both param and endpoint could match different patterns
	tests := []struct {
		name          string
		paramName     string
		endpoint      string
		expectedValue string
		description   string
	}{
		{
			name:          "Parameter name takes precedence",
			paramName:     "environment_id",
			endpoint:      "/kafka/v3/topics", // This would match Kafka, but param name should win
			expectedValue: "env-test123",
			description:   "Parameter name match should take precedence over endpoint match",
		},
		{
			name:          "Endpoint matching when param doesn't match",
			paramName:     "some_generic_param",
			endpoint:      "/kafka/v3/topics",
			expectedValue: "lkc-test456",
			description:   "Should fall back to endpoint matching when parameter name doesn't match",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolveDefaultParam(cfg, tt.paramName, tt.endpoint)
			if result != tt.expectedValue {
				t.Errorf("Expected %q, got %q. %s", tt.expectedValue, result, tt.description)
			}
		})
	}
}
