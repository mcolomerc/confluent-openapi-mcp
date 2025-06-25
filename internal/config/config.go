package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all required environment variables for the server
// All fields are required and validated except LOG which is optional
// Use this struct instead of accessing os.Getenv directly
type Config struct {
	OpenAPISpecURL          string
	TelemetryOpenAPISpecURL string
	ConfluentEnvID          string
	ConfluentCloudAPIKey    string
	ConfluentCloudAPISecret string
	BootstrapServers        string
	KafkaAPIKey             string
	KafkaAPISecret          string
	KafkaRestEndpoint       string
	KafkaClusterID          string
	FlinkOrgID              string
	FlinkRestEndpoint       string
	FlinkEnvName            string
	FlinkDatabaseName       string
	FlinkAPIKey             string
	FlinkAPISecret          string
	FlinkComputePoolID      string
	SchemaRegistryAPIKey    string
	SchemaRegistryAPISecret string
	SchemaRegistryEndpoint  string
	TableflowAPIKey         string
	TableflowAPISecret      string
	LOG                     string // Optional: DEBUG, INFO, etc.
	PromptsFolder           string // Optional: folder path containing prompt .txt files
}

// LoadConfig loads and validates configuration from environment variables
func LoadConfig(path string) (*Config, error) {

	_ = godotenv.Load(path)

	cfg := &Config{
		OpenAPISpecURL:          os.Getenv("OPENAPI_SPEC_URL"),
		TelemetryOpenAPISpecURL: os.Getenv("TELEMETRY_OPENAPI_SPEC_URL"),
		ConfluentEnvID:          os.Getenv("CONFLUENT_ENV_ID"),
		ConfluentCloudAPIKey:    os.Getenv("CONFLUENT_CLOUD_API_KEY"),
		ConfluentCloudAPISecret: os.Getenv("CONFLUENT_CLOUD_API_SECRET"),
		BootstrapServers:        os.Getenv("BOOTSTRAP_SERVERS"),
		KafkaAPIKey:             os.Getenv("KAFKA_API_KEY"),
		KafkaAPISecret:          os.Getenv("KAFKA_API_SECRET"),
		KafkaRestEndpoint:       os.Getenv("KAFKA_REST_ENDPOINT"),
		KafkaClusterID:          os.Getenv("KAFKA_CLUSTER_ID"),
		FlinkOrgID:              os.Getenv("FLINK_ORG_ID"),
		FlinkRestEndpoint:       os.Getenv("FLINK_REST_ENDPOINT"),
		FlinkEnvName:            os.Getenv("FLINK_ENV_NAME"),
		FlinkDatabaseName:       os.Getenv("FLINK_DATABASE_NAME"),
		FlinkAPIKey:             os.Getenv("FLINK_API_KEY"),
		FlinkAPISecret:          os.Getenv("FLINK_API_SECRET"),
		FlinkComputePoolID:      os.Getenv("FLINK_COMPUTE_POOL_ID"),
		SchemaRegistryAPIKey:    os.Getenv("SCHEMA_REGISTRY_API_KEY"),
		SchemaRegistryAPISecret: os.Getenv("SCHEMA_REGISTRY_API_SECRET"),
		SchemaRegistryEndpoint:  os.Getenv("SCHEMA_REGISTRY_ENDPOINT"),
		TableflowAPIKey:         os.Getenv("TABLEFLOW_API_KEY"),
		TableflowAPISecret:      os.Getenv("TABLEFLOW_API_SECRET"),
		LOG:                     os.Getenv("LOG"),            // Optional field
		PromptsFolder:           os.Getenv("PROMPTS_FOLDER"), // Optional field
	}

	missing := []string{}
	fields := map[string]string{
		"CONFLUENT_ENV_ID":           cfg.ConfluentEnvID,
		"CONFLUENT_CLOUD_API_KEY":    cfg.ConfluentCloudAPIKey,
		"CONFLUENT_CLOUD_API_SECRET": cfg.ConfluentCloudAPISecret,
		"BOOTSTRAP_SERVERS":          cfg.BootstrapServers,
		"KAFKA_API_KEY":              cfg.KafkaAPIKey,
		"KAFKA_API_SECRET":           cfg.KafkaAPISecret,
		"KAFKA_REST_ENDPOINT":        cfg.KafkaRestEndpoint,
		"KAFKA_CLUSTER_ID":           cfg.KafkaClusterID,
		"FLINK_ORG_ID":               cfg.FlinkOrgID,
		"FLINK_REST_ENDPOINT":        cfg.FlinkRestEndpoint,
		"FLINK_ENV_NAME":             cfg.FlinkEnvName,
		"FLINK_DATABASE_NAME":        cfg.FlinkDatabaseName,
		"FLINK_API_KEY":              cfg.FlinkAPIKey,
		"FLINK_API_SECRET":           cfg.FlinkAPISecret,
		"FLINK_COMPUTE_POOL_ID":      cfg.FlinkComputePoolID,
		"SCHEMA_REGISTRY_API_KEY":    cfg.SchemaRegistryAPIKey,
		"SCHEMA_REGISTRY_API_SECRET": cfg.SchemaRegistryAPISecret,
		"SCHEMA_REGISTRY_ENDPOINT":   cfg.SchemaRegistryEndpoint,
		"TABLEFLOW_API_KEY":          cfg.TableflowAPIKey,
		"TABLEFLOW_API_SECRET":       cfg.TableflowAPISecret,
	}
	for k, v := range fields {
		if v == "" {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %v", missing)
	}

	// Content validation
	if !strings.HasPrefix(cfg.ConfluentEnvID, "env-") {
		return nil, errors.New("CONFLUENT_ENV_ID must start with 'env-'")
	}
	if !strings.HasPrefix(cfg.KafkaClusterID, "lkc-") {
		return nil, errors.New("KAFKA_CLUSTER_ID must start with 'lkc-'")
	}
	if !strings.HasPrefix(cfg.FlinkComputePoolID, "lfcp-") {
		return nil, errors.New("FLINK_COMPUTE_POOL_ID must start with 'lfcp-'")
	}
	if _, err := url.ParseRequestURI(cfg.SchemaRegistryEndpoint); err != nil {
		return nil, errors.New("SCHEMA_REGISTRY_ENDPOINT must be a valid URL")
	}

	return cfg, nil
}
