package capability

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestBasicChatCapability(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	config := &Config{
		MaxTokens:    2000,
		Temperature:  0.7,
		SystemPrompt: "You are a helpful AI assistant.",
	}

	t.Run("basic functionality", func(t *testing.T) {
		cap := NewBasicChatCapability(logger, config)
		assert.NotNil(t, cap)
		assert.Equal(t, "basic_chat", cap.Name())
		assert.Equal(t, "1.0.0", cap.Version())
	})

	t.Run("initialization", func(t *testing.T) {
		cap := NewBasicChatCapability(logger, config)
		err := cap.Initialize(ctx)
		assert.NoError(t, err)
	})

	t.Run("execute with valid input", func(t *testing.T) {
		cap := NewBasicChatCapability(logger, config)
		input := "Hello, how are you?"

		result, err := cap.Execute(ctx, input)
		assert.NoError(t, err)

		response, ok := result.(map[string]interface{})
		assert.True(t, ok)
		assert.Contains(t, response["formatted_prompt"], input)
		assert.Contains(t, response["formatted_prompt"], config.SystemPrompt)
		assert.Equal(t, config.MaxTokens, response["max_tokens"])
		assert.Equal(t, config.Temperature, response["temperature"])
	})

	t.Run("execute with invalid input", func(t *testing.T) {
		cap := NewBasicChatCapability(logger, config)
		input := 123 // Invalid type

		result, err := cap.Execute(ctx, input)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "input must be a string")
	})

	t.Run("requirements", func(t *testing.T) {
		cap := NewBasicChatCapability(logger, config)
		reqs := cap.Requirements()

		assert.NotNil(t, reqs)
		assert.Equal(t, config.MaxTokens, reqs.MinTokens)
		assert.Empty(t, reqs.Tools)
		assert.Empty(t, reqs.Capabilities)
	})

	t.Run("cleanup", func(t *testing.T) {
		cap := NewBasicChatCapability(logger, config)
		err := cap.Cleanup(ctx)
		assert.NoError(t, err)
	})

	t.Run("prompt formatting", func(t *testing.T) {
		testCases := []struct {
			name     string
			prompt   string
			expected string
		}{
			{
				name:   "simple prompt",
				prompt: "Hello",
				expected: "You are a helpful AI assistant.\n\n" +
					"User: Hello\n" +
					"Assistant:",
			},
			{
				name:   "multi-line prompt",
				prompt: "Hello\nHow are you?",
				expected: "You are a helpful AI assistant.\n\n" +
					"User: Hello\nHow are you?\n" +
					"Assistant:",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				cap := NewBasicChatCapability(logger, config)
				result, err := cap.Execute(ctx, tc.prompt)
				assert.NoError(t, err)

				response := result.(map[string]interface{})
				assert.Equal(t, tc.expected, response["formatted_prompt"])
			})
		}
	})

	t.Run("config validation", func(t *testing.T) {
		testCases := []struct {
			name        string
			config      *Config
			shouldError bool
		}{
			{
				name: "valid config",
				config: &Config{
					MaxTokens:    2000,
					Temperature:  0.7,
					SystemPrompt: "Test prompt",
				},
				shouldError: false,
			},
			{
				name: "zero max tokens",
				config: &Config{
					MaxTokens:    0,
					Temperature:  0.7,
					SystemPrompt: "Test prompt",
				},
				shouldError: false, // Currently not validated
			},
			{
				name: "invalid temperature",
				config: &Config{
					MaxTokens:    2000,
					Temperature:  2.0,
					SystemPrompt: "Test prompt",
				},
				shouldError: false, // Currently not validated
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				cap := NewBasicChatCapability(logger, tc.config)
				err := cap.Initialize(ctx)
				if tc.shouldError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}

func TestBasicChatCapabilityConfig(t *testing.T) {
	t.Run("yaml marshaling", func(t *testing.T) {
		config := &Config{
			MaxTokens:    2000,
			Temperature:  0.7,
			SystemPrompt: "Test prompt",
		}

		// Add YAML marshaling tests when needed
		assert.NotNil(t, config)
	})
} 