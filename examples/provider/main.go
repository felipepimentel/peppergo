package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pimentel/peppergo/pkg/types"
)

// ExampleProvider implements the types.Provider interface
type ExampleProvider struct {
	name   string
	models []string
}

// NewExampleProvider creates a new example provider instance
func NewExampleProvider() *ExampleProvider {
	return &ExampleProvider{
		name: "example",
		models: []string{
			"example-model-1",
			"example-model-2",
		},
	}
}

// Name returns the provider's name
func (p *ExampleProvider) Name() string {
	return p.name
}

// AvailableModels returns the list of available models
func (p *ExampleProvider) AvailableModels() []string {
	return p.models
}

// Chat sends a chat completion request and returns a response
func (p *ExampleProvider) Chat(ctx context.Context, req *types.ChatRequest) (*types.ChatResponse, error) {
	// Simulate some processing time
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(100 * time.Millisecond):
	}

	// Create example response
	return &types.ChatResponse{
		ID:      "example-response-123",
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   req.Model,
		Choices: []types.Choice{
			{
				Index: 0,
				Message: types.Message{
					Role:    "assistant",
					Content: "This is an example response. Your message was: " + req.Messages[len(req.Messages)-1].Content,
				},
				FinishReason: "stop",
			},
		},
		Usage: types.Usage{
			PromptTokens:     100,
			CompletionTokens: 50,
			TotalTokens:      150,
		},
	}, nil
}

// StreamChat streams chat completion responses
func (p *ExampleProvider) StreamChat(ctx context.Context, req *types.ChatRequest) (<-chan *types.ChatResponse, error) {
	responses := make(chan *types.ChatResponse)

	go func() {
		defer close(responses)

		// Set streaming flag
		req.Stream = true

		// Get the response
		resp, err := p.Chat(ctx, req)
		if err != nil {
			log.Printf("Error in stream chat: %v", err)
			return
		}

		// Send the response
		select {
		case <-ctx.Done():
			return
		case responses <- resp:
		}
	}()

	return responses, nil
}

func main() {
	provider := NewExampleProvider()
	fmt.Printf("Provider %s initialized with models: %v\n", provider.Name(), provider.AvailableModels())

	// Example usage
	req := &types.ChatRequest{
		Model: provider.AvailableModels()[0],
		Messages: []types.Message{
			{
				Role:    "user",
				Content: "Hello!",
			},
		},
	}

	resp, err := provider.Chat(context.Background(), req)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Printf("Response: %s\n", resp.Choices[0].Message.Content)
} 