package provider

import (
	"context"
	"testing"
	"time"

	"github.com/anthropic-ai/anthropic-sdk-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"

	"github.com/yourusername/peppergo/pkg/types"
)

// MockAnthropicClient is a mock implementation of the Anthropic client
type MockAnthropicClient struct {
	mock.Mock
}

func (m *MockAnthropicClient) Complete(ctx context.Context, req *anthropic.CompletionRequest) (*anthropic.CompletionResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*anthropic.CompletionResponse), args.Error(1)
}

func (m *MockAnthropicClient) CompleteStream(ctx context.Context, req *anthropic.CompletionRequest) (*anthropic.CompletionStream, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*anthropic.CompletionStream), args.Error(1)
}

func TestAnthropicProvider(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	config := &Config{
		APIKey:      "test-api-key",
		Model:       "claude-2",
		MaxTokens:   2000,
		Temperature: 0.7,
	}

	t.Run("basic functionality", func(t *testing.T) {
		provider := NewAnthropicProvider(logger, config)
		assert.NotNil(t, provider)
		assert.Equal(t, "anthropic", provider.Name())
		assert.Equal(t, 2000, provider.MaxTokens())
		assert.True(t, provider.SupportsStreaming())
	})

	t.Run("initialization", func(t *testing.T) {
		provider := NewAnthropicProvider(logger, config)
		err := provider.Initialize(ctx)
		assert.NoError(t, err)
	})

	t.Run("initialization with invalid config", func(t *testing.T) {
		testCases := []struct {
			name        string
			config      *Config
			shouldError bool
		}{
			{
				name: "missing API key",
				config: &Config{
					Model:       "claude-2",
					MaxTokens:   2000,
					Temperature: 0.7,
				},
				shouldError: true,
			},
			{
				name: "missing model",
				config: &Config{
					APIKey:      "test-api-key",
					MaxTokens:   2000,
					Temperature: 0.7,
				},
				shouldError: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				provider := NewAnthropicProvider(logger, tc.config)
				err := provider.Initialize(ctx)
				if tc.shouldError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("generate completion", func(t *testing.T) {
		provider := NewAnthropicProvider(logger, config)
		mockClient := new(MockAnthropicClient)
		provider.client = mockClient

		expectedResponse := &anthropic.CompletionResponse{
			Completion: "Test response",
			Usage: anthropic.Usage{
				PromptTokens:     10,
				CompletionTokens: 20,
				TotalTokens:      30,
			},
			StopReason: "stop",
		}

		mockClient.On("Complete", ctx, mock.Anything).Return(expectedResponse, nil)

		response, err := provider.Generate(ctx, "Test prompt")
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse.Completion, response.Content)
		assert.Equal(t, expectedResponse.Usage.PromptTokens, response.Usage.PromptTokens)
		assert.Equal(t, expectedResponse.Usage.CompletionTokens, response.Usage.CompletionTokens)
		assert.Equal(t, expectedResponse.Usage.TotalTokens, response.Usage.TotalTokens)
		assert.Equal(t, expectedResponse.StopReason, response.FinishReason)

		mockClient.AssertExpectations(t)
	})

	t.Run("generate with options", func(t *testing.T) {
		provider := NewAnthropicProvider(logger, config)
		mockClient := new(MockAnthropicClient)
		provider.client = mockClient

		expectedResponse := &anthropic.CompletionResponse{
			Completion: "Test response",
			Usage: anthropic.Usage{
				TotalTokens: 30,
			},
		}

		mockClient.On("Complete", ctx, mock.MatchedBy(func(req *anthropic.CompletionRequest) bool {
			return req.Temperature == 0.5 && req.MaxTokens == 1000
		})).Return(expectedResponse, nil)

		response, err := provider.Generate(ctx, "Test prompt",
			types.WithTemperature(0.5),
			types.WithMaxTokens(1000))
		assert.NoError(t, err)
		assert.NotNil(t, response)

		mockClient.AssertExpectations(t)
	})

	t.Run("stream completion", func(t *testing.T) {
		provider := NewAnthropicProvider(logger, config)
		mockClient := new(MockAnthropicClient)
		provider.client = mockClient

		mockStream := &anthropic.CompletionStream{}
		mockClient.On("CompleteStream", ctx, mock.Anything).Return(mockStream, nil)

		stream, err := provider.Stream(ctx, "Test prompt")
		assert.NoError(t, err)
		assert.NotNil(t, stream)

		// Note: We can't easily test the streaming functionality without
		// mocking the stream itself, which would require significant setup.
		// In a real application, this should be tested with integration tests.

		mockClient.AssertExpectations(t)
	})
}

func TestAnthropicProviderConfig(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		config := &Config{
			APIKey:      "test-api-key",
			Model:       "claude-2",
			MaxTokens:   2000,
			Temperature: 0.7,
		}

		assert.Equal(t, "test-api-key", config.APIKey)
		assert.Equal(t, "claude-2", config.Model)
		assert.Equal(t, 2000, config.MaxTokens)
		assert.Equal(t, 0.7, config.Temperature)
	})

	t.Run("config validation", func(t *testing.T) {
		logger := zaptest.NewLogger(t)
		ctx := context.Background()

		testCases := []struct {
			name        string
			config      *Config
			shouldError bool
		}{
			{
				name: "valid config",
				config: &Config{
					APIKey:      "test-api-key",
					Model:       "claude-2",
					MaxTokens:   2000,
					Temperature: 0.7,
				},
				shouldError: false,
			},
			{
				name: "empty API key",
				config: &Config{
					Model:       "claude-2",
					MaxTokens:   2000,
					Temperature: 0.7,
				},
				shouldError: true,
			},
			{
				name: "empty model",
				config: &Config{
					APIKey:      "test-api-key",
					MaxTokens:   2000,
					Temperature: 0.7,
				},
				shouldError: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				provider := NewAnthropicProvider(logger, tc.config)
				err := provider.Initialize(ctx)
				if tc.shouldError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
} 