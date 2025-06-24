// Package tools contains constants and types for API resource parsing and generation.
// This file centralizes magic strings and configuration values to improve maintainability.
package tools

// API Path Constants
const (
	// API Versions - used for identifying different API generations in paths
	KafkaAPIVersion   = "/kafka/v3/"
	ConnectAPIVersion = "/connect/v1/"

	// Version Pattern - used for identifying version segments in paths
	VersionPrefix    = "v"
	MaxVersionLength = 3

	// Resource Names - known resource types in the Confluent API
	ConfigsResource       = "configs"        // Excluded as standalone resource (always sub-resource)
	TopicsResource        = "topics"         // Kafka topics
	BrokerConfigsResource = "broker-configs" // Broker configurations
	ClustersResource      = "clusters"       // Kafka clusters

	// Path Separators and Markers
	PathSeparator   = "/"
	PathParamPrefix = "{"
	PathParamSuffix = "}"

	// Minimum lengths for validation - used in heuristic resource detection
	MinResourceNameLength       = 3 // Minimum length for a valid resource name
	MinPathPartLength           = 2 // Minimum length for any path component
	MinHyphenatedResourceLength = 4 // Minimum length for hyphenated resources

	// Plural endings for resource detection
	PluralSuffix = "s"
)

// Common plural endings for API resources - used in heuristic resource name detection
var CommonPluralEndings = []string{
	"ies", "es", "ings", "ers", "ors", "ants", "ents",
}

// Content Types - used for HTTP request/response handling
const (
	ContentTypeConfluentJSON = "application/vnd.confluent+json" // Confluent-specific JSON format
)
