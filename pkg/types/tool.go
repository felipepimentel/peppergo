package types

import (
	"context"
)

// Tool represents a tool that can be used by an agent
type Tool interface {
	// Name returns the tool's unique identifier
	Name() string

	// Description returns a human-readable description of the tool
	Description() string

	// Initialize sets up the tool with its configuration
	Initialize(ctx context.Context) error

	// Execute runs the tool with the given arguments
	Execute(ctx context.Context, args map[string]interface{}) (interface{}, error)

	// Cleanup performs any necessary cleanup
	Cleanup(ctx context.Context) error

	// Schema returns the JSON schema for the tool's arguments
	Schema() *ToolSchema

	// Version returns the version of this tool
	Version() string
}

// ToolSchema represents the JSON schema for a tool's arguments
type ToolSchema struct {
	// Type specifies the type of the schema (usually "object")
	Type string `json:"type"`

	// Properties defines the properties of the schema
	Properties map[string]*PropertySchema `json:"properties"`

	// Required lists the required property names
	Required []string `json:"required,omitempty"`

	// Description provides a description of the schema
	Description string `json:"description,omitempty"`
}

// PropertySchema represents a property in a tool's schema
type PropertySchema struct {
	// Type specifies the type of the property
	Type string `json:"type"`

	// Description provides a description of the property
	Description string `json:"description,omitempty"`

	// Enum lists possible values for the property
	Enum []interface{} `json:"enum,omitempty"`

	// Default specifies the default value
	Default interface{} `json:"default,omitempty"`

	// Items defines the schema for array items
	Items *PropertySchema `json:"items,omitempty"`
}

// NewToolSchema creates a new ToolSchema instance
func NewToolSchema() *ToolSchema {
	return &ToolSchema{
		Type:       "object",
		Properties: make(map[string]*PropertySchema),
	}
}

// AddProperty adds a property to the schema
func (s *ToolSchema) AddProperty(name string, prop *PropertySchema) *ToolSchema {
	s.Properties[name] = prop
	return s
}

// AddRequired adds a required property name
func (s *ToolSchema) AddRequired(name string) *ToolSchema {
	s.Required = append(s.Required, name)
	return s
}

// NewPropertySchema creates a new PropertySchema instance
func NewPropertySchema(typ string) *PropertySchema {
	return &PropertySchema{
		Type: typ,
	}
} 