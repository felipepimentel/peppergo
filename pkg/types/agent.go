package types

import (
	"context"
)

// Agent represents the core interface for all AI agents in the system.
type Agent interface {
	// Initialize sets up the agent with its configuration and capabilities
	Initialize(ctx context.Context) error

	// Execute runs a task with the given input and returns a response
	Execute(ctx context.Context, task string, opts ...ExecuteOption) (*Response, error)

	// Cleanup performs any necessary cleanup when the agent is done
	Cleanup(ctx context.Context) error

	// AddCapability adds a new capability to the agent
	AddCapability(capability Capability) error

	// AddTool adds a new tool to the agent
	AddTool(tool Tool) error

	// UseProvider sets the AI provider for this agent
	UseProvider(provider Provider) error

	// ID returns the unique identifier for this agent
	ID() string

	// Name returns the human-readable name of this agent
	Name() string

	// Version returns the version of this agent
	Version() string
}

// ExecuteOption represents an option that can be passed to Agent.Execute
type ExecuteOption func(*ExecuteOptions)

// ExecuteOptions contains all possible options for Agent.Execute
type ExecuteOptions struct {
	Temperature float64
	MaxTokens   int
	Stream      bool
}

// Response represents a response from an agent
type Response struct {
	Content    string
	Metadata   map[string]interface{}
	Usage      *Usage
	Timestamp  int64
	FinishReason string
}

// Usage contains token usage information
type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// WithTemperature sets the temperature for generation
func WithTemperature(temp float64) ExecuteOption {
	return func(o *ExecuteOptions) {
		o.Temperature = temp
	}
}

// WithMaxTokens sets the maximum number of tokens to generate
func WithMaxTokens(tokens int) ExecuteOption {
	return func(o *ExecuteOptions) {
		o.MaxTokens = tokens
	}
}

// WithStream enables streaming responses
func WithStream(stream bool) ExecuteOption {
	return func(o *ExecuteOptions) {
		o.Stream = stream
	}
} 