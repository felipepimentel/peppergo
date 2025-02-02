package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"github.com/pimentel/peppergo/internal/provider"
	"github.com/pimentel/peppergo/pkg/types"
)

func main() {
	// Initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Printf("Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Create rate limiter - 10 requests per minute
	limiter := rate.NewLimiter(rate.Every(6*time.Second), 1)

	// Create primary provider configuration
	primaryConfig := &provider.OpenRouterConfig{
		APIKey:      os.Getenv("PEPPERPY_API_KEY"),
		Model:       os.Getenv("PEPPERPY_MODEL"),
		MaxTokens:   2000,
		Temperature: 0.7,
		RateLimiter: limiter,
	}

	// Create fallback provider configuration
	fallbackConfig := &provider.OpenRouterConfig{
		APIKey:      os.Getenv("PEPPERPY_FALLBACK_API_KEY"),
		Model:       os.Getenv("PEPPERPY_FALLBACK_MODEL"),
		MaxTokens:   2000,
		Temperature: 0.7,
	}

	// Create and initialize primary provider
	primaryProvider := provider.NewOpenRouterProvider(logger, primaryConfig)
	ctx := context.Background()

	if err := primaryProvider.Initialize(ctx); err != nil {
		logger.Fatal("Failed to initialize primary provider", zap.Error(err))
	}

	// Create and initialize fallback provider
	fallbackProvider := provider.NewOpenRouterProvider(logger, fallbackConfig)
	if err := fallbackProvider.Initialize(ctx); err != nil {
		logger.Fatal("Failed to initialize fallback provider", zap.Error(err))
	}

	// Example 1: Basic Generation with Primary Provider
	fmt.Println("\n=== Example 1: Basic Generation (Primary Provider) ===")
	response, err := primaryProvider.Generate(ctx, "Explain what is Go programming language in one sentence.")
	if err != nil {
		logger.Error("Failed to generate response from primary provider", zap.Error(err))
		// Try fallback provider
		response, err = fallbackProvider.Generate(ctx, "Explain what is Go programming language in one sentence.")
		if err != nil {
			logger.Fatal("Both providers failed", zap.Error(err))
		}
	}
	fmt.Printf("Response: %s\n", response.Content)
	fmt.Printf("Model Used: %s\n", primaryConfig.Model)
	fmt.Printf("Tokens Used: %d\n", response.Usage.TotalTokens)

	// Example 2: Generation with Options
	fmt.Println("\n=== Example 2: Generation with Options ===")
	response, err = primaryProvider.Generate(ctx,
		"Write a haiku about coding.",
		types.WithTemperature(0.9),
		types.WithMaxTokens(50),
	)
	if err != nil {
		logger.Error("Failed to generate response from primary provider", zap.Error(err))
		// Try fallback provider with same options
		response, err = fallbackProvider.Generate(ctx,
			"Write a haiku about coding.",
			types.WithTemperature(0.9),
			types.WithMaxTokens(50),
		)
		if err != nil {
			logger.Fatal("Both providers failed", zap.Error(err))
		}
	}
	fmt.Printf("Response: %s\n", response.Content)
	fmt.Printf("Model Used: %s\n", primaryConfig.Model)
	fmt.Printf("Tokens Used: %d\n", response.Usage.TotalTokens)

	// Example 3: Generation with Retries
	fmt.Println("\n=== Example 3: Generation with Retries ===")
	response, err = primaryProvider.Generate(ctx,
		"Tell me a short joke.",
		provider.WithRetries(3),
	)
	if err != nil {
		logger.Error("Failed to generate response from primary provider", zap.Error(err))
		// Try fallback provider with retries
		response, err = fallbackProvider.Generate(ctx,
			"Tell me a short joke.",
			provider.WithRetries(3),
		)
		if err != nil {
			logger.Fatal("Both providers failed", zap.Error(err))
		}
	}
	fmt.Printf("Response: %s\n", response.Content)
	fmt.Printf("Model Used: %s\n", primaryConfig.Model)
	fmt.Printf("Tokens Used: %d\n", response.Usage.TotalTokens)

	// Example 4: Error Handling and Fallback
	fmt.Println("\n=== Example 4: Error Handling and Fallback ===")
	_, err = primaryProvider.Generate(ctx, "", types.WithMaxTokens(-1))
	if err != nil {
		fmt.Printf("Primary provider failed as expected: %v\n", err)
		fmt.Println("Trying fallback provider...")
		response, err = fallbackProvider.Generate(ctx, "What is your name?")
		if err != nil {
			logger.Fatal("Fallback provider also failed", zap.Error(err))
		}
		fmt.Printf("Fallback Response: %s\n", response.Content)
		fmt.Printf("Fallback Model Used: %s\n", fallbackConfig.Model)
	}

	// Example 5: Context Cancellation
	fmt.Println("\n=== Example 5: Context Cancellation ===")
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Millisecond)
	defer cancel()

	_, err = primaryProvider.Generate(ctxWithTimeout, "This request will be cancelled.")
	if err != nil {
		fmt.Printf("Expected timeout error: %v\n", err)
	}
} 