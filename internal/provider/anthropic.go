package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/anthropic-ai/anthropic-sdk-go"
	"go.uber.org/zap"

	"github.com/yourusername/peppergo/pkg/types"
)

// AnthropicProvider provides integration with Anthropic's Claude
type AnthropicProvider struct {
	client    *anthropic.Client
	logger    *zap.Logger
	config    *Config
	maxTokens int
}

// Config represents the configuration for AnthropicProvider
type Config struct {
	// APIKey is the Anthropic API key
	APIKey string `yaml:"api_key"`

	// Model is the model to use (e.g., "claude-2")
	Model string `yaml:"model"`

	// MaxTokens is the maximum number of tokens to generate
	MaxTokens int `yaml:"max_tokens"`

	// Temperature controls response randomness
	Temperature float64 `yaml:"temperature"`
}

// NewAnthropicProvider creates a new AnthropicProvider instance
func NewAnthropicProvider(logger *zap.Logger, config *Config) *AnthropicProvider {
	return &AnthropicProvider{
		logger:    logger,
		config:    config,
		maxTokens: config.MaxTokens,
	}
}

// Initialize initializes the provider
func (p *AnthropicProvider) Initialize(ctx context.Context) error {
	// Validate config
	if p.config.APIKey == "" {
		return fmt.Errorf("API key is required")
	}

	if p.config.Model == "" {
		return fmt.Errorf("model is required")
	}

	// Create client
	p.client = anthropic.NewClient(p.config.APIKey)

	p.logger.Info("Initialized Anthropic provider",
		zap.String("model", p.config.Model),
		zap.Int("max_tokens", p.config.MaxTokens))

	return nil
}

// Generate generates a response for the given prompt
func (p *AnthropicProvider) Generate(ctx context.Context, prompt string, opts ...types.GenerateOption) (*types.Response, error) {
	// Apply options
	options := &types.GenerateOptions{
		Temperature: p.config.Temperature,
		MaxTokens:   p.config.MaxTokens,
		Model:      p.config.Model,
	}
	for _, opt := range opts {
		opt(options)
	}

	// Create completion request
	req := &anthropic.CompletionRequest{
		Prompt:      prompt,
		Model:       options.Model,
		MaxTokens:   options.MaxTokens,
		Temperature: options.Temperature,
	}

	// Generate completion
	resp, err := p.client.Complete(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate completion: %w", err)
	}

	p.logger.Debug("Generated completion",
		zap.String("model", options.Model),
		zap.Int("tokens", resp.Usage.TotalTokens))

	return &types.Response{
		Content: resp.Completion,
		Usage: &types.Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
		Timestamp:    time.Now().Unix(),
		FinishReason: resp.StopReason,
	}, nil
}

// Stream streams responses for the given prompt
func (p *AnthropicProvider) Stream(ctx context.Context, prompt string) (<-chan types.Response, error) {
	responseChan := make(chan types.Response)

	go func() {
		defer close(responseChan)

		// Create completion request
		req := &anthropic.CompletionRequest{
			Prompt:      prompt,
			Model:       p.config.Model,
			MaxTokens:   p.config.MaxTokens,
			Temperature: p.config.Temperature,
			Stream:     true,
		}

		// Generate streaming completion
		stream, err := p.client.CompleteStream(ctx, req)
		if err != nil {
			p.logger.Error("Failed to create completion stream",
				zap.Error(err))
			return
		}
		defer stream.Close()

		for {
			select {
			case <-ctx.Done():
				return
			default:
				event, err := stream.Recv()
				if err != nil {
					p.logger.Error("Failed to receive from stream",
						zap.Error(err))
					return
				}

				responseChan <- types.Response{
					Content:    event.Completion,
					Timestamp: time.Now().Unix(),
				}

				if event.Done {
					return
				}
			}
		}
	}()

	return responseChan, nil
}

// Name returns the provider's name
func (p *AnthropicProvider) Name() string {
	return "anthropic"
}

// MaxTokens returns the maximum tokens supported
func (p *AnthropicProvider) MaxTokens() int {
	return p.maxTokens
}

// SupportsStreaming returns whether streaming is supported
func (p *AnthropicProvider) SupportsStreaming() bool {
	return true
}

// Example YAML configuration:
/*
name: anthropic
version: "1.0.0"
description: "Anthropic Claude provider"

config:
  api_key: "your-api-key"
  model: "claude-2"
  max_tokens: 4096
  temperature: 0.7
*/ 