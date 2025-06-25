package server

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mcolomerc/mcp-server/internal/config"
	"mcolomerc/mcp-server/internal/logger"
	"mcolomerc/mcp-server/internal/types"
	"net/http"
	"net/url"
	"strings"
	"time"

	"mcolomerc/mcp-server/internal/openapi"
)

// Re-export types for convenience
type InvokeRequest = types.InvokeRequest
type InvokeResponse = types.InvokeResponse

// Credentials represents an API key-secret pair
type Credentials struct {
	Key    string
	Secret string
}

// Helper to get API credentials based on security type and endpoint
func getAPICredentials(cfg *config.Config, securityType, endpoint string) (apiKey, apiSecret string) {
	logger.Debug("getAPICredentials called with securityType=%s, endpoint=%s", securityType, endpoint)

	// Temporary debug for telemetry
	if strings.Contains(strings.ToLower(endpoint), "telemetry") || strings.Contains(strings.ToLower(endpoint), "metrics") {
		logger.Debug("*** TELEMETRY DEBUG: securityType=%s, endpoint=%s", securityType, endpoint)
	}

	// Special debug for regions
	if strings.Contains(endpoint, "regions") {
		logger.Debug("*** REGIONS CREDENTIALS DEBUG: securityType=%s, endpoint=%s", securityType, endpoint)
	}

	switch securityType {
	case SecurityTypeCloudAPIKey:
		logger.Debug("Using Cloud API credentials for cloud-api-key")
		if strings.Contains(endpoint, "regions") {
			logger.Debug("*** REGIONS: Using Cloud API Key=%s, Secret=%s", cfg.ConfluentCloudAPIKey[:8]+"...", cfg.ConfluentCloudAPISecret[:8]+"...")
		}
		return cfg.ConfluentCloudAPIKey, cfg.ConfluentCloudAPISecret
	case "api-key":
		logger.Debug("Using Cloud API credentials for api-key")
		// Hardcode telemetry credentials to test
		if strings.Contains(strings.ToLower(endpoint), "metrics") || strings.Contains(strings.ToLower(endpoint), "telemetry") {
			logger.Debug("*** TELEMETRY: Using hardcoded credentials")
			return "HE5P5PRAMML3HVTW", "l1FE+CpfyWgV5QGM4olu6NSme0xrvABC7yMBTAeafftEOQ1eLiObb2yQeAGZo3Ua"
		}
		return cfg.ConfluentCloudAPIKey, cfg.ConfluentCloudAPISecret
	case SecurityTypeResourceAPIKey:
		// Check for telemetry endpoints first - they should use Cloud API credentials
		if strings.Contains(strings.ToLower(endpoint), "/v2/metrics/") ||
			strings.Contains(strings.ToLower(endpoint), "/v2/descriptors/") ||
			strings.Contains(strings.ToLower(endpoint), "/telemetry/") {
			logger.Debug("Telemetry endpoint detected, using Cloud API credentials")
			return cfg.ConfluentCloudAPIKey, cfg.ConfluentCloudAPISecret
		}

		// Map endpoint patterns to their corresponding credentials
		resourceCredentials := map[string]Credentials{
			EndpointPatternKafka:               {cfg.KafkaAPIKey, cfg.KafkaAPISecret},
			EndpointPatternKafkaV3:             {cfg.KafkaAPIKey, cfg.KafkaAPISecret},
			EndpointPatternTopics:              {cfg.KafkaAPIKey, cfg.KafkaAPISecret},
			EndpointPatternConsumerGroups:      {cfg.KafkaAPIKey, cfg.KafkaAPISecret},
			EndpointPatternACLs:                {cfg.KafkaAPIKey, cfg.KafkaAPISecret},
			EndpointPatternConfigs:             {cfg.KafkaAPIKey, cfg.KafkaAPISecret},
			EndpointPatternFlink:               {cfg.FlinkAPIKey, cfg.FlinkAPISecret},
			EndpointPatternSchemaRegistry:      {cfg.SchemaRegistryAPIKey, cfg.SchemaRegistryAPISecret},
			EndpointPatternSchemaRegistryShort: {cfg.SchemaRegistryAPIKey, cfg.SchemaRegistryAPISecret},
			EndpointPatternSchemas:             {cfg.SchemaRegistryAPIKey, cfg.SchemaRegistryAPISecret},
			EndpointPatternSubjects:            {cfg.SchemaRegistryAPIKey, cfg.SchemaRegistryAPISecret},
			EndpointPatternMode:                {cfg.SchemaRegistryAPIKey, cfg.SchemaRegistryAPISecret},
			EndpointPatternConfig:              {cfg.SchemaRegistryAPIKey, cfg.SchemaRegistryAPISecret},
			EndpointPatternExporters:           {cfg.SchemaRegistryAPIKey, cfg.SchemaRegistryAPISecret},
			EndpointPatternContexts:            {cfg.SchemaRegistryAPIKey, cfg.SchemaRegistryAPISecret},
			EndpointPatternDekRegistry:         {cfg.SchemaRegistryAPIKey, cfg.SchemaRegistryAPISecret},
			EndpointPatternCatalog:             {cfg.SchemaRegistryAPIKey, cfg.SchemaRegistryAPISecret},
			EndpointPatternTableFlow:           {cfg.TableflowAPIKey, cfg.TableflowAPISecret},
		}

		// Check endpoint against each pattern
		endpointLower := strings.ToLower(endpoint)
		logger.Debug("Checking endpoint '%s' against patterns", endpointLower)
		for pattern, creds := range resourceCredentials {
			logger.Debug("Checking pattern '%s'", pattern)
			// Handle exact matches and prefix matches
			if strings.Contains(endpointLower, pattern) ||
				(strings.HasSuffix(pattern, "/") && endpointLower == strings.TrimSuffix(pattern, "/")) {
				logger.Debug("Pattern '%s' matched! Using credentials: key=%s, secret=%s", pattern, creds.Key[:8]+"...", creds.Secret[:8]+"...")

				// Special logging for catalog/tagdefs
				if strings.Contains(endpointLower, "catalog") || strings.Contains(endpointLower, "tagdefs") {
					logger.Debug("*** CATALOG/TAGDEFS CREDENTIALS: endpoint=%s, pattern=%s, key=%s", endpointLower, pattern, creds.Key[:8]+"...")
				}

				return creds.Key, creds.Secret
			}
		}
		logger.Debug("No patterns matched for endpoint '%s'", endpointLower)
	default:
		// For unknown security types (like "api-key" from telemetry spec), try using Cloud API credentials
		logger.Debug("Unknown security type '%s', trying Cloud API credentials", securityType)
		return cfg.ConfluentCloudAPIKey, cfg.ConfluentCloudAPISecret
	}
	logger.Debug("Returning empty credentials")
	return "", ""
}

// Helper to resolve default parameter values from Config
func resolveDefaultParam(cfg *config.Config, paramName, endpoint string) string {
	paramLower := strings.ToLower(paramName)
	endpointLower := strings.ToLower(endpoint)

	// Map parameter patterns to their corresponding config values and matching conditions
	paramMappings := []struct {
		paramPatterns    []string
		endpointPatterns []string
		getValue         func() string
	}{
		{
			paramPatterns:    []string{ParamEnvironment, ParamEnvironmentID},
			endpointPatterns: []string{EndpointPatternEnvironment},
			getValue:         func() string { return cfg.ConfluentEnvID },
		},
		{
			paramPatterns:    []string{ParamClusterID, ParamKafkaClusterID},
			endpointPatterns: []string{EndpointPatternKafka},
			getValue:         func() string { return cfg.KafkaClusterID },
		},
		{
			paramPatterns:    []string{ParamComputePoolID, ParamPoolID},
			endpointPatterns: []string{EndpointPatternFlink},
			getValue:         func() string { return cfg.FlinkComputePoolID },
		},
		{
			paramPatterns:    []string{ParamOrganizationID, ParamOrgID, ParamOrg},
			endpointPatterns: []string{EndpointPatternOrganization},
			getValue:         func() string { return cfg.FlinkOrgID },
		},
		{
			paramPatterns:    []string{ParamSchemaRegistryEndpoint},
			endpointPatterns: []string{EndpointPatternSchema},
			getValue:         func() string { return cfg.SchemaRegistryEndpoint },
		},
	}

	// Check each mapping for parameter and endpoint matches
	for _, mapping := range paramMappings {
		// Check if parameter name matches any pattern
		paramMatches := false
		for _, pattern := range mapping.paramPatterns {
			if paramLower == pattern || strings.Contains(paramLower, pattern) {
				paramMatches = true
				break
			}
		}

		// Check if endpoint matches any pattern
		endpointMatches := false
		for _, pattern := range mapping.endpointPatterns {
			if strings.Contains(endpointLower, pattern) {
				endpointMatches = true
				break
			}
		}

		// If either parameter or endpoint matches, try to get the value
		if paramMatches || endpointMatches {
			if value := mapping.getValue(); value != "" {
				return value
			}
		}
	}

	return ""
}

// DetermineSecurityTypeFromSpec determines the security type for an endpoint using the OpenAPI specification
func DetermineSecurityTypeFromSpec(spec *openapi.OpenAPISpec, method, path string) string {
	if spec != nil {
		securityType := spec.GetSecurityTypeForEndpoint(method, path)
		if securityType != "" {
			return securityType
		}
	}

	// Fallback to cloud API key if no specific security type is found
	logger.Debug("No specific security type found for %s %s, defaulting to cloud-api-key", method, path)
	return "cloud-api-key"
}

// Resolve all required parameters for an endpoint
func ResolveRequiredParameters(cfg *config.Config, requiredParams []string, providedParams map[string]interface{}, pathPattern string) map[string]interface{} {
	resolved := make(map[string]interface{})

	// Copy provided parameters
	for k, v := range providedParams {
		resolved[k] = v
	}

	// Resolve missing required parameters
	for _, param := range requiredParams {
		if _, exists := resolved[param]; !exists || resolved[param] == "" || resolved[param] == nil {
			if defaultVal := resolveDefaultParam(cfg, param, pathPattern); defaultVal != "" {
				resolved[param] = defaultVal
			}
		}
	}

	return resolved
}

// Execute API call to Confluent Cloud
func ExecuteAPICall(cfg *config.Config, spec *openapi.OpenAPISpec, method, path string, parameters map[string]interface{}, requestBody interface{}) (map[string]interface{}, error) {
	logger.Debug("ExecuteAPICall called with method=%s, path=%s, parameters=%v, requestBody=%v\n", method, path, parameters, requestBody)

	// Special logging for tagdefs
	if strings.Contains(path, "tagdefs") {
		logger.Debug("*** TAGDEFS API CALL: method=%s, path=%s", method, path)
	}

	// Determine security type using the OpenAPI spec or fallback to static approach
	securityType := DetermineSecurityTypeFromSpec(spec, method, path)

	// Get appropriate API credentials
	apiKey, apiSecret := getAPICredentials(cfg, securityType, path)
	if apiKey == "" || apiSecret == "" {
		return nil, fmt.Errorf("missing API credentials for security type: %s", securityType)
	}

	// Determine base URL based on path
	baseURL := getBaseURL(cfg, path)
	if baseURL == "" {
		return nil, fmt.Errorf("could not determine base URL for path: %s", path)
	}

	// Special logging for tagdefs URL construction
	if strings.Contains(path, "tagdefs") {
		logger.Debug("*** TAGDEFS URL: baseURL=%s, path=%s", baseURL, path)
	}

	// Build full URL with query parameters
	fullURL := baseURL + path
	if len(parameters) > 0 && method == "GET" {
		queryValues := url.Values{}
		for key, value := range parameters {
			// Only add parameters that aren't already in the path
			if !strings.Contains(path, "{"+key+"}") {
				queryValues.Add(key, fmt.Sprintf("%v", value))
			}
		}
		if len(queryValues) > 0 {
			fullURL += "?" + queryValues.Encode()
		}
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: HTTPTimeoutSeconds * time.Second,
	}

	// Prepare request body
	var bodyReader io.Reader
	if requestBody != nil {
		bodyBytes, err := json.Marshal(requestBody)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %v", err)
		}
		logger.Debug("Final JSON request body: %s\n", string(bodyBytes))
		logger.Debug("Final JSON request body: %s\n", string(bodyBytes))
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Create HTTP request
	req, err := http.NewRequest(method, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Special logging for tagdefs final URL
	if strings.Contains(path, "tagdefs") {
		logger.Debug("*** TAGDEFS FINAL REQUEST: %s %s", method, fullURL)
	}

	// Set headers
	req.Header.Set(HeaderContentType, ContentTypeJSON)
	req.Header.Set(HeaderAccept, ContentTypeJSON)

	// Set authentication
	auth := base64.StdEncoding.EncodeToString([]byte(apiKey + ":" + apiSecret))
	req.Header.Set(HeaderAuth, AuthBasicPrefix+auth)

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Check status code
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	// Parse JSON response
	var result map[string]interface{}
	if len(responseBody) > 0 {
		if err := json.Unmarshal(responseBody, &result); err != nil {
			// If JSON parsing fails, return raw response
			return map[string]interface{}{
				"raw_response": string(responseBody),
				"status_code":  resp.StatusCode,
			}, nil
		}
	}

	// Add status code to result
	if result == nil {
		result = make(map[string]interface{})
	}
	result["status_code"] = resp.StatusCode

	return result, nil
}

// Get base URL based on the API path
func getBaseURL(cfg *config.Config, path string) string {
	pathLower := strings.ToLower(path)

	// Map path patterns to their corresponding base URLs and config fields
	pathMappings := []struct {
		patterns []string
		getURL   func() string
	}{
		{
			patterns: []string{"/v2/metrics/", "/v2/descriptors/", "/telemetry/"},
			getURL:   func() string { return BaseURLConfluentTelemetry },
		},
		{
			patterns: []string{"/kafka/", EndpointPatternTopics, EndpointPatternConsumerGroups, EndpointPatternACLs},
			getURL:   func() string { return cfg.KafkaRestEndpoint },
		},
		{
			patterns: []string{"/flink/", EndpointPatternComputePools, EndpointPatternStatements},
			getURL:   func() string { return cfg.FlinkRestEndpoint },
		},
		{
			patterns: []string{EndpointPatternSchemas, EndpointPatternSubjects, EndpointPatternMode, EndpointPatternConfig, EndpointPatternCatalog, EndpointPatternExporters, EndpointPatternContexts, EndpointPatternDekRegistry},
			getURL:   func() string { return cfg.SchemaRegistryEndpoint },
		},
		{
			patterns: []string{EndpointPatternTF},
			getURL:   func() string { return BaseURLConfluentCloud },
		},
	}

	// Check path against each pattern group
	for _, mapping := range pathMappings {
		for _, pattern := range mapping.patterns {
			if strings.Contains(pathLower, pattern) ||
				(strings.HasSuffix(pattern, "/") && pathLower == strings.TrimSuffix(pattern, "/")) {
				if baseURL := mapping.getURL(); baseURL != "" {
					// Special logging for catalog/tagdefs
					if strings.Contains(pathLower, "catalog") || strings.Contains(pathLower, "tagdefs") {
						logger.Debug("*** CATALOG/TAGDEFS BASE URL: path=%s, pattern=%s, baseURL=%s", pathLower, pattern, baseURL)
					}
					return baseURL
				}
			}
		}
	}

	// Default to Confluent Cloud API
	return BaseURLConfluentCloud
}
