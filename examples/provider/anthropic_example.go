package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/yourusername/peppergo/internal/provider"
	"github.com/yourusername/peppergo/pkg/types"
)

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

	// Example 1: Basic Generation
	fmt.Println("\n=== Example 1: Basic Generation ===")
	response, err := anthropicProvider.Generate(ctx, "Explain what is Go programming language in one sentence.")
	if err != nil {
		logger.Error("Failed to generate response", zap.Error(err))
	} else {
		fmt.Printf("Response: %s\n", response.Content)
		fmt.Printf("Tokens Used: %d\n", response.Usage.TotalTokens)
	}

	// Example 2: Generation with Options
	fmt.Println("\n=== Example 2: Generation with Options ===")
	response, err = anthropicProvider.Generate(ctx,
		"Write a haiku about coding.",
		types.WithTemperature(0.9),
		types.WithMaxTokens(50),
	)
	if err != nil {
		logger.Error("Failed to generate response", zap.Error(err))
	} else {
		fmt.Printf("Response: %s\n", response.Content)
		fmt.Printf("Tokens Used: %d\n", response.Usage.TotalTokens)
	}

	// Example 3: Streaming Response
	fmt.Println("\n=== Example 3: Streaming Response ===")
	stream, err := anthropicProvider.Stream(ctx, "Count from 1 to 5 slowly.")
	if err != nil {
		logger.Error("Failed to create stream", zap.Error(err))
	} else {
		for response := range stream {
			fmt.Print(response.Content)
			time.Sleep(100 * time.Millisecond) // Simulate slow printing
		}
		fmt.Println()
	}

	// Example 4: Error Handling
	fmt.Println("\n=== Example 4: Error Handling ===")
	_, err = anthropicProvider.Generate(ctx, "", types.WithMaxTokens(-1))
	if err != nil {
		fmt.Printf("Expected error occurred: %v\n", err)
	}

	// Example 5: Context Cancellation
	fmt.Println("\n=== Example 5: Context Cancellation ===")
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	_, err = anthropicProvider.Generate(ctxWithTimeout, "This request will be cancelled due to timeout.")
	if err != nil {
		fmt.Printf("Expected timeout error: %v\n", err)
	}
} 