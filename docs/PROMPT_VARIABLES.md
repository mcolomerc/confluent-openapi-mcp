# Prompt Variable Format Examples

This document demonstrates the two supported variable formats in prompts, making them consistent with tool parameters.

## Supported Variable Formats

### 1. Parameter Format (Recommended)

This format matches the parameter names used in tools, making it more user-friendly:

```
Environment: {environment_id}
Cluster: {cluster_id}
Compute Pool: {compute_pool_id}
Organization: {org_id}
Schema Registry: {schema_registry_endpoint}
Bootstrap Servers: {bootstrap_servers}
```

### 2. Environment Variable Format (Legacy)

This format matches the actual environment variable names:

```
Environment: {CONFLUENT_ENV_ID}
Cluster: {KAFKA_CLUSTER_ID}
Compute Pool: {FLINK_COMPUTE_POOL_ID}
Organization: {FLINK_ORG_ID}
Schema Registry: {SCHEMA_REGISTRY_ENDPOINT}
Bootstrap Servers: {BOOTSTRAP_SERVERS}
```

## Complete Variable Mapping

| Parameter Format | Environment Variable Format | Config Field |
|-----------------|----------------------------|--------------|
| `{environment_id}` | `{CONFLUENT_ENV_ID}` | ConfluentEnvID |
| `{environment}` | `{CONFLUENT_ENV_ID}` | ConfluentEnvID |
| `{cluster_id}` | `{KAFKA_CLUSTER_ID}` | KafkaClusterID |
| `{kafka_cluster_id}` | `{KAFKA_CLUSTER_ID}` | KafkaClusterID |
| `{compute_pool_id}` | `{FLINK_COMPUTE_POOL_ID}` | FlinkComputePoolID |
| `{pool_id}` | `{FLINK_COMPUTE_POOL_ID}` | FlinkComputePoolID |
| `{organization_id}` | `{FLINK_ORG_ID}` | FlinkOrgID |
| `{org_id}` | `{FLINK_ORG_ID}` | FlinkOrgID |
| `{org}` | `{FLINK_ORG_ID}` | FlinkOrgID |
| `{schema_registry_endpoint}` | `{SCHEMA_REGISTRY_ENDPOINT}` | SchemaRegistryEndpoint |
| `{bootstrap_servers}` | `{BOOTSTRAP_SERVERS}` | BootstrapServers |
| `{kafka_rest_endpoint}` | `{KAFKA_REST_ENDPOINT}` | KafkaRestEndpoint |
| `{flink_rest_endpoint}` | `{FLINK_REST_ENDPOINT}` | FlinkRestEndpoint |
| `{flink_env_name}` | `{FLINK_ENV_NAME}` | FlinkEnvName |
| `{flink_database_name}` | `{FLINK_DATABASE_NAME}` | FlinkDatabaseName |

## Example Usage

Both formats work identically:

```markdown
# Kafka Topic Analysis
Analyze topics in cluster {cluster_id} within environment {environment_id}.

# Same as:
Analyze topics in cluster {KAFKA_CLUSTER_ID} within environment {CONFLUENT_ENV_ID}.
```

## Client Override Support

Clients can override default values using prompt arguments:

```json
{
  "name": "my-prompt",
  "arguments": {
    "cluster_id": "lkc-custom123",
    "environment_id": "env-custom456"
  }
}
```
 
