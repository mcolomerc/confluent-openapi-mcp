package resource

import "strings"

// URI and Protocol Constants
const (
	// URI scheme for Confluent resources
	ConfluentURIScheme = "confluent://"

	// URI format: confluent://resourceType/resourceId
	URIPathSeparator = "/"
)

// Resource Field Names - used for extracting resource information from API responses
var (
	// Common ID field names (in order of preference)
	CommonIDFields = []string{"id", "name", "topic_name", "cluster_id", "connector_name"}

	// Common description field names (in order of preference)
	CommonDescriptionFields = []string{"description", "summary", "status"}

	// Common array field names in API responses (in order of preference)
	CommonArrayFields = []string{"data", "items", "results"}
)

// Resource Type Parameter Mappings - maps resource types to their identifier parameter names
var ResourceTypeIDMappings = map[string][]string{
	"topics":           {"topic_name", "topicName"},
	"connectors":       {"connector_name", "name"},
	"service-accounts": {"service_account_id", "id"},
}

// Resource types that should be excluded from MCP resource registration
// These are typically sub-resources, configurations, or metadata that are not standalone resources
var ExcludedResourceTypes = []string{
	"configs",         // Topic/cluster configurations - these are properties, not resources
	"mode",            // Schema registry compatibility mode - metadata, not a resource
	"config",          // Global configuration - metadata, not a resource
	"compatibility",   // Compatibility settings - metadata, not a resource
	"versions",        // Schema versions - sub-resources of subjects
	"status",          // Status information - metadata, not a resource
	"offsets",         // Consumer group offsets - metadata, not a resource
	"lags",            // Consumer group lag information - metadata, not a resource
	"partitions",      // Topic partitions - sub-resources of topics
	"default-configs", // Default configurations - metadata templates
}

// Generic identifier field patterns for fallback
var GenericIDFieldPatterns = []string{"id", "name", "_id", "_name"}

// IsExcludedResourceType checks if a resource type should be excluded from MCP resource registration
func IsExcludedResourceType(resourceType string) bool {
	for _, excluded := range ExcludedResourceTypes {
		if strings.EqualFold(resourceType, excluded) {
			return true
		}
	}
	return false
}
