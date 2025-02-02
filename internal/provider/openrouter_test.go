package provider

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"golang.org/x/time/rate"

	"github.com/pimentel/peppergo/pkg/types"
)

// testConfig holds test configuration
type testConfig struct {
	apiKey         string
	model          string
	fallbackAPIKey string
	fallbackModel  string
}

// getTestConfig gets configuration from environment variables
func getTestConfig(t *testing.T) testConfig {
	t.Helper()
	return testConfig{
		apiKey:         os.Getenv("PEPPERPY_API_KEY"),
		model:          os.Getenv("PEPPERPY_MODEL"),
		fallbackAPIKey: os.Getenv("PEPPERPY_FALLBACK_API_KEY"),
		fallbackModel:  os.Getenv("PEPPERPY_FALLBACK_MODEL"),
	}
}

func TestOpenRouterProviderIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := getTestConfig(t)
	if cfg.apiKey == "" {
		t.Skip("PEPPERPY_API_KEY not set")
	}

	if cfg.model == "" {
		cfg.model = "google/gemini-2.0-flash-exp:free" // Default model
	}

	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	// Create rate limiter - 10 requests per minute
	limiter := rate.NewLimiter(rate.Every(6*time.Second), 1)

	config := &OpenRouterConfig{
		APIKey:      cfg.apiKey,
		Model:       cfg.model,
		MaxTokens:   2000,
		Temperature: 0.7,
		RateLimiter: limiter,
	}

	t.Run("full provider lifecycle", func(t *testing.T) {
		provider := NewOpenRouterProvider(logger, config)
		err := provider.Initialize(ctx)
		require.NoError(t, err)

		// Test basic generation
		response, err := provider.Generate(ctx, "Say 'Hello, World!'")
		require.NoError(t, err)
		assert.NotEmpty(t, response.Content)
		assert.Greater(t, response.Usage.TotalTokens, 0)

		// Wait for rate limit
		time.Sleep(6 * time.Second)

		// Test generation with options
		response, err = provider.Generate(ctx,
			"Count to 3",
			types.WithTemperature(0.1),
			types.WithMaxTokens(20),
		)
		require.NoError(t, err)
		assert.NotEmpty(t, response.Content)
		assert.LessOrEqual(t, response.Usage.TotalTokens, 20)
	})

	t.Run("error handling", func(t *testing.T) {
		provider := NewOpenRouterProvider(logger, config)
		err := provider.Initialize(ctx)
		require.NoError(t, err)

		// Test empty prompt
		_, err = provider.Generate(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty prompt")

		// Test invalid temperature
		_, err = provider.Generate(ctx, "test", types.WithTemperature(2.0))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid temperature")

		// Test context cancellation
		ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Millisecond)
		defer cancel()
		_, err = provider.Generate(ctxWithTimeout, "This should timeout")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context")
	})

	t.Run("retry mechanism", func(t *testing.T) {
		provider := NewOpenRouterProvider(logger, config)
		err := provider.Initialize(ctx)
		require.NoError(t, err)

		// Test with retries
		response, err := provider.Generate(ctx, "Test with retries", types.WithRetries(3))
		require.NoError(t, err)
		assert.NotEmpty(t, response.Content)
	})

	t.Run("fallback provider", func(t *testing.T) {
		if cfg.fallbackAPIKey == "" || cfg.fallbackModel == "" {
			t.Skip("Fallback configuration not set")
		}

		fallbackConfig := &OpenRouterConfig{
			APIKey:      cfg.fallbackAPIKey,
			Model:       cfg.fallbackModel,
			MaxTokens:   2000,
			Temperature: 0.7,
			RateLimiter: rate.NewLimiter(rate.Every(6*time.Second), 1),
		}

		provider := NewOpenRouterProvider(logger, fallbackConfig)
		err := provider.Initialize(ctx)
		require.NoError(t, err)

		response, err := provider.Generate(ctx, "Test fallback provider")
		require.NoError(t, err)
		assert.NotEmpty(t, response.Content)
	})
}

func TestOpenRouterProviderConfig(t *testing.T) {
	logger := zaptest.NewLogger(t)

	t.Run("validate config", func(t *testing.T) {
		testCases := []struct {
			name        string
			config      *OpenRouterConfig
			shouldError bool
		}{
			{
				name: "valid config",
				config: &OpenRouterConfig{
					APIKey:      "test-key",
					Model:       "test-model",
					MaxTokens:   2000,
					Temperature: 0.7,
					RateLimiter: rate.NewLimiter(rate.Every(time.Second), 1),
				},
				shouldError: false,
			},
			{
				name: "missing API key",
				config: &OpenRouterConfig{
					Model:       "test-model",
					MaxTokens:   2000,
					Temperature: 0.7,
				},
				shouldError: true,
			},
			{
				name: "missing model",
				config: &OpenRouterConfig{
					APIKey:      "test-key",
					MaxTokens:   2000,
					Temperature: 0.7,
				},
				shouldError: true,
			},
			{
				name: "invalid temperature",
				config: &OpenRouterConfig{
					APIKey:      "test-key",
					Model:       "test-model",
					MaxTokens:   2000,
					Temperature: 1.5,
				},
				shouldError: true,
			},
			{
				name: "invalid max tokens",
				config: &OpenRouterConfig{
					APIKey:      "test-key",
					Model:       "test-model",
					MaxTokens:   -1,
					Temperature: 0.7,
				},
				shouldError: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				provider := NewOpenRouterProvider(logger, tc.config)
				err := provider.Initialize(context.Background())
				if tc.shouldError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
} 