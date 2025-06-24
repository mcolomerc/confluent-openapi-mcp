package resource

import (
	"fmt"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// ConvertToMCPResources converts API response data to MCP resource objects
func (m *Manager) ConvertToMCPResources(resourceType string, apiResult interface{}) ([]mcp.Resource, error) {
	var resources []mcp.Resource

	// Try to extract items from the API response
	// This handles common patterns like {"data": [...]} or direct arrays
	var items []interface{}

	if resultMap, ok := apiResult.(map[string]interface{}); ok {
		// Check for common array field names
		arrayFields := append(CommonArrayFields, resourceType)
		for _, field := range arrayFields {
			if fieldValue, exists := resultMap[field]; exists {
				if itemsArray, ok := fieldValue.([]interface{}); ok {
					items = itemsArray
					break
				}
			}
		}
		// If no array field found, treat the entire object as a single item
		if len(items) == 0 {
			items = []interface{}{apiResult}
		}
	} else if itemsArray, ok := apiResult.([]interface{}); ok {
		// Direct array response
		items = itemsArray
	} else {
		// Single item response
		items = []interface{}{apiResult}
	}

	// Convert each item to an MCP resource
	for i, item := range items {
		if itemMap, ok := item.(map[string]interface{}); ok {
			resource := m.convertItemToMCPResource(resourceType, itemMap, i)
			resources = append(resources, resource)
		}
	}

	fmt.Fprintf(os.Stderr, "Converted %d %s items to MCP resources\n", len(resources), resourceType)
	return resources, nil
}

// convertItemToMCPResource converts a single API item to an MCP resource
func (m *Manager) convertItemToMCPResource(resourceType string, item map[string]interface{}, index int) mcp.Resource {
	// Try to get a meaningful identifier and name for the resource
	var id, name, description string

	// Try common ID field names
	for _, field := range CommonIDFields {
		if value, exists := item[field]; exists {
			if strValue, ok := value.(string); ok && strValue != "" {
				if id == "" {
					id = strValue
				}
				if name == "" {
					name = strValue
				}
			}
		}
	}

	// Fallback to index if no ID found
	if id == "" {
		id = fmt.Sprintf("%s-%d", resourceType, index)
	}
	if name == "" {
		name = id
	}

	// Try to get a description
	for _, field := range CommonDescriptionFields {
		if value, exists := item[field]; exists {
			if strValue, ok := value.(string); ok && strValue != "" {
				description = strValue
				break
			}
		}
	}

	if description == "" {
		description = fmt.Sprintf("%s resource: %s", strings.Title(resourceType), name)
	}

	// Create the URI for this resource
	uri := fmt.Sprintf("%s%s%s%s", ConfluentURIScheme, resourceType, URIPathSeparator, id)

	return mcp.Resource{
		URI:         uri,
		Name:        name,
		Description: description,
		MIMEType:    "application/json",
	}
}
