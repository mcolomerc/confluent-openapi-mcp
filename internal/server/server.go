package server

import (
	"context"
	"encoding/json"
	"fmt"
	"mcolomerc/mcp-server/internal/config"
	"mcolomerc/mcp-server/internal/monitoring"
	"mcolomerc/mcp-server/internal/openapi"
	"mcolomerc/mcp-server/internal/prompts"
	"mcolomerc/mcp-server/internal/resource"
	"mcolomerc/mcp-server/internal/tools"
	"net/http"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// MCPServer wraps the library's MCP server with our business logic
type MCPServer struct {
	tools           []tools.Tool
	config          *config.Config
	spec            *openapi.OpenAPISpec
	telemetrySpec   *openapi.OpenAPISpec
	promptManager   *prompts.PromptManager
	mcpServer       *server.MCPServer   // Core MCP server from library
	resourceManager *resource.Manager   // Resource management
	monitor         *monitoring.Monitor // Resource monitoring
}

// NewCompositeServer creates an MCPServer with provided config, main spec, telemetry spec and semanticTools
func NewCompositeServer(cfg *config.Config, spec *openapi.OpenAPISpec, telemetrySpec *openapi.OpenAPISpec, semanticTools []tools.Tool) *MCPServer {
	// Initialize prompt manager
	promptManager := prompts.NewPromptManager(cfg.PromptsFolder, cfg)

	// Load prompts (ignore errors for now, just log them)
	if err := promptManager.LoadPrompts(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to load prompts: %v\n", err)
	} else {
		loadedPrompts := promptManager.GetPrompts()
		fmt.Fprintf(os.Stderr, "Successfully loaded %d prompts: ", len(loadedPrompts))
		for _, p := range loadedPrompts {
			fmt.Fprintf(os.Stderr, "%s ", p.Name)
		}
		fmt.Fprintf(os.Stderr, "\n")
	}

	// Create the core MCP server from the library
	mcpServer := server.NewMCPServer("go-openapi-mcp", "0.1.0",
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(true, false), // Enable resource listing, no notifications yet
		server.WithPromptCapabilities(true),
		server.WithLogging(),
	)

	// Create our composite server
	compositeServer := &MCPServer{
		tools:         semanticTools,
		config:        cfg,
		spec:          spec,
		telemetrySpec: telemetrySpec,
		promptManager: promptManager,
		mcpServer:     mcpServer,
	}

	// Create the resource manager
	compositeServer.resourceManager = resource.NewManager(compositeServer)

	// Register semantic tools with the MCP server
	for _, tool := range semanticTools {
		mcpTool := convertToMCPTool(tool)
		mcpServer.AddTool(mcpTool, compositeServer.createToolHandler(tool.Name))
	}

	// Add special prompt management tools
	compositeServer.addPromptManagementTools(mcpServer)

	// Register prompts with the MCP server
	loadedPrompts := promptManager.GetPrompts()
	fmt.Fprintf(os.Stderr, "Registering %d prompts with MCP server\n", len(loadedPrompts))
	for _, prompt := range loadedPrompts {
		fmt.Fprintf(os.Stderr, "Registering prompt: %s - %s\n", prompt.Name, prompt.Description)
		mcpServer.AddPrompt(prompt, compositeServer.createPromptHandler(prompt.Name))
	}

	// Dynamically discover and register resources using the resource manager
	compositeServer.resourceManager.DiscoverAndRegisterResources(mcpServer)

	return compositeServer
}

// Start starts both stdio and HTTP servers
func (s *MCPServer) Start(addr string) error {
	// Start MCP stdio server in a goroutine
	go func() {
		fmt.Fprintf(os.Stderr, "Starting MCP stdio server...\n")
		if err := server.ServeStdio(s.mcpServer); err != nil {
			fmt.Fprintf(os.Stderr, "MCP stdio server error: %v\n", err)
		}
	}()

	// Start HTTP server using the library's StreamableHTTP server
	fmt.Fprintf(os.Stderr, "Starting StreamableHTTP server on %s\n", addr)
	httpServer := server.NewStreamableHTTPServer(s.mcpServer,
		server.WithEndpointPath("/mcp"),
	)
	return httpServer.Start(addr)
}

// StartWithMode starts the server in the specified mode
func (s *MCPServer) StartWithMode(mode string, addr string) error {
	switch mode {
	case "stdio":
		fmt.Fprintf(os.Stderr, "Starting MCP stdio server only...\n")
		return server.ServeStdio(s.mcpServer)
	case "http":
		fmt.Fprintf(os.Stderr, "Starting StreamableHTTP server only on %s\n", addr)
		httpServer := server.NewStreamableHTTPServer(s.mcpServer,
			server.WithEndpointPath("/mcp"),
		)
		return httpServer.Start(addr)
	case "both":
		return s.Start(addr)
	default:
		return fmt.Errorf("invalid mode: %s. Valid modes are: stdio, http, both", mode)
	}
}

// GetConfig returns the server's configuration
func (s *MCPServer) GetConfig() *config.Config {
	return s.config
}

// GetTools returns the server's tools
func (s *MCPServer) GetTools() []tools.Tool {
	return s.tools
}

// ResolveRequiredParameters wraps the package-level function with config access
func (s *MCPServer) ResolveRequiredParameters(requiredParams []string, providedParams map[string]interface{}, pathPattern string) map[string]interface{} {
	return ResolveRequiredParameters(s.config, requiredParams, providedParams, pathPattern)
}

// ExecuteAPICall wraps the package-level function with config and spec access
func (s *MCPServer) ExecuteAPICall(method, path string, parameters map[string]interface{}, requestBody interface{}) (map[string]interface{}, error) {
	return ExecuteAPICall(s.config, s.spec, method, path, parameters, requestBody)
}

// GetPrompts returns all loaded prompts
func (s *MCPServer) GetPrompts() []mcp.Prompt {
	if s.promptManager == nil {
		return []mcp.Prompt{}
	}
	return s.promptManager.GetPrompts()
}

// GetPrompt returns a specific prompt by name
func (s *MCPServer) GetPrompt(name string) (*mcp.Prompt, bool) {
	if s.promptManager == nil {
		return nil, false
	}
	return s.promptManager.GetPrompt(name)
}

// GetPromptContent returns the content of a specific prompt
func (s *MCPServer) GetPromptContent(name string) (string, error) {
	if s.promptManager == nil {
		return "", fmt.Errorf("prompt manager not initialized")
	}
	return s.promptManager.GetPromptContent(name)
}

// GetPromptContentWithSubstitution returns the content of a specific prompt with variable substitution
func (s *MCPServer) GetPromptContentWithSubstitution(name string) (string, error) {
	if s.promptManager == nil {
		return "", fmt.Errorf("prompt manager not initialized")
	}
	return s.promptManager.GetPromptContentWithSubstitution(name)
}

// ReloadPrompts reloads all prompts from the configured folder
func (s *MCPServer) ReloadPrompts() error {
	if s.promptManager == nil {
		return fmt.Errorf("prompt manager not initialized")
	}
	return s.promptManager.ReloadPrompts()
}

// convertToMCPTool converts our internal Tool to an MCP Tool
func convertToMCPTool(tool tools.Tool) mcp.Tool {
	// Create input schema from tool parameters
	var inputSchema mcp.ToolInputSchema
	if tool.Parameters != nil {
		if schemaType, ok := tool.Parameters["type"].(string); ok {
			inputSchema.Type = schemaType
		} else {
			inputSchema.Type = "object"
		}
		if properties, ok := tool.Parameters["properties"].(map[string]interface{}); ok {
			inputSchema.Properties = properties
		} else {
			inputSchema.Properties = map[string]any{}
		}
		if required, ok := tool.Parameters["required"].([]string); ok {
			inputSchema.Required = required
		}
	} else {
		inputSchema = mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
		}
	}

	return mcp.Tool{
		Name:        tool.Name,
		Description: tool.Description,
		InputSchema: inputSchema,
	}
}

// createToolHandler creates a tool handler function for the MCP server
func (s *MCPServer) createToolHandler(toolName string) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: "Error: Invalid arguments format",
					},
				},
			}, nil
		}

		invokeReq := InvokeRequest{
			Tool:      toolName,
			Arguments: args,
		}
		resp := s.InvokeTool(invokeReq)

		if resp.Error != "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: "Error: " + resp.Error,
					},
				},
			}, nil
		}

		// If this was a successful create operation, register the new resource
		if toolName == tools.ActionCreate {
			s.resourceManager.HandleResourceCreation(s.mcpServer, args, resp.Result)
		}

		// If this was a successful delete operation, unregister the resource
		if toolName == tools.ActionDelete {
			s.resourceManager.HandleResourceDeletion(args)
		}

		resultJSON, err := json.Marshal(resp.Result)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: "Failed to format result",
					},
				},
			}, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: string(resultJSON),
				},
			},
		}, nil
	}
}

// createPromptHandler creates a prompt handler function for the MCP server
func (s *MCPServer) createPromptHandler(promptName string) func(context.Context, mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	return func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		content, err := s.GetPromptContentWithSubstitution(promptName)
		if err != nil {
			return nil, fmt.Errorf("failed to get prompt content: %w", err)
		}

		return &mcp.GetPromptResult{
			Description: fmt.Sprintf("Prompt: %s", promptName),
			Messages: []mcp.PromptMessage{
				{
					Role: "user",
					Content: mcp.TextContent{
						Type: "text",
						Text: content,
					},
				},
			},
		}, nil
	}
}

// addPromptManagementTools adds special tools for managing prompts
func (s *MCPServer) addPromptManagementTools(mcpServer *server.MCPServer) {
	// Tool to list available prompts
	listPromptsSchema := mcp.ToolInputSchema{
		Type:       "object",
		Properties: map[string]any{},
		Required:   []string{},
	}

	listPromptsTool := mcp.Tool{
		Name:        "prompts",
		Description: "List all available prompts with their descriptions",
		InputSchema: listPromptsSchema,
	}

	mcpServer.AddTool(listPromptsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		prompts := s.GetPrompts()

		var promptList []string
		promptList = append(promptList, fmt.Sprintf("Found %d available prompts:\n", len(prompts)))

		for _, prompt := range prompts {
			promptList = append(promptList, fmt.Sprintf("• **%s**: %s", prompt.Name, prompt.Description))
		}

		result := strings.Join(promptList, "\n")

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: result,
				},
			},
		}, nil
	})

	// Tool to get specific prompt content
	getPromptSchema := mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]any{
			"name": map[string]any{
				"type":        "string",
				"description": "The name of the prompt to retrieve",
			},
		},
		Required: []string{"name"},
	}

	getPromptTool := mcp.Tool{
		Name:        "get_prompt",
		Description: "Get the content of a specific prompt by name",
		InputSchema: getPromptSchema,
	}

	mcpServer.AddTool(getPromptTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: "Error: Invalid arguments format",
					},
				},
			}, nil
		}

		promptName, ok := args["name"].(string)
		if !ok {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: "Error: 'name' parameter is required and must be a string",
					},
				},
			}, nil
		}

		prompt, exists := s.GetPrompt(promptName)
		if !exists {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: fmt.Sprintf("Error: Prompt '%s' not found", promptName),
					},
				},
			}, nil
		}

		content, err := s.GetPromptContent(promptName)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: fmt.Sprintf("Error getting prompt content: %v", err),
					},
				},
			}, nil
		}

		result := fmt.Sprintf("**Prompt: %s**\n\n**Description:** %s\n\n**Content:**\n%s",
			prompt.Name, prompt.Description, content)

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: result,
				},
			},
		}, nil
	})
}

// RegisterMetricsHandlers registers HTTP handlers for metrics
func (s *MCPServer) RegisterMetricsHandlers(mux *http.ServeMux) {
	if s.monitor == nil {
		// If no monitor is set, provide basic info
		mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"error":"monitoring not enabled"}`))
		})
		return
	}

	// Setup HTTP handler with the monitor
	httpHandler := monitoring.NewHTTPHandler(s.monitor)

	// Register endpoints
	mux.HandleFunc("/metrics", httpHandler.MetricsHandler)               // JSON format
	mux.HandleFunc("/metrics/prometheus", httpHandler.PrometheusHandler) // Prometheus format
	mux.HandleFunc("/health", httpHandler.HealthHandler)
	mux.HandleFunc("/gc", httpHandler.GCHandler)
}

// SetMonitor sets the resource monitor for the server
func (s *MCPServer) SetMonitor(monitor *monitoring.Monitor) {
	s.monitor = monitor
}
