package server

import (
	"encoding/json"
	"fmt"
	"mcolomerc/mcp-server/internal/logger"
	"mcolomerc/mcp-server/internal/openapi"
	"mcolomerc/mcp-server/internal/tools"
	"strings"
)

// Tool invocation business logic and helper functions

// InvokeTool executes a tool with the given request
func (s *MCPServer) InvokeTool(req InvokeRequest) InvokeResponse {
	logger.Debug("InvokeTool called with tool=%s, arguments=%v\n", req.Tool, req.Arguments)
	var tool *tools.Tool
	for i := range s.tools {
		if s.tools[i].Name == req.Tool {
			tool = &s.tools[i]
			break
		}
	}
	if tool == nil {
		return InvokeResponse{Error: "Tool not found"}
	}
	securityType := "cloud-api-key"
	endpoint := tool.Endpoint
	_, _ = getAPICredentials(s.config, securityType, endpoint)

	// --- Begin required parameter validation and auto-translation ---
	action := tool.Name
	resource := ""

	// For semantic tools, get resource from arguments
	if action == "create" || action == "update" || action == "delete" || action == "get" || action == "list" {
		if res, ok := req.Arguments["resource"].(string); ok {
			resource = res
		}
	} else {
		// For non-semantic tools, try to extract from endpoint
		if parts := strings.Split(endpoint, " "); len(parts) == 2 {
			resource = tools.ExtractResourceFromPath(parts[1])
		}
	}

	logger.Debug("action=%s, resource=%s\n", action, resource)

	// Debug: Show required parameters for this action/resource combination
	if resource != "" && (action == "create" || action == "update" || action == "delete" || action == "get" || action == "list") {
		required, _ := tools.GetRequiredParametersForResource(action, resource)
		logger.Debug("Required parameters for %s %s: %v\n", action, resource, required)
	}

	// --- Apply default parameter values first ---
	for k, v := range req.Arguments {
		if v == nil || v == "" {
			if def := resolveDefaultParam(s.config, k, tool.Endpoint); def != "" {
				req.Arguments[k] = def
			}
		}
	}
	// Also check for missing required parameters and apply defaults
	if resource != "" && (action == "create" || action == "update" || action == "delete" || action == "get" || action == "list") {
		required, _ := tools.GetRequiredParametersForResource(action, resource)
		for _, param := range required {
			if _, ok := req.Arguments[param]; !ok {
				if def := resolveDefaultParam(s.config, param, tool.Endpoint); def != "" {
					req.Arguments[param] = def
				}
			}
		}
	}
	// --- End default parameter application ---

	// --- Begin required parameter validation and auto-translation ---
	if resource != "" && (action == "create" || action == "update" || action == "delete" || action == "get" || action == "list") {
		required, _ := tools.GetRequiredParametersForResource(action, resource)
		missing := []string{}
		translated := false

		// For semantic tools, extract parameters from nested 'parameters' object
		var paramsToCheck map[string]interface{}
		if params, ok := req.Arguments["parameters"].(map[string]interface{}); ok {
			// Merge nested parameters with top-level arguments
			paramsToCheck = make(map[string]interface{})
			for k, v := range req.Arguments {
				paramsToCheck[k] = v
			}
			for k, v := range params {
				paramsToCheck[k] = v
			}
			logger.Debug("Extracted parameters from nested object: %v\n", paramsToCheck)
		} else {
			paramsToCheck = req.Arguments
		}

		for _, param := range required {
			if _, ok := paramsToCheck[param]; !ok {
				// Check if this parameter can be resolved from defaults
				if def := resolveDefaultParam(s.config, param, tool.Endpoint); def != "" {
					paramsToCheck[param] = def
					logger.Debug("Auto-resolved parameter %s from config: %s\n", param, def)
					continue
				}
				// If param contains 'name' and 'name' is present, auto-translate
				if strings.Contains(param, "name") && paramsToCheck["name"] != nil {
					paramsToCheck[param] = paramsToCheck["name"]
					translated = true
					logger.Debug("Auto-translated 'name' to parameter %s: %v\n", param, paramsToCheck["name"])
					continue
				}
				missing = append(missing, param)
			}
		}

		// Update req.Arguments with the merged parameters
		req.Arguments = paramsToCheck

		if len(missing) > 0 {
			logger.Debug("Missing required parameters for %s %s: %v\n", action, resource, missing)
			logger.Debug("Available arguments: %v\n", req.Arguments)
			return InvokeResponse{
				Result: map[string]interface{}{
					"status":         "missing_required_params",
					"requiredParams": missing,
					"message":        "Please provide the following required parameters.",
				},
			}
		}
		if translated {
			return InvokeResponse{Result: map[string]interface{}{
				"info":      "Parameter 'name' was auto-translated to the required parameter.",
				"arguments": req.Arguments,
			}}
		}
	}
	// --- End required parameter validation and auto-translation ---

	// --- Build request body if schema is present ---
	var requestBody interface{} = nil
	if resource != "" && (action == "create" || action == "update") {
		logger.Debug("Starting request body build for action=%s resource=%s\n", action, resource)
		mapping, _ := tools.GetEndpointMapping(action, resource)
		logger.Debug("Building request body for %s %s, schema available: %v\n", action, resource, mapping.RequestBodySchema != nil)
		logger.Debug("Building request body for %s %s, schema available: %v\n", action, resource, mapping.RequestBodySchema != nil)
		if mapping.RequestBodySchema != nil {
			// For semantic tools, parameters can be under req.Arguments["parameters"] or directly in req.Arguments
			var dataArgs map[string]interface{}
			if params, ok := req.Arguments["parameters"].(map[string]interface{}); ok {
				dataArgs = params
				logger.Debug("Found parameters under req.Arguments[parameters]: %v\n", dataArgs)
				logger.Debug("Found parameters under req.Arguments[parameters]: %v\n", dataArgs)
			} else {
				// Fallback to using req.Arguments directly and try to map them to schema properties
				dataArgs = req.Arguments
				logger.Debug("Using req.Arguments directly, attempting schema mapping: %v\n", dataArgs)
				logger.Debug("Using req.Arguments directly, attempting schema mapping: %v\n", dataArgs)

				// Try to intelligently map common argument names to schema properties
				if schema, ok := mapping.RequestBodySchema["schema"].(*openapi.Schema); ok && schema != nil {
					mappedArgs := make(map[string]interface{})

					// Get schema property names for debugging
					schemaProps := getSchemaPropertyNames(schema)
					logger.Debug("Schema properties available: %v\n", schemaProps)

					// Smart mapping rules for common parameters
					for argKey, argValue := range dataArgs {
						if argKey == "resource" {
							continue // Skip the resource parameter
						}

						// Map common argument names to schema properties
						mapped := false
						for _, prop := range schemaProps {
							if mapArgumentToProperty(argKey, prop) {
								mappedArgs[prop] = argValue
								mapped = true
								logger.Debug("Mapped argument '%s' to schema property '%s'\n", argKey, prop)
								break
							}
						}

						if !mapped {
							// If no mapping found, use the original key
							mappedArgs[argKey] = argValue
						}
					}

					dataArgs = mappedArgs
					logger.Debug("Final mapped arguments: %v\n", dataArgs)
				}
			}

			// Try to get schema as *openapi.Schema first
			logger.Debug("Schema type before assertion: %T\n", mapping.RequestBodySchema["schema"])
			if schema, ok := mapping.RequestBodySchema["schema"].(*openapi.Schema); ok && schema != nil {
				requestBody = buildRequestBodyFromSchema(schema, dataArgs)
				logger.Debug("Built request body from Schema struct: %v\n", requestBody)
				logger.Debug("Built request body from Schema struct: %v\n", requestBody)
			} else if schemaMap, ok := mapping.RequestBodySchema["schema"].(map[string]interface{}); ok && schemaMap != nil {
				// Handle resolved schema as map - but this shouldn't happen anymore since we resolve to *openapi.Schema
				logger.Debug("Using schema map path, map has %d keys\n", len(schemaMap))
				requestBody = buildRequestBodyFromSchemaMap(schemaMap, dataArgs)
				logger.Debug("Built request body from schema map: %v\n", requestBody)
				logger.Debug("Built request body from schema map: %v\n", requestBody)
			} else {
				logger.Debug("Schema type: %T, value: %v\n", mapping.RequestBodySchema["schema"], mapping.RequestBodySchema["schema"])
				logger.Debug("Schema conversion failed or schema is nil. Type: %T\n", mapping.RequestBodySchema["schema"])
				// Additional debug: check if it's a converted schema
				if mapping.RequestBodySchema["schema"] != nil {
					logger.Debug("Raw schema value: %+v\n", mapping.RequestBodySchema["schema"])
				}
			}
		} else {
			logger.Debug("No request body schema found for %s %s\n", action, resource)
		}
	}
	// --- End request body build ---

	// --- Actually call the API if this is a semantic tool ---
	if resource != "" {
		mapping, _ := tools.GetEndpointMapping(action, resource)
		apiPath := tools.BuildAPIPath(mapping.PathPattern, req.Arguments)
		logger.Debug("About to call API with method=%s, path=%s, parameters=%v, requestBody=%#v\n", mapping.Method, apiPath, req.Arguments, requestBody)
		result, err := ExecuteAPICall(s.config, s.spec, mapping.Method, apiPath, req.Arguments, requestBody)
		if err != nil {
			return InvokeResponse{Error: err.Error()}
		}
		return InvokeResponse{Result: result}
	}
	// fallback: return error for non-semantic tool
	return InvokeResponse{Error: "Invalid or unsupported tool invocation"}
}

// Helper functions for tool invocation

// mapArgumentToProperty maps common argument names to schema property names
func mapArgumentToProperty(argName, propName string) bool {
	// Direct match
	if argName == propName {
		return true
	}

	// Common mappings for topic creation
	mappings := map[string][]string{
		"name":        {"topic_name", "display_name", "name"},
		"partitions":  {"partitions_count", "partition_count"},
		"replication": {"replication_factor"},
	}

	if targets, ok := mappings[argName]; ok {
		for _, target := range targets {
			if target == propName {
				return true
			}
		}
	}

	return false
}

// Helper functions for building request bodies and schema handling

// getSchemaPropertyNames extracts property names from an OpenAPI schema
func getSchemaPropertyNames(schema *openapi.Schema) []string {
	var names []string
	if schema != nil && schema.Properties != nil {
		for name := range schema.Properties {
			names = append(names, name)
		}
	}
	return names
}

// buildRequestBodyFromSchema builds a request body from the OpenAPI schema and arguments
func buildRequestBodyFromSchema(schema *openapi.Schema, args map[string]interface{}) map[string]interface{} {
	requestBody := make(map[string]interface{})

	if schema == nil || schema.Properties == nil {
		return requestBody
	}

	// Map arguments to schema properties
	for propName, propSchema := range schema.Properties {
		if value, exists := args[propName]; exists {
			// Handle different property types
			if propSchema.Type == PropertyTypeArray && propName == ParamConfigs {
				// Special handling for configs arrays
				requestBody[propName] = transformConfigsParameter(value)
			} else {
				requestBody[propName] = value
			}
		}
	}

	return requestBody
}

// buildRequestBodyFromSchemaMap builds a request body from a resolved schema map and arguments
func buildRequestBodyFromSchemaMap(schemaMap map[string]interface{}, args map[string]interface{}) map[string]interface{} {
	requestBody := make(map[string]interface{})

	// Extract properties from schema map
	if properties, ok := schemaMap["properties"].(map[string]interface{}); ok {
		for propName := range properties {
			if value, exists := args[propName]; exists {
				if propName == ParamConfigs {
					// Special handling for configs
					requestBody[propName] = transformConfigsParameter(value)
				} else {
					requestBody[propName] = value
				}
			}
		}
	}

	return requestBody
}

// getMapKeys returns the keys of a map[string]interface{}
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// transformConfigsParameter ensures configs parameter is in the correct array format
func transformConfigsParameter(configs interface{}) interface{} {
	if configs == nil {
		return nil
	}

	// If it's already an array, return as is
	if configArray, ok := configs.([]interface{}); ok {
		return configArray
	}

	// If it's a map, convert to array format expected by API
	if configMap, ok := configs.(map[string]interface{}); ok {
		var configArray []map[string]interface{}
		for key, value := range configMap {
			configArray = append(configArray, map[string]interface{}{
				"name":  key,
				"value": fmt.Sprintf("%v", value),
			})
		}
		return configArray
	}

	// If it's a string (JSON), try to parse it
	if configStr, ok := configs.(string); ok {
		var parsed interface{}
		if err := json.Unmarshal([]byte(configStr), &parsed); err == nil {
			return transformConfigsParameter(parsed)
		}
	}

	// Return as is if we can't transform it
	return configs
}
