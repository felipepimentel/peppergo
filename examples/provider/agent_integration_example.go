package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/pimentel/peppergo/internal/agent"
	"github.com/pimentel/peppergo/internal/provider"
	"github.com/pimentel/peppergo/pkg/types"

	"golang.org/x/time/rate"
)

// SimpleAgent is a basic agent that uses the Anthropic provider
type SimpleAgent struct {
	provider types.Provider
	logger   *zap.Logger
}

// NewSimpleAgent creates a new SimpleAgent instance
func NewSimpleAgent(provider types.Provider, logger *zap.Logger) *SimpleAgent {
	return &SimpleAgent{
		provider: provider,
		logger:   logger,
	}
}

// ProcessTask processes a task using the provider
func (a *SimpleAgent) ProcessTask(ctx context.Context, task string) error {
	a.logger.Info("Processing task", zap.String("task", task))

	// Generate initial response
	response, err := a.provider.Generate(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to generate response: %w", err)
	}

	fmt.Printf("Initial Response: %s\n", response.Content)

	// If the response needs clarification, ask a follow-up question
	if len(response.Content) < 50 {
		followUp := fmt.Sprintf("Could you elaborate more on: %s", response.Content)
		response, err = a.provider.Generate(ctx, followUp)
		if err != nil {
			return fmt.Errorf("failed to generate follow-up response: %w", err)
		}
		fmt.Printf("Follow-up Response: %s\n", response.Content)
	}

	return nil
}

func main() {
	// Initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Printf("Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Create provider configuration
	config := &provider.Config{
		APIKey:      os.Getenv("ANTHROPIC_API_KEY"),
		Model:       "claude-2",
		MaxTokens:   2000,
		Temperature: 0.7,
	}

	// Create and initialize provider
	anthropicProvider := provider.NewAnthropicProvider(logger, config)
	ctx := context.Background()

	if err := anthropicProvider.Initialize(ctx); err != nil {
		logger.Fatal("Failed to initialize provider", zap.Error(err))
	}

	// Create agent
	agent := NewSimpleAgent(anthropicProvider, logger)

	// Example tasks
	tasks := []string{
		"What are the key principles of Go programming?",
		"Give me a short code example of error handling in Go.",
		"Explain concurrency in Go.",
	}

	// Process each task
	for i, task := range tasks {
		fmt.Printf("\n=== Task %d ===\n", i+1)
		if err := agent.ProcessTask(ctx, task); err != nil {
			logger.Error("Failed to process task",
				zap.String("task", task),
				zap.Error(err))
		}
	}
} 