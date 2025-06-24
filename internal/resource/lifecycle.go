package resource

import (
	"fmt"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// HandleResourceCreation registers a newly created resource with the MCP server
func (m *Manager) HandleResourceCreation(mcpServer *server.MCPServer, args map[string]interface{}, result interface{}) {
	// Extract resource type from arguments
	resourceType, ok := args["resource"].(string)
	if !ok {
		fmt.Fprintf(os.Stderr, "Warning: Could not extract resource type from creation arguments\n")
		return
	}

	// Try to extract the created resource information from the result
	resource, err := m.extractResourceFromCreationResult(resourceType, result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not extract resource info from creation result: %v\n", err)
		return
	}

	// Register the new resource with the MCP server
	handler := m.CreateResourceReadHandler(resourceType)
	mcpServer.AddResource(resource, handler)

	fmt.Fprintf(os.Stderr, "Auto-registered new resource: %s (%s)\n", resource.Name, resource.URI)
}

// extractResourceFromCreationResult extracts resource information from a creation API response
func (m *Manager) extractResourceFromCreationResult(resourceType string, result interface{}) (mcp.Resource, error) {
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return mcp.Resource{}, fmt.Errorf("result is not a map")
	}

	// Extract resource identifier and name based on resource type
	var id, name string

	switch resourceType {
	case "topics":
		if topicName, exists := resultMap["topic_name"]; exists {
			if nameStr, ok := topicName.(string); ok {
				id = nameStr
				name = nameStr
			}
		}
	case "connectors":
		if connectorName, exists := resultMap["name"]; exists {
			if nameStr, ok := connectorName.(string); ok {
				id = nameStr
				name = nameStr
			}
		}
	case "service-accounts":
		if saId, exists := resultMap["id"]; exists {
			if idStr, ok := saId.(string); ok {
				id = idStr
			}
		}
		if displayName, exists := resultMap["display_name"]; exists {
			if nameStr, ok := displayName.(string); ok {
				name = nameStr
			}
		}
		if name == "" {
			name = id
		}
	default:
		// Generic extraction for other resource types
		for _, field := range CommonIDFields {
			if value, exists := resultMap[field]; exists {
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
	}

	if id == "" {
		return mcp.Resource{}, fmt.Errorf("could not extract resource ID from result")
	}

	if name == "" {
		name = id
	}

	description := fmt.Sprintf("Auto-registered %s resource: %s", strings.Title(resourceType), name)

	// Create the URI for this resource
	uri := fmt.Sprintf("%s%s%s%s", ConfluentURIScheme, resourceType, URIPathSeparator, id)

	return mcp.Resource{
		URI:         uri,
		Name:        name,
		Description: description,
		MIMEType:    "application/json",
	}, nil
}

// HandleResourceDeletion unregisters a deleted resource from the MCP server
func (m *Manager) HandleResourceDeletion(args map[string]interface{}) {
	// Extract resource type from arguments
	resourceType, ok := args["resource"].(string)
	if !ok {
		fmt.Fprintf(os.Stderr, "Warning: Could not extract resource type from deletion arguments\n")
		return
	}

	// Extract the resource identifier from the arguments
	resourceID := m.extractResourceIDFromDeletionArgs(resourceType, args)
	if resourceID == "" {
		fmt.Fprintf(os.Stderr, "Warning: Could not extract resource ID from deletion arguments\n")
		return
	}

	// Create the URI for the deleted resource
	uri := fmt.Sprintf("%s%s%s%s", ConfluentURIScheme, resourceType, URIPathSeparator, resourceID)

	// Note: The MCP library doesn't appear to have a RemoveResource method,
	// so we log the deletion for now. In a real implementation, you might:
	// 1. Maintain your own registry of resources
	// 2. Use resource notifications to inform clients
	// 3. Return appropriate errors when clients try to access deleted resources

	fmt.Fprintf(os.Stderr, "Resource deleted (manual cleanup may be needed): %s (%s)\n", resourceID, uri)
}

// extractResourceIDFromDeletionArgs extracts the resource identifier from deletion arguments
func (m *Manager) extractResourceIDFromDeletionArgs(resourceType string, args map[string]interface{}) string {
	// Check resource-specific mappings first
	if fieldNames, exists := ResourceTypeIDMappings[resourceType]; exists {
		for _, fieldName := range fieldNames {
			if value, exists := args[fieldName]; exists {
				if strValue, ok := value.(string); ok {
					return strValue
				}
			}
		}
	}

	// Generic fallback - try common identifier field names
	for _, pattern := range GenericIDFieldPatterns {
		// Try exact field name
		if value, exists := args[pattern]; exists {
			if strValue, ok := value.(string); ok {
				return strValue
			}
		}

		// Try with resource type prefix
		fieldName := resourceType + pattern
		if value, exists := args[fieldName]; exists {
			if strValue, ok := value.(string); ok {
				return strValue
			}
		}
	}

	return ""
}
