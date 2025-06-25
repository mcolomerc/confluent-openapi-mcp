package openapi

import (
	"encoding/json"
	"fmt"
	"io"
	"mcolomerc/mcp-server/internal/logger"
	"net/http"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// OpenAPISpec represents the structure of the OpenAPI specification.
type OpenAPISpec struct {
	OpenAPI    string                `json:"openapi"`
	Info       Info                  `json:"info"`
	Paths      map[string]PathItem   `json:"paths"`
	Security   []map[string][]string `json:"security,omitempty"`
	Components *Components           `json:"components,omitempty"`
}

// Components holds reusable components, including security schemes.
type Components struct {
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`
	RequestBodies   map[string]RequestBody    `json:"requestBodies,omitempty"`
	Schemas         map[string]Schema         `json:"schemas,omitempty"`
	// ... add other component fields as needed ...
}

// SecurityScheme describes a security scheme definition.
type SecurityScheme struct {
	Type             string `json:"type"`
	Description      string `json:"description,omitempty"`
	Name             string `json:"name,omitempty"`
	In               string `json:"in,omitempty"`
	Scheme           string `json:"scheme,omitempty"`
	BearerFormat     string `json:"bearerFormat,omitempty"`
	OpenIdConnectUrl string `json:"openIdConnectUrl,omitempty"`
	// ... add other fields as needed ...
}

// Info contains metadata about the API.
type Info struct {
	Title   string `json:"title"`
	Version string `json:"version"`
}

// PathItem describes the operations available on a single path.
type PathItem struct {
	Get    *Operation `json:"get,omitempty"`
	Post   *Operation `json:"post,omitempty"`
	Put    *Operation `json:"put,omitempty"`
	Delete *Operation `json:"delete,omitempty"`
	Patch  *Operation `json:"patch,omitempty"`
	// Add other HTTP methods as needed
}

// Operation describes a single API operation.
type Operation struct {
	Summary     string                `json:"summary"`
	Description string                `json:"description"`
	Parameters  []Parameter           `json:"parameters,omitempty"`
	RequestBody *RequestBody          `json:"requestBody,omitempty"`
	Security    []map[string][]string `json:"security,omitempty"`
}

// Parameter describes a single parameter for an operation.
type Parameter struct {
	Name     string  `json:"name"`
	In       string  `json:"in"` // e.g., "query", "header", "path", "cookie"
	Required bool    `json:"required"`
	Schema   *Schema `json:"schema,omitempty"`
}

// Schema describes the structure of a parameter's schema.
type Schema struct {
	Type       string             `json:"type"`
	Properties map[string]*Schema `json:"properties,omitempty"`
	Required   []string           `json:"required,omitempty"`
	Items      *Schema            `json:"items,omitempty"`
}

// RequestBody describes the request body of an operation.
type RequestBody struct {
	Ref     string               `json:"$ref,omitempty"`
	Content map[string]MediaType `json:"content,omitempty"`
}

// MediaType describes a single media type.
type MediaType struct {
	Schema interface{} `json:"schema,omitempty"`
}

// ParseOpenAPISpec reads and parses the OpenAPI specification from a file.
func ParseOpenAPISpec(filePath string) (*OpenAPISpec, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var spec OpenAPISpec
	if err := json.Unmarshal(bytes, &spec); err != nil {
		return nil, err
	}

	return &spec, nil
}

// ParseOpenAPISpecBytes parses the OpenAPI spec from a byte slice.
func ParseOpenAPISpecBytes(data []byte) (*OpenAPISpec, error) {
	var spec OpenAPISpec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, err
	}
	return &spec, nil
}

// ParseOpenAPISpecYAML parses the OpenAPI spec from a YAML file.
func ParseOpenAPISpecYAML(filename string) (*OpenAPISpec, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var spec OpenAPISpec
	if err := yaml.Unmarshal(bytes, &spec); err != nil {
		return nil, err
	}

	return &spec, nil
}

// ParseOpenAPISpecBytesYAML parses the OpenAPI spec from a YAML byte slice.
func ParseOpenAPISpecBytesYAML(data []byte) (*OpenAPISpec, error) {
	var spec OpenAPISpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, err
	}
	return &spec, nil
}

// LoadSpec loads an OpenAPI spec from a file path or URL, or from the default if empty.
func LoadSpec() (*OpenAPISpec, error) {
	specPath := os.Getenv("OPENAPI_SPEC_URL")
	if specPath == "" {
		specPath = "api-spec/confluent-apispec.json"
	}

	if strings.HasPrefix(specPath, "http://") || strings.HasPrefix(specPath, "https://") {
		resp, err := http.Get(specPath)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch OpenAPI spec from remote: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("failed to fetch OpenAPI spec: HTTP %d", resp.StatusCode)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read OpenAPI spec body: %w", err)
		}
		return ParseOpenAPISpecBytes(body)
	}
	return ParseOpenAPISpec(specPath)
}

// LoadTelemetrySpec loads the Confluent Telemetry OpenAPI spec from a file path or URL.
func LoadTelemetrySpec() (*OpenAPISpec, error) {
	specPath := os.Getenv("TELEMETRY_OPENAPI_SPEC_URL")
	if specPath == "" {
		specPath = "api-spec/confluent-telemetry-apispec.yaml"
	}

	if strings.HasPrefix(specPath, "http://") || strings.HasPrefix(specPath, "https://") {
		resp, err := http.Get(specPath)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch Telemetry OpenAPI spec from remote: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("failed to fetch Telemetry OpenAPI spec: HTTP %d", resp.StatusCode)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read Telemetry OpenAPI spec body: %w", err)
		}
		return ParseOpenAPISpecBytesYAML(body)
	}
	
	// Determine if it's YAML or JSON based on file extension
	if strings.HasSuffix(strings.ToLower(specPath), ".yaml") || strings.HasSuffix(strings.ToLower(specPath), ".yml") {
		return ParseOpenAPISpecYAML(specPath)
	}
	return ParseOpenAPISpec(specPath)
}

// LoadBothSpecs loads both the main Confluent API spec and the Telemetry API spec.
func LoadBothSpecs() (*OpenAPISpec, *OpenAPISpec, error) {
	mainSpec, err := LoadSpec()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load main OpenAPI spec: %w", err)
	}

	telemetrySpec, err := LoadTelemetrySpec()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load telemetry OpenAPI spec: %w", err)
	}

	return mainSpec, telemetrySpec, nil
}

// ResolveRequestBodyRef resolves a RequestBody reference if needed
func (spec *OpenAPISpec) ResolveRequestBodyRef(requestBody *RequestBody) *RequestBody {
	logger.Debug("ResolveRequestBodyRef called with requestBody: %+v\n", requestBody)

	if requestBody == nil {
		logger.Debug("requestBody is nil\n")
		return nil
	}

	// If this is a reference, resolve it
	if requestBody.Ref != "" {
		logger.Debug("Found reference: %s\n", requestBody.Ref)
		// Extract the reference path (e.g., "#/components/requestBodies/CreateTopicRequest")
		if strings.HasPrefix(requestBody.Ref, "#/components/requestBodies/") {
			refName := strings.TrimPrefix(requestBody.Ref, "#/components/requestBodies/")
			logger.Debug("Looking for requestBody component: %s\n", refName)

			if spec.Components != nil && spec.Components.RequestBodies != nil {
				logger.Debug("Components found, requestBodies count: %d\n", len(spec.Components.RequestBodies))
				if resolvedRequestBody, exists := spec.Components.RequestBodies[refName]; exists {
					logger.Debug("Found resolved requestBody: %+v\n", resolvedRequestBody)
					return &resolvedRequestBody
				} else {
					logger.Debug("RequestBody component '%s' not found\n", refName)
				}
			} else {
				logger.Debug("Components or RequestBodies is nil\n")
			}
		}
		logger.Debug("Reference not found, returning original requestBody\n")
		return requestBody // Reference not found, return original
	}

	// Not a reference, return as-is
	logger.Debug("Not a reference, returning original requestBody\n")
	return requestBody
}

// ResolveSchemaRef resolves a schema reference if needed
func (spec *OpenAPISpec) ResolveSchemaRef(schema interface{}) interface{} {
	logger.Debug("ResolveSchemaRef called with schema: %+v\n", schema)

	if schema == nil {
		return nil
	}

	// Check if it's a map with a $ref
	if schemaMap, ok := schema.(map[string]interface{}); ok {
		if ref, hasRef := schemaMap["$ref"]; hasRef {
			if refStr, ok := ref.(string); ok {
				logger.Debug("Found schema reference: %s\n", refStr)
				// Extract the reference path (e.g., "#/components/schemas/CreateTopicRequestData")
				if strings.HasPrefix(refStr, "#/components/schemas/") {
					refName := strings.TrimPrefix(refStr, "#/components/schemas/")
					logger.Debug("Looking for schema component: %s\n", refName)

					if spec.Components != nil && spec.Components.Schemas != nil {
						logger.Debug("Components found, schemas count: %d\n", len(spec.Components.Schemas))
						if resolvedSchema, exists := spec.Components.Schemas[refName]; exists {
							logger.Debug("Found resolved schema: %+v\n", resolvedSchema)
							// Convert Schema struct to map for consistency
							return map[string]interface{}{
								"type":       resolvedSchema.Type,
								"properties": resolvedSchema.Properties,
								"required":   resolvedSchema.Required,
								"items":      resolvedSchema.Items,
							}
						} else {
							logger.Debug("Schema component '%s' not found\n", refName)
						}
					} else {
						logger.Debug("Components or Schemas is nil\n")
					}
				}
				logger.Debug("Schema reference not found, returning original\n")
				return schema
			}
		}
	}

	// Not a reference, return as-is
	logger.Debug("Not a schema reference, returning original\n")
	return schema
}

// GetSecurityTypeForEndpoint determines the security type for a given HTTP method and path
// by looking up the endpoint in the OpenAPI specification
func (spec *OpenAPISpec) GetSecurityTypeForEndpoint(method, path string) string {
	if spec == nil || spec.Paths == nil {
		return "cloud-api-key" // Default fallback
	}

	// Find the path item (exact match first, then pattern matching)
	pathItem := spec.findPathItem(path)
	if pathItem == nil {
		// Fall back to global security if path not found
		if spec.Security != nil && len(spec.Security) > 0 {
			return extractSecurityType(spec.Security)
		}
		return "" // No path found, no security
	}

	// Get the operation for the HTTP method
	var operation *Operation
	switch strings.ToUpper(method) {
	case "GET":
		operation = pathItem.Get
	case "POST":
		operation = pathItem.Post
	case "PUT":
		operation = pathItem.Put
	case "DELETE":
		operation = pathItem.Delete
	case "PATCH":
		operation = pathItem.Patch
	}

	if operation == nil {
		// Fall back to global security if operation not found but path exists
		if spec.Security != nil && len(spec.Security) > 0 {
			return extractSecurityType(spec.Security)
		}
		return "" // No operation found, no security
	}

	// Check operation-level security first
	if operation.Security != nil && len(operation.Security) > 0 {
		return extractSecurityType(operation.Security)
	}

	// Fall back to global security if no operation-level security
	if spec.Security != nil && len(spec.Security) > 0 {
		return extractSecurityType(spec.Security)
	}

	return "" // No security found
}

// findPathItem finds the path item for a given path, supporting OpenAPI path templates
func (spec *OpenAPISpec) findPathItem(path string) *PathItem {
	// Try exact match first
	if pathItem, exists := spec.Paths[path]; exists {
		return &pathItem
	}

	// Try pattern matching for OpenAPI path templates
	for specPath, pathItem := range spec.Paths {
		if matchesPathPattern(path, specPath) {
			return &pathItem
		}
	}

	return nil
}

// matchesPathPattern checks if a request path matches an OpenAPI path pattern
func matchesPathPattern(requestPath, specPath string) bool {
	// Simple pattern matching for OpenAPI path templates like /topics/{topic-name}
	requestParts := strings.Split(strings.Trim(requestPath, "/"), "/")
	specParts := strings.Split(strings.Trim(specPath, "/"), "/")

	if len(requestParts) != len(specParts) {
		return false
	}

	for i, specPart := range specParts {
		// If spec part is a path parameter (enclosed in braces), it matches any value
		if strings.HasPrefix(specPart, "{") && strings.HasSuffix(specPart, "}") {
			continue
		}
		// Otherwise, require exact match
		if requestParts[i] != specPart {
			return false
		}
	}

	return true
}

// extractSecurityType extracts the security type from a security requirement array
func extractSecurityType(securityRequirements []map[string][]string) string {
	if len(securityRequirements) == 0 {
		return "cloud-api-key"
	}

	// Check the first security requirement (most OpenAPI specs have only one)
	firstReq := securityRequirements[0]

	// Return the first security scheme name found
	for secType := range firstReq {
		return secType
	}

	return "cloud-api-key" // Default fallback
}
