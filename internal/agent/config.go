package agent

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v3"
	"go.uber.org/zap"

	"github.com/pimentel/peppergo/pkg/types"
)

// Config represents an agent configuration
type Config struct {
	// Name is the agent's name
	Name string `yaml:"name"`

	// Version is the agent's version
	Version string `yaml:"version"`

	// Description is a human-readable description
	Description string `yaml:"description"`

	// Capabilities lists the agent's capabilities
	Capabilities []string `yaml:"capabilities"`

	// Tools lists the agent's tools
	Tools []string `yaml:"tools"`

	// Role defines the agent's role
	Role *RoleConfig `yaml:"role"`

	// Settings contains agent-specific settings
	Settings map[string]interface{} `yaml:"settings"`

	// Metadata contains additional metadata
	Metadata map[string]interface{} `yaml:"metadata"`
}

// RoleConfig represents an agent's role configuration
type RoleConfig struct {
	// Name is the role's name
	Name string `yaml:"name"`

	// Description is a human-readable description
	Description string `yaml:"description"`

	// Instructions are the role-specific instructions
	Instructions string `yaml:"instructions"`
}

// LoadFromYAML loads an agent configuration from a YAML file
func LoadFromYAML(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

// FromYAML creates a new agent from a YAML configuration file
func FromYAML(path string, logger *zap.Logger, registry *Registry) (*BaseAgent, error) {
	config, err := LoadFromYAML(path)
	if err != nil {
		return nil, err
	}

	agent := NewBaseAgent(config.Name, config.Version, config.Description, logger)

	// Add capabilities
	for _, name := range config.Capabilities {
		capability, err := registry.GetCapability(name)
		if err != nil {
			return nil, fmt.Errorf("failed to get capability %s: %w", name, err)
		}
		if err := agent.AddCapability(capability); err != nil {
			return nil, fmt.Errorf("failed to add capability %s: %w", name, err)
		}
	}

	// Add tools
	for _, name := range config.Tools {
		tool, err := registry.GetTool(name)
		if err != nil {
			return nil, fmt.Errorf("failed to get tool %s: %w", name, err)
		}
		if err := agent.AddTool(tool); err != nil {
			return nil, fmt.Errorf("failed to add tool %s: %w", name, err)
		}
	}

	return agent, nil
}

// Registry manages available capabilities and tools
type Registry struct {
	capabilities map[string]types.Capability
	tools        map[string]types.Tool
	logger       *zap.Logger
}

// NewRegistry creates a new Registry instance
func NewRegistry(logger *zap.Logger) *Registry {
	return &Registry{
		capabilities: make(map[string]types.Capability),
		tools:        make(map[string]types.Tool),
		logger:       logger,
	}
}

// RegisterCapability registers a capability
func (r *Registry) RegisterCapability(capability types.Capability) error {
	name := capability.Name()
	if _, exists := r.capabilities[name]; exists {
		return fmt.Errorf("capability %s already registered", name)
	}
	r.capabilities[name] = capability
	return nil
}

// RegisterTool registers a tool
func (r *Registry) RegisterTool(tool types.Tool) error {
	name := tool.Name()
	if _, exists := r.tools[name]; exists {
		return fmt.Errorf("tool %s already registered", name)
	}
	r.tools[name] = tool
	return nil
}

// GetCapability returns a registered capability
func (r *Registry) GetCapability(name string) (types.Capability, error) {
	capability, exists := r.capabilities[name]
	if !exists {
		return nil, fmt.Errorf("capability %s not found", name)
	}
	return capability, nil
}

// GetTool returns a registered tool
func (r *Registry) GetTool(name string) (types.Tool, error) {
	tool, exists := r.tools[name]
	if !exists {
		return nil, fmt.Errorf("tool %s not found", name)
	}
	return tool, nil
}

// LoadCapabilitiesFromDir loads all capability configurations from a directory
func (r *Registry) LoadCapabilitiesFromDir(dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return fmt.Errorf("failed to list capability files: %w", err)
	}

	for _, file := range files {
		if err := r.loadCapabilityFromFile(file); err != nil {
			r.logger.Error("Failed to load capability",
				zap.String("file", file),
				zap.Error(err))
		}
	}

	return nil
}

// LoadToolsFromDir loads all tool configurations from a directory
func (r *Registry) LoadToolsFromDir(dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return fmt.Errorf("failed to list tool files: %w", err)
	}

	for _, file := range files {
		if err := r.loadToolFromFile(file); err != nil {
			r.logger.Error("Failed to load tool",
				zap.String("file", file),
				zap.Error(err))
		}
	}

	return nil
}

func (r *Registry) loadCapabilityFromFile(path string) error {
	// Implementation depends on capability factory system
	return nil
}

func (r *Registry) loadToolFromFile(path string) error {
	// Implementation depends on tool factory system
	return nil
} 