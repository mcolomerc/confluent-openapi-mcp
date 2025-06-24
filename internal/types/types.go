package types

// InvokeRequest represents a tool invocation request
type InvokeRequest struct {
	Tool      string                 `json:"tool"`
	Arguments map[string]interface{} `json:"arguments"`
}

// InvokeResponse represents a tool invocation response
type InvokeResponse struct {
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"error,omitempty"`
}
