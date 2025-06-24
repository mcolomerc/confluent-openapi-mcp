package resource

import (
	"context"
	"encoding/json"
	"fmt"
	"mcolomerc/mcp-server/internal/tools"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Manager handles resource discovery, registration, and lifecycle management
type Manager struct {
	invoker ToolInvoker // Interface for invoking tools
}

// ToolInvoker interface for invoking tools (allows for dependency injection)
type ToolInvoker interface {
	InvokeTool(req InvokeRequest) InvokeResponse
}

// NewManager creates a new resource manager
func NewManager(invoker ToolInvoker) *Manager {
	return &Manager{
		invoker: invoker,
	}
}

// DiscoverAndRegisterResources dynamically discovers and registers individual resource instances
func (m *Manager) DiscoverAndRegisterResources(mcpServer *server.MCPServer) {
	// Check if resource discovery is disabled via environment variable
	if os.Getenv("DISABLE_RESOURCE_DISCOVERY") == "true" {
		fmt.Fprintf(os.Stderr, "Resource discovery disabled via DISABLE_RESOURCE_DISCOVERY environment variable\n")
		return
	}

	if tools.GlobalSemanticRegistry == nil {
		fmt.Fprintf(os.Stderr, "Semantic registry not initialized, skipping resource discovery\n")
		return
	}

	// Get all resources that support the 'list' action
	listResources, exists := tools.GlobalSemanticRegistry.Mappings[tools.ActionList]
	if !exists || len(listResources) == 0 {
		fmt.Fprintf(os.Stderr, "No resources support 'list' action\n")
		return
	}

	fmt.Fprintf(os.Stderr, "Discovering and registering resources for %d resource types\n", len(listResources))

	// For each resource type, get the list of instances and register them
	for resourceType := range listResources {
		fmt.Fprintf(os.Stderr, "Discovering %s resources...\n", resourceType)

		resources, err := m.getResourceInstancesOfType(resourceType)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to discover %s resources: %v\n", resourceType, err)
			continue
		}

		// Register each discovered resource instance
		for _, resource := range resources {
			handler := m.CreateResourceReadHandler(resourceType)
			mcpServer.AddResource(resource, handler)
			fmt.Fprintf(os.Stderr, "Registered resource: %s (%s)\n", resource.Name, resource.URI)
		}
	}
}

// CreateResourceReadHandler creates a read handler for a specific resource type
func (m *Manager) CreateResourceReadHandler(resourceType string) func(context.Context, mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	return func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		return m.HandleResourceRead(ctx, request)
	}
}

// getResourceInstancesOfType gets all instances of a specific resource type
func (m *Manager) getResourceInstancesOfType(resourceType string) ([]mcp.Resource, error) {
	// Skip resource discovery for certain resource types that don't support general listing
	skipDiscovery := []string{"tags", "businessmetadatadefs"} // Add problematic resources
	for _, skip := range skipDiscovery {
		if resourceType == skip {
			fmt.Fprintf(os.Stderr, "Skipping discovery for %s (requires specific entity parameters)\n", resourceType)
			// Return a placeholder resource to indicate the resource type is available
			return []mcp.Resource{
				{
					URI:         fmt.Sprintf("confluent://%s/%s-placeholder", resourceType, resourceType),
					Name:        fmt.Sprintf("%s-placeholder", resourceType),
					Description: fmt.Sprintf("Placeholder for %s resource type - use tools to interact", resourceType),
					MIMEType:    "application/json",
				},
			}, nil
		}
	}

	// Use the 'list' tool to get all instances of this resource type
	invokeReq := InvokeRequest{
		Tool: tools.ActionList,
		Arguments: map[string]interface{}{
			"resource": resourceType,
		},
	}

	resp := m.invoker.InvokeTool(invokeReq)
	if resp.Error != "" {
		return nil, fmt.Errorf("failed to list %s: %s", resourceType, resp.Error)
	}

	// Convert the API response to MCP resources
	return m.ConvertToMCPResources(resourceType, resp.Result)
}

// HandleResourceRead handles reading a specific resource
func (m *Manager) HandleResourceRead(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	// Extract resource type and ID from URI (e.g., "confluent://topics/my-topic")
	uri := request.Params.URI
	if !strings.HasPrefix(uri, ConfluentURIScheme) {
		return nil, fmt.Errorf("unsupported resource URI scheme: %s", uri)
	}

	// Parse URI: confluent://resourceType/resourceId
	parts := strings.Split(strings.TrimPrefix(uri, ConfluentURIScheme), URIPathSeparator)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid resource URI format: %s", uri)
	}

	resourceType := parts[0]
	resourceID := parts[1]

	// Check if this resource type supports 'get' action
	if tools.GlobalSemanticRegistry == nil {
		return nil, fmt.Errorf("semantic registry not initialized")
	}

	getResources, exists := tools.GlobalSemanticRegistry.Mappings[tools.ActionGet]
	if !exists {
		return nil, fmt.Errorf("no resources support 'get' action")
	}

	if _, supported := getResources[resourceType]; !supported {
		return nil, fmt.Errorf("resource type '%s' does not support 'get' action", resourceType)
	}

	// Use the 'get' tool to fetch this specific resource
	invokeReq := InvokeRequest{
		Tool: tools.ActionGet,
		Arguments: map[string]interface{}{
			"resource": resourceType,
			// Add the resource identifier as a parameter
			strings.TrimSuffix(resourceType, "s") + "Id": resourceID, // topics -> topicId
		},
	}

	resp := m.invoker.InvokeTool(invokeReq)
	if resp.Error != "" {
		return nil, fmt.Errorf("failed to get %s %s: %s", resourceType, resourceID, resp.Error)
	}

	// Convert the API response to resource contents
	resultJSON, err := json.Marshal(resp.Result)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize %s data: %v", resourceType, err)
	}

	return []mcp.ResourceContents{mcp.TextResourceContents{
		URI:      uri,
		MIMEType: "application/json",
		Text:     string(resultJSON),
	}}, nil
}
