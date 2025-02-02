package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"golang.org/x/time/rate"

	"github.com/pimentel/peppergo/internal/provider"
	"github.com/pimentel/peppergo/pkg/types"
)

func TestAnthropicProviderIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	config := &provider.Config{
		APIKey:      apiKey,
		Model:       "claude-2",
		MaxTokens:   2000,
		Temperature: 0.7,
	}

	t.Run("full provider lifecycle", func(t *testing.T) {
		// Initialize provider
		provider := provider.NewAnthropicProvider(logger, config)
		err := provider.Initialize(ctx)
		require.NoError(t, err)

		// Test basic generation
		response, err := provider.Generate(ctx, "Say 'Hello, World!'")
		require.NoError(t, err)
		assert.Contains(t, response.Content, "Hello")
		assert.Greater(t, response.Usage.TotalTokens, 0)

		// Test generation with options
		response, err = provider.Generate(ctx,
			"Count to 3",
			types.WithTemperature(0.1),
			types.WithMaxTokens(20),
		)
		require.NoError(t, err)
		assert.Contains(t, response.Content, "1")
		assert.LessOrEqual(t, response.Usage.TotalTokens, 20)

		// Test streaming
		stream, err := provider.Stream(ctx, "Say 'Hi' slowly")
		require.NoError(t, err)

		var streamContent string
		for response := range stream {
			streamContent += response.Content
		}
		assert.Contains(t, streamContent, "Hi")
	})

	t.Run("error handling", func(t *testing.T) {
		provider := provider.NewAnthropicProvider(logger, config)
		err := provider.Initialize(ctx)
		require.NoError(t, err)

		// Test empty prompt
		_, err = provider.Generate(ctx, "")
		assert.Error(t, err)

		// Test context cancellation
		ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Millisecond)
		defer cancel()
		_, err = provider.Generate(ctxWithTimeout, "This should timeout")
		assert.Error(t, err)
	})

	t.Run("concurrent requests", func(t *testing.T) {
		provider := provider.NewAnthropicProvider(logger, config)
		err := provider.Initialize(ctx)
		require.NoError(t, err)

		// Run multiple requests concurrently
		const numRequests = 3
		results := make(chan error, numRequests)

		for i := 0; i < numRequests; i++ {
			go func() {
				_, err := provider.Generate(ctx, "Quick test")
				results <- err
			}()
		}

		// Check all requests completed successfully
		for i := 0; i < numRequests; i++ {
			err := <-results
			assert.NoError(t, err)
		}
	})

	t.Run("agent integration", func(t *testing.T) {
		provider := provider.NewAnthropicProvider(logger, config)
		err := provider.Initialize(ctx)
		require.NoError(t, err)

		agent := NewSimpleAgent(provider, logger)
		err = agent.ProcessTask(ctx, "What is 2+2?")
		assert.NoError(t, err)
	})
}

func TestSimpleAgentIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	config := &provider.Config{
		APIKey:      apiKey,
		Model:       "claude-2",
		MaxTokens:   2000,
		Temperature: 0.7,
	}

	provider := provider.NewAnthropicProvider(logger, config)
	err := provider.Initialize(ctx)
	require.NoError(t, err)

	t.Run("process multiple tasks", func(t *testing.T) {
		agent := NewSimpleAgent(provider, logger)

		tasks := []string{
			"What is Go?",
			"Explain error handling",
			"Define concurrency",
		}

		for _, task := range tasks {
			err := agent.ProcessTask(ctx, task)
			assert.NoError(t, err)
		}
	})

	t.Run("handle short responses", func(t *testing.T) {
		agent := NewSimpleAgent(provider, logger)

		// This should trigger a follow-up question due to short response
		err := agent.ProcessTask(ctx, "Say hi")
		assert.NoError(t, err)
	})
} 