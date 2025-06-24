package tools

import (
	"mcolomerc/mcp-server/internal/openapi"
	"sync"
)

// HTTP method constants
const (
	HTTPMethodGet    = "GET"
	HTTPMethodPost   = "POST"
	HTTPMethodPut    = "PUT"
	HTTPMethodPatch  = "PATCH"
	HTTPMethodDelete = "DELETE"
)

// Common parameter types
const (
	ParamTypeString  = "string"
	ParamTypeInteger = "integer"
	ParamTypeNumber  = "number"
	ParamTypeBoolean = "boolean"
	ParamTypeArray   = "array"
	ParamTypeObject  = "object"
)

// Content type constants
const (
	ContentTypeJSON = "application/json"
)

// HTTPOperation represents an HTTP operation with its details
type HTTPOperation struct {
	Method    string
	Operation *openapi.Operation
	HasBody   bool
}

// Tool represents a dynamically generated tool based on OpenAPI spec
type Tool struct {
	Name        string
	Description string
	Endpoint    string
	Parameters  map[string]interface{} // JSON Schema parameters
}

// Semantic action constants
const (
	ActionCreate = "create"
	ActionList   = "list"
	ActionGet    = "get"
	ActionUpdate = "update"
	ActionDelete = "delete"
)

// getAllSemanticActions returns all supported semantic actions
func getAllSemanticActions() []string {
	return []string{ActionCreate, ActionList, ActionGet, ActionUpdate, ActionDelete}
}

// EndpointMapping represents the mapping from semantic action+resource to API endpoint
type EndpointMapping struct {
	Method            string                 // HTTP method
	PathPattern       string                 // API path pattern with {placeholders}
	RequiredParams    []string               // Required parameters for this endpoint
	OptionalParams    []string               // Optional parameters
	RequestBodySchema map[string]interface{} // Schema for request body if applicable
}

// SemanticToolRegistry holds all the mappings for semantic tools
type SemanticToolRegistry struct {
	Mappings map[string]map[string]EndpointMapping // action -> resource -> endpoint mapping
	Spec     *openapi.OpenAPISpec                  // Reference to the spec for resolving references
	mutex    sync.RWMutex                          // Protects concurrent access
}

// EnvironmentVariable holds the mapping between path parameters and environment variables
type EnvironmentVariable struct {
	Parameter string
	EnvVar    string
}

// RequestBodyInfo holds schema and content type for request bodies
type RequestBodyInfo struct {
	Schema      interface{}
	ContentType string
}

// getDefaultEnvVarMappings returns default environment variable mappings
func getDefaultEnvVarMappings() []EnvironmentVariable {
	return []EnvironmentVariable{
		{"clusterId", "KAFKA_CLUSTER_ID"},
		{"environmentId", "CONFLUENT_ENV_ID"},
		{"orgId", "FLINK_ORG_ID"},
		{"apiKey", "KAFKA_API_KEY"},
		{"computePoolId", "FLINK_COMPUTE_POOL_ID"},
		{"databaseName", "FLINK_DATABASE_NAME"},
		{"envId", "CONFLUENT_ENV_ID"},
		{"apiKeyId", "SCHEMA_REGISTRY_API_KEY"},
	}
}

// PostSpecialOperations contains special path suffixes that indicate update action for POST
var PostSpecialOperations = []string{":batch", ":alter", "/request", "/undelete"}

// CollectionEndpoints lists common collection paths
var CollectionEndpoints = []string{
	"/topics", "/clusters", "/subjects", "/schemas", "/connectors",
	"/consumers", "/partitions", "/configs",
}

// SpecificResourceEndpoints lists paths that indicate specific resource operations
var SpecificResourceEndpoints = []string{"/offsets", "/status", "/versions"}
