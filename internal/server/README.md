# Server Package Structure

This document describes the refactored structure of the `internal/server` package for better organization and separation of concerns.

## File Organization

### Core Files

- **`server.go`** - Core server logic and composition
  - `MCPServer` struct definition
  - Server lifecycle management (Start, StartWithMode)
  - Configuration and tools access methods
  - Stdio request handling
  - Server startup and shutdown logic

- **`handlers.go`** - HTTP request handlers
  - HTTP endpoint handlers (`/tools`, `/invoke`, `/mcp`)
  - JSON-RPC message handling for MCP protocol
  - HTTP response formatting
  - MCP-specific types and error handling

- **`tool_invocation.go`** - Tool invocation business logic
  - `InvokeTool` method with parameter validation
  - Request body building from OpenAPI schemas
  - API call execution logic
  - Helper functions for schema processing

### Supporting Files

- **`httpmux.go`** - HTTP multiplexer wrapper
  - Wraps `http.ServeMux` for extensibility
  - Provides clean interface for HTTP routing

- **`mcpstdio.go`** - MCP stdio server wrapper
  - Wraps the mark3labs/mcp-go stdio server
  - Tool registration and management
  - Stdio protocol handling

- **`invoke.go`** - Low-level API invocation utilities
  - HTTP client configuration
  - API authentication handling
  - Parameter resolution from config
  - Raw API call execution

## Architecture Benefits

### Separation of Concerns
- **HTTP concerns** are isolated in `handlers.go`
- **Stdio concerns** remain in `mcpstdio.go` 
- **Core business logic** for tool invocation is in `tool_invocation.go`
- **Server orchestration** stays in `server.go`

### Maintainability
- Each file has a focused responsibility
- Easier to locate and modify specific functionality
- Reduced coupling between HTTP and stdio implementations

### Extensibility
- New HTTP endpoints can be added to `handlers.go`
- New tool invocation logic can be added to `tool_invocation.go`
- Server modes and startup logic can be modified in `server.go`

## Usage

The `MCPServer` struct uses composition to combine:
- HTTP functionality via `HTTPMuxWrapper`
- Stdio functionality via `MCPStdioServerWrapper`
- Shared tool invocation logic

The server can run in three modes:
- `stdio` - MCP stdio protocol only
- `http` - HTTP endpoints only  
- `both` - Both protocols simultaneously

This structure maintains backward compatibility while providing cleaner separation for future development.
