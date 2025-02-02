package agent

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/pimentel/peppergo/pkg/types"
)

// BaseAgent provides common functionality for all agents
type BaseAgent struct {
	id           string
	name         string
	version      string
	description  string
	provider     types.Provider
	capabilities map[string]types.Capability
	tools        map[string]types.Tool
	logger       *zap.Logger
	mu           sync.RWMutex
}

// NewBaseAgent creates a new BaseAgent instance
func NewBaseAgent(name, version, description string, logger *zap.Logger) *BaseAgent {
	return &BaseAgent{
		id:           uuid.New().String(),
		name:         name,
		version:      version,
		description:  description,
		capabilities: make(map[string]types.Capability),
		tools:        make(map[string]types.Tool),
		logger:       logger,
	}
}

// ID returns the agent's unique identifier
func (a *BaseAgent) ID() string {
	return a.id
}

// Name returns the agent's name
func (a *BaseAgent) Name() string {
	return a.name
}

// Version returns the agent's version
func (a *BaseAgent) Version() string {
	return a.version
}

// Initialize initializes the agent and its components
func (a *BaseAgent) Initialize(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Initialize provider
	if a.provider != nil {
		if err := a.provider.Initialize(ctx); err != nil {
			return fmt.Errorf("failed to initialize provider: %w", err)
		}
	}

	// Initialize capabilities
	for name, cap := range a.capabilities {
		if err := cap.Initialize(ctx); err != nil {
			return fmt.Errorf("failed to initialize capability %s: %w", name, err)
		}
	}

	// Initialize tools
	for name, tool := range a.tools {
		if err := tool.Initialize(ctx); err != nil {
			return fmt.Errorf("failed to initialize tool %s: %w", name, err)
		}
	}

	return nil
}

// Execute runs a task with the given input
func (a *BaseAgent) Execute(ctx context.Context, task string, opts ...types.ExecuteOption) (*types.Response, error) {
	if a.provider == nil {
		return nil, fmt.Errorf("no provider configured")
	}

	// Apply options
	options := &types.ExecuteOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// Generate response
	return a.provider.Generate(ctx, task, types.WithTemperature(options.Temperature))
}

// Cleanup performs cleanup of the agent and its components
func (a *BaseAgent) Cleanup(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	var errs []error

	// Cleanup provider
	if a.provider != nil {
		if err := a.provider.Initialize(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to cleanup provider: %w", err))
		}
	}

	// Cleanup capabilities
	for name, cap := range a.capabilities {
		if err := cap.Cleanup(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to cleanup capability %s: %w", name, err))
		}
	}

	// Cleanup tools
	for name, tool := range a.tools {
		if err := tool.Cleanup(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to cleanup tool %s: %w", name, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("cleanup errors: %v", errs)
	}
	return nil
}

// AddCapability adds a new capability to the agent
func (a *BaseAgent) AddCapability(capability types.Capability) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	name := capability.Name()
	if _, exists := a.capabilities[name]; exists {
		return fmt.Errorf("capability %s already exists", name)
	}

	// Check requirements
	reqs := capability.Requirements()
	if reqs != nil {
		// Check required tools
		for _, toolName := range reqs.Tools {
			if _, exists := a.tools[toolName]; !exists {
				return fmt.Errorf("missing required tool %s for capability %s", toolName, name)
			}
		}

		// Check required capabilities
		for _, capName := range reqs.Capabilities {
			if _, exists := a.capabilities[capName]; !exists {
				return fmt.Errorf("missing required capability %s for capability %s", capName, name)
			}
		}
	}

	a.capabilities[name] = capability
	return nil
}

// AddTool adds a new tool to the agent
func (a *BaseAgent) AddTool(tool types.Tool) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	name := tool.Name()
	if _, exists := a.tools[name]; exists {
		return fmt.Errorf("tool %s already exists", name)
	}

	a.tools[name] = tool
	return nil
}

// UseProvider sets the AI provider for this agent
func (a *BaseAgent) UseProvider(provider types.Provider) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	a.provider = provider
	return nil
} 