package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"github.com/pimentel/peppergo/internal/provider"
	"github.com/pimentel/peppergo/pkg/types"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v\n", err)
	}

	// Initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Printf("Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Get API key from environment
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		logger.Fatal("OPENROUTER_API_KEY environment variable is required")
	}

	// Get model from environment or use default
	model := os.Getenv("OPENROUTER_MODEL")
	if model == "" {
		model = "openai/gpt-3.5-turbo"
	}

	// Create rate limiter (3 requests per minute)
	limiter := rate.NewLimiter(rate.Every(20*time.Second), 1)

	// Configure provider
	config := &provider.OpenRouterConfig{
		APIKey:      apiKey,
		Model:       model,
		MaxTokens:   2000,
		Temperature: 0.7,
		RateLimiter: limiter,
	}

	// Create provider
	p := provider.NewOpenRouterProvider(logger, config)

	// Initialize provider
	if err := p.Initialize(context.Background()); err != nil {
		logger.Fatal("Failed to initialize provider", zap.Error(err))
	}

	// Simple prompt
	prompt := "Explain what is Go programming language in one sentence."

	// Print the prompt
	fmt.Printf("\nPrompt: %s\n", prompt)

	// Generate response
	resp, err := p.Generate(context.Background(), prompt,
		types.WithTemperature(0.7),
		types.WithMaxTokens(100),
		types.WithRetries(3),
	)
	if err != nil {
		logger.Fatal("Failed to generate response", zap.Error(err))
	}

	// Print response
	fmt.Printf("\nResponse: %s\n", resp.Content)
} 