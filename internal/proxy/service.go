package proxy

import (
	"context"
	"fmt"
	"sync"

	"github.com/pimentel/peppergo/pkg/types"
)

// Service represents the LLM proxy service
type Service struct {
	providers map[string]types.Provider
	mu        sync.RWMutex
}

// NewService creates a new proxy service
func NewService() *Service {
	return &Service{
		providers: make(map[string]types.Provider),
	}
}

// RegisterProvider registers a new provider with the service
func (s *Service) RegisterProvider(provider types.Provider) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	name := provider.Name()
	if _, exists := s.providers[name]; exists {
		return fmt.Errorf("provider %s already registered", name)
	}

	s.providers[name] = provider
	return nil
}

// GetProvider returns a provider by name
func (s *Service) GetProvider(name string) (types.Provider, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	provider, exists := s.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}

	return provider, nil
}

// Chat handles a chat completion request
func (s *Service) Chat(ctx context.Context, providerName string, req *types.ChatRequest) (*types.ChatResponse, error) {
	provider, err := s.GetProvider(providerName)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	// Here we could add request normalization if needed
	resp, err := provider.Chat(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("provider %s chat failed: %w", providerName, err)
	}

	// Here we could add response normalization if needed
	return resp, nil
}

// StreamChat handles a streaming chat completion request
func (s *Service) StreamChat(ctx context.Context, providerName string, req *types.ChatRequest) (<-chan *types.ChatResponse, error) {
	provider, err := s.GetProvider(providerName)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	// Here we could add request normalization if needed
	respChan, err := provider.StreamChat(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("provider %s stream chat failed: %w", providerName, err)
	}

	// Create a new channel for normalized responses
	normalizedChan := make(chan *types.ChatResponse)

	// Start a goroutine to normalize responses
	go func() {
		defer close(normalizedChan)
		for resp := range respChan {
			// Here we could add response normalization if needed
			normalizedChan <- resp
		}
	}()

	return normalizedChan, nil
}

// ListProviders returns a list of registered providers
func (s *Service) ListProviders() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	providers := make([]string, 0, len(s.providers))
	for name := range s.providers {
		providers = append(providers, name)
	}
	return providers
} 