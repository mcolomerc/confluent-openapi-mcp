package tools

import (
	"fmt"
	"mcolomerc/mcp-server/internal/openapi"
	"regexp"
	"strings"
)

// GenerateTools takes an OpenAPI specification and generates tools
func GenerateTools(spec openapi.OpenAPISpec) ([]Tool, error) {
	// Use semantic tool generation instead of direct mapping
	return GenerateSemanticTools(spec)
}

// extractHTTPOperations extracts all HTTP operations from a path item
func extractHTTPOperations(pathItem *openapi.PathItem) []HTTPOperation {
	var operations []HTTPOperation

	if pathItem.Get != nil {
		operations = append(operations, HTTPOperation{
			Method:    HTTPMethodGet,
			Operation: pathItem.Get,
			HasBody:   false,
		})
	}

	if pathItem.Post != nil {
		operations = append(operations, HTTPOperation{
			Method:    HTTPMethodPost,
			Operation: pathItem.Post,
			HasBody:   true,
		})
	}

	if pathItem.Put != nil {
		operations = append(operations, HTTPOperation{
			Method:    HTTPMethodPut,
			Operation: pathItem.Put,
			HasBody:   true,
		})
	}

	if pathItem.Patch != nil {
		operations = append(operations, HTTPOperation{
			Method:    HTTPMethodPatch,
			Operation: pathItem.Patch,
			HasBody:   true,
		})
	}

	if pathItem.Delete != nil {
		operations = append(operations, HTTPOperation{
			Method:    HTTPMethodDelete,
			Operation: pathItem.Delete,
			HasBody:   false,
		})
	}

	return operations
}

// createToolFromOperation creates a Tool from an HTTP operation
func createToolFromOperation(path string, httpOp HTTPOperation) (Tool, error) {
	name := getOperationName(httpOp.Operation, httpOp.Method, path)
	description := getOperationDescription(httpOp.Operation, httpOp.Method, path)

	var parameters map[string]interface{}
	if httpOp.HasBody {
		parameters = extractRequestBodyOrParameterSchema(
			httpOp.Operation.RequestBody,
			httpOp.Operation.Parameters,
			path,
		)
	} else {
		parameters = generateParameterSchema(httpOp.Operation.Parameters, path)
	}

	return Tool{
		Name:        normalizeToolName(name),
		Description: description,
		Endpoint:    fmt.Sprintf("%s %s", httpOp.Method, path),
		Parameters:  parameters,
	}, nil
}

// getOperationName gets the name for an operation
func getOperationName(operation *openapi.Operation, method, path string) string {
	if operation.Summary != "" {
		return operation.Summary
	}
	return fmt.Sprintf("%s %s", method, path)
}

// getOperationDescription gets the description for an operation
func getOperationDescription(operation *openapi.Operation, method, path string) string {
	if operation.Description != "" {
		return operation.Description
	}
	return fmt.Sprintf("Performs a %s request to %s", method, path)
}

// extractRequestBodyOrParameterSchema returns the request body schema if present, otherwise falls back to parameter schema
func extractRequestBodyOrParameterSchema(requestBody *openapi.RequestBody, params []openapi.Parameter, path string) map[string]interface{} {
	if requestBody != nil && requestBody.Content != nil {
		if schema := getSchemaFromRequestBody(requestBody); schema != nil {
			return schemaToJSONSchema(schema)
		}
	}
	return generateParameterSchema(params, path)
}

// getSchemaFromRequestBody extracts schema from request body, preferring JSON content type
func getSchemaFromRequestBody(requestBody *openapi.RequestBody) *openapi.Schema {
	// Prefer application/json
	if media, ok := requestBody.Content[ContentTypeJSON]; ok && media.Schema != nil {
		if schema, ok := media.Schema.(*openapi.Schema); ok {
			return schema
		}
	}

	// Fallback to any available content type
	for _, media := range requestBody.Content {
		if media.Schema != nil {
			if schema, ok := media.Schema.(*openapi.Schema); ok {
				return schema
			}
		}
	}

	return nil
}

// schemaToJSONSchema converts an *openapi.Schema to a map[string]interface{} (JSON Schema)
func schemaToJSONSchema(schema *openapi.Schema) map[string]interface{} {
	if schema == nil {
		return map[string]interface{}{"type": ParamTypeObject}
	}

	result := make(map[string]interface{})

	if schema.Type != "" {
		result["type"] = schema.Type
	}

	if len(schema.Properties) > 0 {
		result["properties"] = convertPropertiesToJSONSchema(schema.Properties)
	}

	if len(schema.Required) > 0 {
		result["required"] = schema.Required
	}

	if schema.Items != nil {
		result["items"] = schemaToJSONSchema(schema.Items)
	}

	return result
}

// convertPropertiesToJSONSchema converts schema properties to JSON Schema format
func convertPropertiesToJSONSchema(properties map[string]*openapi.Schema) map[string]interface{} {
	props := make(map[string]interface{})
	for k, v := range properties {
		props[k] = schemaToJSONSchema(v)
	}
	return props
}

// normalizeToolName converts a tool name to MCP-compliant format (lowercase, [a-z0-9_-] only)
func normalizeToolName(name string) string {
	if name == "" {
		return "unnamed-tool"
	}

	// Convert to lowercase and replace invalid characters
	name = strings.ToLower(name)
	name = replaceInvalidCharacters(name)
	name = cleanupHyphens(name)

	return name
}

// replaceInvalidCharacters replaces non-alphanumeric characters with hyphens
func replaceInvalidCharacters(name string) string {
	re := regexp.MustCompile(`[^a-z0-9_-]+`)
	return re.ReplaceAllString(name, "-")
}

// cleanupHyphens removes leading/trailing hyphens and collapses multiple hyphens
func cleanupHyphens(name string) string {
	name = strings.Trim(name, "-")
	re := regexp.MustCompile(`-+`)
	return re.ReplaceAllString(name, "-")
}

// generateParameterSchema converts OpenAPI parameters to MCP JSON Schema format
func generateParameterSchema(params []openapi.Parameter, path string) map[string]interface{} {
	schema := map[string]interface{}{
		"type":       ParamTypeObject,
		"properties": make(map[string]interface{}),
	}

	properties := schema["properties"].(map[string]interface{})
	var required []string

	// Add path parameters
	required = addPathParameters(properties, path, required)

	// Add parameters from OpenAPI spec
	required = addOpenAPIParameters(properties, params, required)

	if len(required) > 0 {
		schema["required"] = required
	}

	return schema
}

// addPathParameters adds path parameters to the schema
func addPathParameters(properties map[string]interface{}, path string, required []string) []string {
	pathParams := extractPathParameters(path)
	for _, pathParam := range pathParams {
		properties[pathParam] = map[string]interface{}{
			"type":        ParamTypeString,
			"description": fmt.Sprintf("Path parameter: %s", pathParam),
		}
		required = append(required, pathParam)
	}
	return required
}

// addOpenAPIParameters adds OpenAPI parameters to the schema
func addOpenAPIParameters(properties map[string]interface{}, params []openapi.Parameter, required []string) []string {
	for _, param := range params {
		paramSchema := map[string]interface{}{
			"type":        getParameterType(param.Schema),
			"description": fmt.Sprintf("Parameter: %s (in: %s)", param.Name, param.In),
		}

		properties[param.Name] = paramSchema

		if param.Required {
			required = append(required, param.Name)
		}
	}
	return required
}

// extractPathParameters extracts parameter names from OpenAPI path templates
func extractPathParameters(path string) []string {
	re := regexp.MustCompile(`\{([^}]+)\}`)
	matches := re.FindAllStringSubmatch(path, -1)

	var params []string
	for _, match := range matches {
		if len(match) > 1 {
			params = append(params, match[1])
		}
	}

	return params
}

// getParameterType converts OpenAPI schema type to JSON Schema type
func getParameterType(schema *openapi.Schema) string {
	if schema == nil {
		return ParamTypeString
	}

	switch schema.Type {
	case ParamTypeInteger, ParamTypeNumber, ParamTypeBoolean, ParamTypeArray, ParamTypeObject:
		return schema.Type
	default:
		return ParamTypeString
	}
}
