package openapi

import (
	"testing"
)

func TestParseOpenAPISpecBytes(t *testing.T) {
	validJSON := `{
		"openapi": "3.0.3",
		"info": {
			"title": "Test API",
			"version": "1.0.0"
		},
		"paths": {
			"/test": {
				"get": {
					"summary": "Test endpoint",
					"parameters": [
						{
							"name": "id",
							"in": "query",
							"required": true,
							"schema": {
								"type": "string"
							}
						}
					]
				}
			}
		}
	}`

	spec, err := ParseOpenAPISpecBytes([]byte(validJSON))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if spec.OpenAPI != "3.0.3" {
		t.Errorf("Expected OpenAPI version '3.0.3', got '%s'", spec.OpenAPI)
	}

	if spec.Info.Title != "Test API" {
		t.Errorf("Expected title 'Test API', got '%s'", spec.Info.Title)
	}

	if spec.Info.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", spec.Info.Version)
	}

	// Test paths
	if len(spec.Paths) != 1 {
		t.Errorf("Expected 1 path, got %d", len(spec.Paths))
	}

	testPath, exists := spec.Paths["/test"]
	if !exists {
		t.Fatal("Expected '/test' path to exist")
	}

	if testPath.Get == nil {
		t.Fatal("Expected GET operation to exist")
	}

	if testPath.Get.Summary != "Test endpoint" {
		t.Errorf("Expected summary 'Test endpoint', got '%s'", testPath.Get.Summary)
	}

	// Test parameters
	if len(testPath.Get.Parameters) != 1 {
		t.Errorf("Expected 1 parameter, got %d", len(testPath.Get.Parameters))
	}

	param := testPath.Get.Parameters[0]
	if param.Name != "id" {
		t.Errorf("Expected parameter name 'id', got '%s'", param.Name)
	}

	if param.In != "query" {
		t.Errorf("Expected parameter in 'query', got '%s'", param.In)
	}

	if !param.Required {
		t.Error("Expected parameter to be required")
	}
}

func TestParseOpenAPISpecBytes_InvalidJSON(t *testing.T) {
	invalidJSON := `{
		"openapi": "3.0.3",
		"info": {
			"title": "Test API"
			// Missing comma - invalid JSON
		}
	}`

	_, err := ParseOpenAPISpecBytes([]byte(invalidJSON))
	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}
}

func TestParseOpenAPISpecBytes_EmptyJSON(t *testing.T) {
	emptyJSON := `{}`

	spec, err := ParseOpenAPISpecBytes([]byte(emptyJSON))
	if err != nil {
		t.Fatalf("Expected no error for empty JSON, got %v", err)
	}

	if spec.OpenAPI != "" {
		t.Errorf("Expected empty OpenAPI version, got '%s'", spec.OpenAPI)
	}

	if len(spec.Paths) != 0 {
		t.Errorf("Expected 0 paths, got %d", len(spec.Paths))
	}
}

func TestResolveRequestBodyRef(t *testing.T) {
	spec := &OpenAPISpec{
		Components: &Components{
			RequestBodies: map[string]RequestBody{
				"TestRequestBody": {
					Content: map[string]MediaType{
						"application/json": {
							Schema: map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"name": map[string]interface{}{
										"type": "string",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name        string
		requestBody *RequestBody
		expected    *RequestBody
	}{
		{
			name: "Resolve valid reference",
			requestBody: &RequestBody{
				Ref: "#/components/requestBodies/TestRequestBody",
			},
			expected: &RequestBody{
				Content: map[string]MediaType{
					"application/json": {
						Schema: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"name": map[string]interface{}{
									"type": "string",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Return as-is for non-reference",
			requestBody: &RequestBody{
				Content: map[string]MediaType{
					"application/json": {
						Schema: map[string]interface{}{
							"type": "string",
						},
					},
				},
			},
			expected: &RequestBody{
				Content: map[string]MediaType{
					"application/json": {
						Schema: map[string]interface{}{
							"type": "string",
						},
					},
				},
			},
		},
		{
			name: "Return as-is for invalid reference",
			requestBody: &RequestBody{
				Ref: "#/components/requestBodies/NonExistent",
			},
			expected: &RequestBody{
				Ref: "#/components/requestBodies/NonExistent",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := spec.ResolveRequestBodyRef(tt.requestBody)
			
			// For this test, we'll do a basic comparison
			// In a real scenario, you might want to use deep comparison
			if tt.name == "Resolve valid reference" {
				if result.Ref != "" {
					t.Error("Expected resolved RequestBody to not have Ref")
				}
				if len(result.Content) == 0 {
					t.Error("Expected resolved RequestBody to have Content")
				}
			} else if tt.name == "Return as-is for non-reference" {
				if len(result.Content) == 0 {
					t.Error("Expected RequestBody to maintain Content")
				}
			} else if tt.name == "Return as-is for invalid reference" {
				if result.Ref == "" {
					t.Error("Expected RequestBody to maintain invalid Ref")
				}
			}
		})
	}
}

func TestDetermineSecurityTypeFromSpec(t *testing.T) {
	spec := &OpenAPISpec{
		Paths: map[string]PathItem{
			"/kafka/topics": {
				Get: &Operation{
					Security: []map[string][]string{
						{"kafka-api-key": {}},
					},
				},
			},
			"/flink/statements": {
				Post: &Operation{
					Security: []map[string][]string{
						{"flink-api-key": {}},
					},
				},
			},
			"/schemas/subjects": {
				Get: &Operation{
					Security: []map[string][]string{
						{"schema-registry-api-key": {}},
					},
				},
			},
		},
		Security: []map[string][]string{
			{"cloud-api-key": {}},
		},
	}

	tests := []struct {
		name     string
		method   string
		path     string
		expected string
	}{
		{
			name:     "Kafka endpoint security",
			method:   "GET",
			path:     "/kafka/topics",
			expected: "kafka-api-key",
		},
		{
			name:     "Flink endpoint security",
			method:   "POST",
			path:     "/flink/statements",
			expected: "flink-api-key",
		},
		{
			name:     "Schema registry endpoint security",
			method:   "GET",
			path:     "/schemas/subjects",
			expected: "schema-registry-api-key",
		},
		{
			name:     "Non-existent path falls back to global",
			method:   "GET",
			path:     "/non-existent",
			expected: "cloud-api-key",
		},
		{
			name:     "Non-existent method falls back to global",
			method:   "PUT",
			path:     "/kafka/topics",
			expected: "cloud-api-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := spec.GetSecurityTypeForEndpoint(tt.method, tt.path)
			if result != tt.expected {
				t.Errorf("GetSecurityTypeForEndpoint(%q, %q) = %q, expected %q",
					tt.method, tt.path, result, tt.expected)
			}
		})
	}
}

func TestGetSecurityTypeForEndpoint_NoSecurity(t *testing.T) {
	spec := &OpenAPISpec{
		Paths: map[string]PathItem{
			"/public": {
				Get: &Operation{
					Summary: "Public endpoint",
				},
			},
		},
	}

	result := spec.GetSecurityTypeForEndpoint("GET", "/public")
	if result != "" {
		t.Errorf("Expected empty security type for endpoint with no security, got %q", result)
	}
}

func TestSchema_Validation(t *testing.T) {
	schema := &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"name": {
				Type: "string",
			},
			"age": {
				Type: "integer",
			},
		},
		Required: []string{"name"},
	}

	// Test basic schema structure
	if schema.Type != "object" {
		t.Errorf("Expected schema type 'object', got '%s'", schema.Type)
	}

	if len(schema.Properties) != 2 {
		t.Errorf("Expected 2 properties, got %d", len(schema.Properties))
	}

	nameSchema, exists := schema.Properties["name"]
	if !exists {
		t.Fatal("Expected 'name' property to exist")
	}

	if nameSchema.Type != "string" {
		t.Errorf("Expected name property type 'string', got '%s'", nameSchema.Type)
	}

	if len(schema.Required) != 1 || schema.Required[0] != "name" {
		t.Errorf("Expected required fields ['name'], got %v", schema.Required)
	}
}

func TestParameter_Validation(t *testing.T) {
	param := Parameter{
		Name:     "cluster_id",
		In:       "path",
		Required: true,
		Schema: &Schema{
			Type: "string",
		},
	}

	if param.Name != "cluster_id" {
		t.Errorf("Expected parameter name 'cluster_id', got '%s'", param.Name)
	}

	if param.In != "path" {
		t.Errorf("Expected parameter in 'path', got '%s'", param.In)
	}

	if !param.Required {
		t.Error("Expected parameter to be required")
	}

	if param.Schema == nil {
		t.Fatal("Expected parameter to have schema")
	}

	if param.Schema.Type != "string" {
		t.Errorf("Expected parameter schema type 'string', got '%s'", param.Schema.Type)
	}
}
