package agent

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/yourusername/peppergo/pkg/types"
)

// ExampleAgent demonstrates how to implement a custom agent
type ExampleAgent struct {
	*BaseAgent
	customSetting string
}

// NewExampleAgent creates a new ExampleAgent instance
func NewExampleAgent(logger *zap.Logger) *ExampleAgent {
	base := NewBaseAgent(
		"example-agent",
		"1.0.0",
		"An example agent implementation",
		logger,
	)

	return &ExampleAgent{
		BaseAgent:      base,
		customSetting: "default",
	}
}

// Execute overrides the base Execute method to add custom behavior
func (a *ExampleAgent) Execute(ctx context.Context, task string, opts ...types.ExecuteOption) (*types.Response, error) {
	// Log the incoming task
	a.logger.Info("Executing task",
		zap.String("task", task),
		zap.String("agent", a.Name()),
		zap.String("custom_setting", a.customSetting))

	// Use capabilities if needed
	for name, cap := range a.capabilities {
		a.logger.Debug("Using capability",
			zap.String("name", name),
			zap.String("version", cap.Version()))

		result, err := cap.Execute(ctx, task)
		if err != nil {
			return nil, fmt.Errorf("capability %s failed: %w", name, err)
		}

		// Process capability result
		a.logger.Debug("Capability result",
			zap.String("name", name),
			zap.Any("result", result))
	}

	// Use tools if needed
	for name, tool := range a.tools {
		a.logger.Debug("Using tool",
			zap.String("name", name),
			zap.String("version", tool.Version()))

		result, err := tool.Execute(ctx, map[string]interface{}{
			"task": task,
		})
		if err != nil {
			return nil, fmt.Errorf("tool %s failed: %w", name, err)
		}

		// Process tool result
		a.logger.Debug("Tool result",
			zap.String("name", name),
			zap.Any("result", result))
	}

	// Call base implementation for provider interaction
	return a.BaseAgent.Execute(ctx, task, opts...)
}

// SetCustomSetting sets the custom setting
func (a *ExampleAgent) SetCustomSetting(value string) {
	a.customSetting = value
}

// Example YAML configuration for this agent:
/*
name: example-agent
version: "1.0.0"
description: "An example agent implementation"

capabilities:
  - basic_chat
  - code_review

tools:
  - file_reader
  - code_analyzer

role:
  name: "Example Agent"
  description: "Demonstrates agent implementation"
  instructions: |
    You are an example agent that demonstrates
    how to implement custom behavior while using
    the base agent functionality.

settings:
  custom_setting: "custom value"

metadata:
  author: "PepperGo Team"
  tags: ["example", "demo"]
*/ 