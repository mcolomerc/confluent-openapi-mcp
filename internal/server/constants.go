package server

// Security Types
const (
	SecurityTypeCloudAPIKey    = "cloud-api-key"
	SecurityTypeResourceAPIKey = "resource-api-key"
)

// Service/Endpoint Patterns
const (
	// Kafka patterns
	EndpointPatternKafka          = "kafka"
	EndpointPatternTopics         = "/topics/"
	EndpointPatternConsumerGroups = "/consumer-groups/"
	EndpointPatternACLs           = "/acls"
	EndpointPatternConfigs        = "/configs"
	EndpointPatternKafkaV3        = "/kafka/v3/"

	// Flink patterns
	EndpointPatternFlink        = "flink"
	EndpointPatternComputePools = "/compute-pools/"
	EndpointPatternStatements   = "/statements/"

	// Schema Registry patterns
	EndpointPatternSchemaRegistry      = "schema-registry"
	EndpointPatternSchemaRegistryShort = "schemaregistry"
	EndpointPatternSchemas             = "/schemas/"
	EndpointPatternSubjects            = "/subjects/"
	EndpointPatternMode                = "/mode"
	EndpointPatternConfig              = "/config"
	EndpointPatternExporters           = "/exporters"
	EndpointPatternContexts            = "/contexts"
	EndpointPatternDekRegistry         = "/dek-registry/"
	EndpointPatternCatalog             = "/catalog/"

	// TableFlow patterns
	EndpointPatternTableFlow = "tableflow"
	EndpointPatternTF        = "/tableflow/"

	// General patterns
	EndpointPatternEnvironment  = "environment"
	EndpointPatternOrganization = "organization"
	EndpointPatternSchema       = "schema"
)

// Parameter Names
const (
	// Environment parameters
	ParamEnvironment   = "environment"
	ParamEnvironmentID = "environment_id"

	// Cluster parameters
	ParamClusterID      = "cluster_id"
	ParamKafkaClusterID = "kafka_cluster_id"

	// Compute Pool parameters
	ParamComputePoolID = "compute_pool_id"
	ParamPoolID        = "pool_id"

	// Organization parameters
	ParamOrganizationID = "organization_id"
	ParamOrgID          = "org_id"
	ParamOrg            = "org"

	// Schema Registry parameters
	ParamSchemaRegistryEndpoint = "schema_registry_endpoint"

	// Configuration parameters - used in request body transformation
	ParamConfigs = "configs" // Array of configuration objects
	ParamConfig  = "config"  // Single configuration object
)

// Property Types - used for schema validation and transformation
const (
	PropertyTypeArray = "array" // JSON Schema array type
)

// Default Base URLs
const (
	BaseURLConfluentCloud     = "https://api.confluent.cloud"
	BaseURLConfluentTelemetry = "https://api.telemetry.confluent.cloud"
)

// HTTP Configuration
const (
	HTTPTimeoutSeconds = 30
	ContentTypeJSON    = "application/json"
	HeaderContentType  = "Content-Type"
	HeaderAccept       = "Accept"
	HeaderAuth         = "Authorization"
	AuthBasicPrefix    = "Basic "
)
