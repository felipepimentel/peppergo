// Package types provides common types and interfaces for the PepperGo system
package types

import (
	"context"
)

// Message represents a standardized chat message format
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest represents a standardized request format for chat completions
type ChatRequest struct {
	Model       string     `json:"model"`
	Messages    []Message  `json:"messages"`
	MaxTokens   int        `json:"max_tokens,omitempty"`
	Temperature float32    `json:"temperature,omitempty"`
	Stream      bool       `json:"stream,omitempty"`
}

// ChatResponse represents a standardized response format for chat completions
type ChatResponse struct {
	ID      string    `json:"id"`
	Object  string    `json:"object"`
	Created int64     `json:"created"`
	Model   string    `json:"model"`
	Choices []Choice  `json:"choices"`
	Usage   Usage     `json:"usage"`
}

// Choice represents a completion choice in the response
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Provider defines the interface that all LLM providers must implement
type Provider interface {
	// Chat sends a chat completion request to the provider
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
	
	// StreamChat streams chat completion responses from the provider
	StreamChat(ctx context.Context, req *ChatRequest) (<-chan *ChatResponse, error)
	
	// Name returns the provider's name
	Name() string
	
	// AvailableModels returns the list of available models for this provider
	AvailableModels() []string
} 