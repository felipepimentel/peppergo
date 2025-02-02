package types

import (
	"context"
)

// Provider represents an AI provider interface
type Provider interface {
	// Initialize sets up the provider with its configuration
	Initialize(ctx context.Context) error

	// Generate generates a response for the given prompt
	Generate(ctx context.Context, prompt string, opts ...GenerateOption) (*Response, error)

	// Stream streams responses for the given prompt
	Stream(ctx context.Context, prompt string) (<-chan Response, error)

	// Name returns the provider's name
	Name() string

	// MaxTokens returns the maximum tokens supported by this provider
	MaxTokens() int

	// SupportsStreaming returns whether this provider supports streaming
	SupportsStreaming() bool
}

// GenerateOption represents an option that can be passed to Provider.Generate
type GenerateOption func(*GenerateOptions)

// GenerateOptions contains all possible options for Provider.Generate
type GenerateOptions struct {
	Temperature    float64
	MaxTokens     int
	TopP          float64
	FrequencyPenalty float64
	PresencePenalty  float64
	Stop           []string
	Model         string
}

// WithModel sets the model to use for generation
func WithModel(model string) GenerateOption {
	return func(o *GenerateOptions) {
		o.Model = model
	}
}

// WithTopP sets the top-p sampling parameter
func WithTopP(topP float64) GenerateOption {
	return func(o *GenerateOptions) {
		o.TopP = topP
	}
}

// WithFrequencyPenalty sets the frequency penalty
func WithFrequencyPenalty(penalty float64) GenerateOption {
	return func(o *GenerateOptions) {
		o.FrequencyPenalty = penalty
	}
}

// WithPresencePenalty sets the presence penalty
func WithPresencePenalty(penalty float64) GenerateOption {
	return func(o *GenerateOptions) {
		o.PresencePenalty = penalty
	}
}

// WithStop sets the stop sequences
func WithStop(stop []string) GenerateOption {
	return func(o *GenerateOptions) {
		o.Stop = stop
	}
} 