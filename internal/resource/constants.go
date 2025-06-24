package resource

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

// Generic identifier field patterns for fallback
var GenericIDFieldPatterns = []string{"id", "name", "_id", "_name"}
