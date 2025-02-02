package capability

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/yourusername/peppergo/pkg/types"
)

// BasicChatCapability provides basic chat functionality
type BasicChatCapability struct {
	logger *zap.Logger
	config *Config
}

// Config represents the configuration for BasicChatCapability
type Config struct {
	// MaxTokens is the maximum number of tokens to generate
	MaxTokens int `yaml:"max_tokens"`

	// Temperature controls response randomness
	Temperature float64 `yaml:"temperature"`

	// SystemPrompt is the system prompt to use
	SystemPrompt string `yaml:"system_prompt"`
}

// NewBasicChatCapability creates a new BasicChatCapability instance
func NewBasicChatCapability(logger *zap.Logger, config *Config) *BasicChatCapability {
	return &BasicChatCapability{
		logger: logger,
		config: config,
	}
}

// Name returns the capability's name
func (c *BasicChatCapability) Name() string {
	return "basic_chat"
}

// Description returns the capability's description
func (c *BasicChatCapability) Description() string {
	return "Provides basic chat functionality"
}

// Initialize initializes the capability
func (c *BasicChatCapability) Initialize(ctx context.Context) error {
	c.logger.Info("Initializing basic chat capability",
		zap.Int("max_tokens", c.config.MaxTokens),
		zap.Float64("temperature", c.config.Temperature))
	return nil
}

// Execute runs the capability
func (c *BasicChatCapability) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	prompt, ok := input.(string)
	if !ok {
		return nil, fmt.Errorf("input must be a string")
	}

	c.logger.Debug("Executing basic chat capability",
		zap.String("prompt", prompt))

	// Format the prompt with system instructions
	formattedPrompt := fmt.Sprintf("%s\n\nUser: %s\nAssistant:", c.config.SystemPrompt, prompt)

	return map[string]interface{}{
		"formatted_prompt": formattedPrompt,
		"max_tokens":      c.config.MaxTokens,
		"temperature":     c.config.Temperature,
	}, nil
}

// Cleanup performs cleanup
func (c *BasicChatCapability) Cleanup(ctx context.Context) error {
	return nil
}

// Requirements returns capability requirements
func (c *BasicChatCapability) Requirements() *types.Requirements {
	reqs := types.NewRequirements()
	reqs.SetMinTokens(c.config.MaxTokens)
	return reqs
}

// Version returns the capability version
func (c *BasicChatCapability) Version() string {
	return "1.0.0"
}

// Example YAML configuration:
/*
name: basic_chat
version: "1.0.0"
description: "Basic chat capability"

config:
  max_tokens: 2000
  temperature: 0.7
  system_prompt: |
    You are a helpful AI assistant.
    Please provide clear and concise responses.
*/ 