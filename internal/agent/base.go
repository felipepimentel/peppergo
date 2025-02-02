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

// NewBaseAgent creates a new base agent with the given parameters
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

// Initialize sets up the agent
func (a *BaseAgent) Initialize(ctx context.Context) error {
	return nil
}

// Execute processes a task using the agent's capabilities
func (a *BaseAgent) Execute(ctx context.Context, task string, opts ...types.ExecuteOption) (*types.Response, error) {
	if a.provider == nil {
		return nil, fmt.Errorf("no provider configured")
	}

	// TODO: Implement task execution logic
	return &types.Response{
		Content: "Task execution not implemented",
	}, nil
}

// Cleanup performs any necessary cleanup
func (a *BaseAgent) Cleanup(ctx context.Context) error {
	return nil
}

// AddCapability adds a new capability to the agent
func (a *BaseAgent) AddCapability(capability types.Capability) error {
	if capability == nil {
		return fmt.Errorf("capability cannot be nil")
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	name := capability.Name()
	if _, exists := a.capabilities[name]; exists {
		return fmt.Errorf("capability %s already exists", name)
	}

	a.capabilities[name] = capability
	return nil
}

// AddTool adds a new tool to the agent
func (a *BaseAgent) AddTool(tool types.Tool) error {
	if tool == nil {
		return fmt.Errorf("tool cannot be nil")
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	name := tool.Name()
	if _, exists := a.tools[name]; exists {
		return fmt.Errorf("tool %s already exists", name)
	}

	a.tools[name] = tool
	return nil
}

// UseProvider sets the provider for the agent
func (a *BaseAgent) UseProvider(provider types.Provider) error {
	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	a.provider = provider
	return nil
} 