package types

import (
	"context"
)

// Capability represents a specific capability that can be added to an agent
type Capability interface {
	// Name returns the capability's unique identifier
	Name() string

	// Description returns a human-readable description of the capability
	Description() string

	// Initialize sets up the capability with its configuration
	Initialize(ctx context.Context) error

	// Execute runs the capability with the given input
	Execute(ctx context.Context, input interface{}) (interface{}, error)

	// Cleanup performs any necessary cleanup
	Cleanup(ctx context.Context) error

	// Requirements returns a list of required tools and capabilities
	Requirements() *Requirements

	// Version returns the version of this capability
	Version() string
}

// Requirements represents the requirements for a capability
type Requirements struct {
	// Tools lists the required tool names
	Tools []string

	// Capabilities lists the required capability names
	Capabilities []string

	// MinTokens specifies the minimum number of tokens needed
	MinTokens int

	// RequiresStreaming indicates if streaming support is required
	RequiresStreaming bool
}

// NewRequirements creates a new Requirements instance
func NewRequirements() *Requirements {
	return &Requirements{
		Tools:       make([]string, 0),
		Capabilities: make([]string, 0),
	}
}

// AddTool adds a required tool
func (r *Requirements) AddTool(name string) *Requirements {
	r.Tools = append(r.Tools, name)
	return r
}

// AddCapability adds a required capability
func (r *Requirements) AddCapability(name string) *Requirements {
	r.Capabilities = append(r.Capabilities, name)
	return r
}

// SetMinTokens sets the minimum required tokens
func (r *Requirements) SetMinTokens(tokens int) *Requirements {
	r.MinTokens = tokens
	return r
}

// SetRequiresStreaming sets whether streaming is required
func (r *Requirements) SetRequiresStreaming(required bool) *Requirements {
	r.RequiresStreaming = required
	return r
} 