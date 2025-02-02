// Package types provides common types and interfaces for the PepperGo system
package types

import "context"

// Provider represents an AI provider interface
type Provider interface {
	// Initialize sets up the provider
	Initialize(ctx context.Context) error

	// Generate generates a response for the given prompt
	Generate(ctx context.Context, prompt string, opts ...ExecuteOption) (*Response, error)
}

// GenerateOption is an alias for ExecuteOption for backward compatibility
type GenerateOption = ExecuteOption

// GenerateOptions is an alias for ExecuteOptions for backward compatibility
type GenerateOptions = ExecuteOptions 