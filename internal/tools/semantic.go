package tools

import (
	"fmt"
	"mcolomerc/mcp-server/internal/logger"
	"mcolomerc/mcp-server/internal/openapi"
	"os"
	"strings"
	"sync"
)

// GlobalSemanticRegistry is the global registry for semantic tools
var GlobalSemanticRegistry *SemanticToolRegistry
var registryMutex sync.RWMutex

// initializeSemanticRegistry sets up the semantic tool mappings dynamically from OpenAPI spec
func initializeSemanticRegistry(spec openapi.OpenAPISpec) {
	registryMutex.Lock()
	defer registryMutex.Unlock()

	logger.Debug("Building semantic registry from OpenAPI spec with %d paths\n", len(spec.Paths))

	GlobalSemanticRegistry = &SemanticToolRegistry{
		Mappings: make(map[string]map[string]EndpointMapping),
		Spec:     &spec,
	}

	// Initialize action maps
	actions := getAllSemanticActions()
	for _, action := range actions {
		GlobalSemanticRegistry.Mappings[action] = make(map[string]EndpointMapping)
	}

	// Parse OpenAPI paths and categorize them
	for path, pathItem := range spec.Paths {
		resource := ExtractResourceFromPath(path)
		if resource == "" {
			continue
		}

		// Special debug logging for tags resource
		if resource == "tags" || resource == "tagdefs" {
			logger.Debug("Processing %s resource from path: %s\n", resource, path)
		}

		// Process each HTTP method using the operations we extracted
		operations := extractHTTPOperations(&pathItem)
		for _, op := range operations {
			action := determineSemanticAction(op.Method, path)
			if action != "" {
				mapping := createEndpointMapping(op.Method, path, op.Operation, &spec)
				GlobalSemanticRegistry.Mappings[action][resource] = mapping

				// Special debug logging for tags resource
				if resource == "tags" || resource == "tagdefs" {
					logger.Debug("*** TAGS DEBUG: Mapped %s %s -> %s %s (required params: %v)\n",
						action, resource, mapping.Method, mapping.PathPattern, mapping.RequiredParams)
				} else {
					logger.Debug("Mapped %s %s -> %s %s\n", action, resource, mapping.Method, mapping.PathPattern)
				}
			} else if resource == "tags" || resource == "tagdefs" {
				logger.Debug("*** TAGS DEBUG: No action determined for %s %s (path: %s)\n", op.Method, resource, path)
			}
		}
	}

	// Log summary
	for action, resources := range GlobalSemanticRegistry.Mappings {
		if len(resources) > 0 {
			logger.Debug("Action '%s' supports %d resources\n", action, len(resources))
		}
	}

	// Log discovered resources for validation
	logDiscoveredResources(&spec)
}

// GenerateSemanticTools creates semantic tools from OpenAPI spec
func GenerateSemanticTools(spec openapi.OpenAPISpec) ([]Tool, error) {
	logger.Debug("Generating semantic tools from %d paths\n", len(spec.Paths))

	// Initialize the semantic registry with the OpenAPI spec
	initializeSemanticRegistry(spec)

	var tools []Tool

	// Create semantic tools based on our registry
	for action, resourceMappings := range GlobalSemanticRegistry.Mappings {
		if len(resourceMappings) == 0 {
			continue // Skip actions with no resources
		}

		var supportedResources []string
		for resource := range resourceMappings {
			supportedResources = append(supportedResources, resource)
		}

		tool := Tool{
			Name:        action,
			Description: fmt.Sprintf("%s resources. Supported resources: %s", strings.Title(action), strings.Join(supportedResources, ", ")),
			Endpoint:    action,
			Parameters:  createSemanticToolParameters(action, supportedResources),
		}

		tools = append(tools, tool)
	}

	logger.Debug("Generated %d semantic tools\n", len(tools))
	return tools, nil
}

// createSemanticToolParameters creates parameters for semantic tools
func createSemanticToolParameters(action string, supportedResources []string) map[string]interface{} {
	properties := map[string]interface{}{
		"resource": map[string]interface{}{
			"type":        "string",
			"description": fmt.Sprintf("The type of resource to %s", action),
			"enum":        supportedResources,
		},
	}

	// Add dynamic parameters section that will be populated based on resource choice
	properties["parameters"] = map[string]interface{}{
		"type":        "object",
		"description": "Parameters specific to the chosen resource and action",
		"properties":  map[string]interface{}{},
	}

	return map[string]interface{}{
		"type":       "object",
		"properties": properties,
		"required":   []string{"resource"},
	}
}

// GetEndpointMapping retrieves the endpoint mapping for a given action and resource
func GetEndpointMapping(action, resource string) (*EndpointMapping, error) {
	registryMutex.RLock()
	defer registryMutex.RUnlock()

	if GlobalSemanticRegistry == nil {
		return nil, fmt.Errorf("semantic registry not initialized")
	}

	resourceMappings, exists := GlobalSemanticRegistry.Mappings[action]
	if !exists {
		return nil, fmt.Errorf("action '%s' not supported", action)
	}

	mapping, exists := resourceMappings[resource]
	if !exists {
		return nil, fmt.Errorf("resource '%s' not supported for action '%s'", resource, action)
	}

	return &mapping, nil
}

// GetRequiredParametersForResource returns the required parameters for a specific action+resource combination
func GetRequiredParametersForResource(action, resource string) ([]string, error) {
	mapping, err := GetEndpointMapping(action, resource)
	if err != nil {
		return nil, err
	}
	return mapping.RequiredParams, nil
}

// GetParameterSchemaForResource returns the full parameter schema (request body schema) for a specific action+resource combination
func GetParameterSchemaForResource(action, resource string) (map[string]interface{}, error) {
	mapping, err := GetEndpointMapping(action, resource)
	if err != nil {
		return nil, err
	}
	return mapping.RequestBodySchema, nil
}

// PathParamEnvVarMap returns the mapping of path parameters to environment variables
func PathParamEnvVarMap() map[string]string {
	envMap := make(map[string]string)
	for _, mapping := range getDefaultEnvVarMappings() {
		envMap[mapping.Parameter] = mapping.EnvVar
	}
	return envMap
}

// BuildAPIPath builds the actual API path by replacing placeholders with values
func BuildAPIPath(pathPattern string, params map[string]interface{}) string {
	path := pathPattern

	// First, fill from params if present
	for key, value := range params {
		placeholder := fmt.Sprintf("{%s}", key)
		if strings.Contains(path, placeholder) {
			path = strings.ReplaceAll(path, placeholder, fmt.Sprintf("%v", value))
		}
	}

	// Then, fill from env vars if still present
	envVarMap := PathParamEnvVarMap()
	for param, envVar := range envVarMap {
		placeholder := fmt.Sprintf("{%s}", param)
		if strings.Contains(path, placeholder) {
			if val := os.Getenv(envVar); val != "" {
				path = strings.ReplaceAll(path, placeholder, val)
			}
		}
	}

	return path
}

// ExtractResourceFromPath extracts the primary resource name from an API path
func ExtractResourceFromPath(path string) string {
	parts := strings.Split(strings.TrimPrefix(path, PathSeparator), PathSeparator)
	candidateResources := findCandidateResources(parts)

	if len(candidateResources) > 1 {
		return selectBestResource(path, candidateResources)
	}

	if len(candidateResources) == 1 {
		return candidateResources[0]
	}

	return findFallbackResource(parts)
}

// findCandidateResources identifies potential resource names from path parts
func findCandidateResources(parts []string) []string {
	var candidateResources []string

	for _, part := range parts {
		if isVersionOrCommonPath(part) || isPathParameter(part) {
			continue
		}

		if isLikelyResourceName(part) {
			candidateResources = append(candidateResources, part)
		}
	}

	return candidateResources
}

// selectBestResource chooses the most appropriate resource from candidates
func selectBestResource(path string, candidates []string) string {
	// For Kafka REST API paths, prioritize the final resource
	if strings.Contains(path, KafkaAPIVersion) || strings.Contains(path, ConnectAPIVersion) {
		return candidates[len(candidates)-1]
	}

	// For other nested cases, prefer the last resource mentioned
	return candidates[len(candidates)-1]
}

// findFallbackResource finds a resource when no clear candidate exists
func findFallbackResource(parts []string) string {
	for i := len(parts) - 1; i >= 0; i-- {
		part := parts[i]
		if !isPathParameter(part) && !isVersionOrCommonPath(part) {
			return part
		}
	}
	return ""
}

// isPathParameter checks if a part is a path parameter
func isPathParameter(part string) bool {
	return strings.HasPrefix(part, PathParamPrefix) && strings.HasSuffix(part, PathParamSuffix)
}

// isVersionOrCommonPath checks if a path part is a version identifier or likely non-resource
func isVersionOrCommonPath(part string) bool {
	// Check for version patterns (v1, v2, v3, etc.)
	if isVersionPattern(part) {
		return true
	}

	// Use simple heuristics - if it's not likely a resource, it's probably infrastructure
	return !isLikelyResourceName(part)
}

// isVersionPattern checks if a part matches version pattern
func isVersionPattern(part string) bool {
	return strings.HasPrefix(part, VersionPrefix) && len(part) <= MaxVersionLength
}

// isLikelyResourceName determines if a path component looks like a resource name
func isLikelyResourceName(part string) bool {
	// Exclude "configs" as it's not a standalone resource - it's always a sub-resource
	if part == ConfigsResource {
		return false
	}

	// Use heuristics only - no hardcoded lists
	return isPluralResourceName(part)
}

// isPluralResourceName checks if a part looks like a plural resource name using improved heuristics
func isPluralResourceName(part string) bool {
	// Basic length check
	if len(part) <= MinPathPartLength {
		return false
	}

	// Check for common plural patterns
	if hasCommonPluralPattern(part) {
		return true
	}

	// Check for hyphenated resources (like "api-keys", "role-bindings")
	if strings.Contains(part, "-") && len(part) > MinHyphenatedResourceLength {
		// Split and check if the last part looks plural
		parts := strings.Split(part, "-")
		lastPart := parts[len(parts)-1]
		return strings.HasSuffix(lastPart, PluralSuffix) && len(lastPart) > MinPathPartLength
	}

	// Simple plural check as fallback
	return strings.HasSuffix(part, PluralSuffix) && len(part) > MinResourceNameLength
}

// hasCommonPluralPattern checks for common plural patterns in API resources
func hasCommonPluralPattern(part string) bool {
	for _, ending := range CommonPluralEndings {
		if strings.HasSuffix(part, ending) && len(part) > len(ending)+1 {
			return true
		}
	}

	return false
}

// determineSemanticAction maps HTTP method and path pattern to semantic action
func determineSemanticAction(httpMethod, path string) string {
	// Special handling for catalog entity tag operations
	if strings.Contains(path, "/catalog/v1/entity/tags") && !strings.Contains(path, "/{") {
		// Bulk tag operations (no path parameters)
		switch httpMethod {
		case HTTPMethodPost:
			return ActionCreate
		case HTTPMethodPut:
			return ActionUpdate
		}
	}

	if strings.Contains(path, "/catalog/v1/entity/type/{typeName}/name/{qualifiedName}/tags") {
		// Individual entity tag operations
		switch httpMethod {
		case HTTPMethodGet:
			if strings.Contains(path, "/{tagName}") {
				return ActionGet // Get specific tag
			}
			return ActionList // List all tags for entity
		case HTTPMethodDelete:
			return ActionDelete
		}
	}

	// Standard logic for other endpoints
	switch httpMethod {
	case HTTPMethodGet:
		return determineGetAction(path)
	case HTTPMethodPost:
		return determinePostAction(path)
	case HTTPMethodPut:
		return ActionUpdate
	case HTTPMethodPatch:
		return ActionUpdate
	case HTTPMethodDelete:
		return ActionDelete
	default:
		return ""
	}
}

// determineGetAction determines if GET operation is list or get
func determineGetAction(path string) string {
	// If path has no parameters, it's likely a list operation
	if !strings.Contains(path, "{") {
		return ActionList
	}

	// Check for collection endpoints (these are list operations even with params)
	if isCollectionEndpoint(path) {
		return ActionList
	}

	// Check for specific resource endpoints (these are get operations)
	if isSpecificResourceEndpoint(path) {
		return ActionGet
	}

	// Default: if path contains parameters, it's likely a get operation
	return ActionGet
}

// determinePostAction determines the action for POST operations
func determinePostAction(path string) string {
	// POST operations with special suffixes are usually update operations
	for _, op := range PostSpecialOperations {
		if strings.Contains(path, op) {
			return ActionUpdate
		}
	}
	return ActionCreate
}

// isCollectionEndpoint checks if path ends with a collection name
func isCollectionEndpoint(path string) bool {
	for _, endpoint := range CollectionEndpoints {
		if strings.HasSuffix(path, endpoint) {
			return true
		}
	}
	return false
}

// isSpecificResourceEndpoint checks if path points to a specific resource
func isSpecificResourceEndpoint(path string) bool {
	for _, endpoint := range SpecificResourceEndpoints {
		if strings.Contains(path, endpoint) {
			return true
		}
	}
	return false
}

// createEndpointMapping creates an EndpointMapping from HTTP method, path, and operation
func createEndpointMapping(httpMethod, path string, operation *openapi.Operation, spec *openapi.OpenAPISpec) EndpointMapping {
	if operation == nil || spec == nil {
		logger.Debug("Warning: nil operation or spec provided to createEndpointMapping")
		return EndpointMapping{
			Method:      httpMethod,
			PathPattern: path,
		}
	}

	mapping := EndpointMapping{
		Method:      httpMethod,
		PathPattern: path,
	}

	// Extract parameters from operation
	mapping.RequiredParams, mapping.OptionalParams = extractOperationParameters(operation)

	// Extract path parameters and ensure they're marked as required
	mapping.RequiredParams = ensurePathParametersRequired(path, mapping.RequiredParams)

	// Extract request body info if present
	if operation.RequestBody != nil {
		if info := extractRequestBodySchema(operation.RequestBody, spec); info != nil {
			// Store schema and content type in a map
			mapping.RequestBodySchema = map[string]interface{}{
				"schema":      info.Schema,
				"contentType": info.ContentType,
			}
			// If schema is a map, add its required fields
			if schemaMap, ok := info.Schema.(map[string]interface{}); ok {
				mapping.RequiredParams = addRequiredFieldsFromSchema(
					map[string]interface{}{"schema": schemaMap}, mapping.RequiredParams,
				)
			}
		}
	}

	return mapping
}

// extractOperationParameters extracts required and optional parameters from operation
func extractOperationParameters(operation *openapi.Operation) (required, optional []string) {
	for _, param := range operation.Parameters {
		if param.Required {
			required = append(required, param.Name)
		} else {
			optional = append(optional, param.Name)
		}
	}
	return required, optional
}

// ensurePathParametersRequired ensures all path parameters are marked as required
func ensurePathParametersRequired(path string, existingRequired []string) []string {
	pathParams := ExtractPathParameters(path)
	requiredSet := make(map[string]bool)

	// Add existing required params to set
	for _, param := range existingRequired {
		requiredSet[param] = true
	}

	// Add path parameters that aren't already in the set
	for _, param := range pathParams {
		if !requiredSet[param] {
			existingRequired = append(existingRequired, param)
			requiredSet[param] = true
		}
	}

	return existingRequired
}

// addRequiredFieldsFromSchema adds required fields from request body schema
func addRequiredFieldsFromSchema(requestBodySchema map[string]interface{}, existingRequired []string) []string {
	if requestBodySchema == nil {
		return existingRequired
	}

	if schema, ok := requestBodySchema["schema"].(map[string]interface{}); ok {
		if required, ok := schema["required"].([]interface{}); ok {
			for _, field := range required {
				if fieldName, ok := field.(string); ok {
					existingRequired = append(existingRequired, fieldName)
				}
			}
		}
	}

	return existingRequired
}

// extractRequestBodySchema extracts schema information from request body
func extractRequestBodySchema(requestBody *openapi.RequestBody, spec *openapi.OpenAPISpec) *RequestBodyInfo {
	logger.Debug("extractRequestBodySchema called with requestBody: %+v\n", requestBody)

	if requestBody == nil {
		logger.Debug("requestBody is nil\n")
		return nil
	}

	// Resolve reference if needed
	resolvedRequestBody := spec.ResolveRequestBodyRef(requestBody)
	logger.Debug("resolvedRequestBody: %+v\n", resolvedRequestBody)

	if resolvedRequestBody == nil || resolvedRequestBody.Content == nil {
		logger.Debug("resolvedRequestBody is nil or has no content\n")
		return nil
	}

	// Look for JSON content type first
	for contentType, mediaType := range resolvedRequestBody.Content {
		if contentType == ContentTypeJSON || contentType == ContentTypeConfluentJSON {
			if mediaType.Schema != nil {
				// Resolve schema reference if needed
				resolvedSchema := spec.ResolveSchemaRef(mediaType.Schema)
				logger.Debug("Resolved schema: %+v\n", resolvedSchema)

				// Handle *Schema struct from OpenAPI parser
				if schema, ok := resolvedSchema.(*openapi.Schema); ok {
					return &RequestBodyInfo{
						Schema:      schema,
						ContentType: contentType,
					}
				}
				// fallback for map[string]interface{} (legacy)
				if schemaMap, ok := resolvedSchema.(map[string]interface{}); ok {
					return &RequestBodyInfo{
						Schema:      schemaMap,
						ContentType: contentType,
					}
				}
			}
		}
	}

	// Fallback to any available content type
	for contentType, mediaType := range resolvedRequestBody.Content {
		if mediaType.Schema != nil {
			// Resolve schema reference if needed
			resolvedSchema := spec.ResolveSchemaRef(mediaType.Schema)
			logger.Debug("Fallback resolved schema: %+v\n", resolvedSchema)

			if schema, ok := resolvedSchema.(*openapi.Schema); ok {
				return &RequestBodyInfo{
					Schema:      schema,
					ContentType: contentType,
				}
			}
			if schemaMap, ok := resolvedSchema.(map[string]interface{}); ok {
				return &RequestBodyInfo{
					Schema:      schemaMap,
					ContentType: contentType,
				}
			}
		}
	}

	return nil
}

// ExtractPathParameters extracts parameter names from OpenAPI path templates
func ExtractPathParameters(path string) []string {
	parts := strings.Split(path, "/")
	var params []string
	for _, part := range parts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			param := strings.Trim(part, "{}")
			params = append(params, param)
		}
	}
	return params
}

// ResolveResourceSchema resolves resource schema for a given action and resource
func ResolveResourceSchema(action, resource string) (map[string]interface{}, error) {
	schema, err := GetParameterSchemaForResource(action, resource)
	if err != nil {
		return nil, err
	}
	return schema, nil
}

// logDiscoveredResources logs all resources discovered from the OpenAPI spec for debugging
func logDiscoveredResources(spec *openapi.OpenAPISpec) {
	if spec == nil {
		return
	}

	resourceSet := extractResourcesFromSpec(spec)
	if len(resourceSet) > 0 {
		var resources []string
		for resource := range resourceSet {
			resources = append(resources, resource)
		}
		logger.Debug("Discovered resources from OpenAPI spec: %v\n", resources)
	} else {
		logger.Debug("No resources discovered from OpenAPI spec\n")
	}
}

// extractResourcesFromSpec dynamically extracts all resources from the OpenAPI spec
func extractResourcesFromSpec(spec *openapi.OpenAPISpec) map[string]bool {
	resourceSet := make(map[string]bool)

	if spec == nil {
		return resourceSet
	}

	// Extract resources from all paths in the spec
	for path := range spec.Paths {
		parts := strings.Split(strings.TrimPrefix(path, PathSeparator), PathSeparator)

		for _, part := range parts {
			// Skip parameters and version patterns
			if isPathParameter(part) || isVersionPattern(part) {
				continue
			}

			// If it looks like a resource (plural noun), add it
			if isPluralResourceName(part) {
				resourceSet[part] = true
			}
		}
	}

	return resourceSet
}
