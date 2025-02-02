package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"github.com/pimentel/peppergo/pkg/types"
)

type OpenRouterConfig struct {
	APIKey      string
	Model       string
	MaxTokens   int
	Temperature float64
	RateLimiter *rate.Limiter
}

type OpenRouterProvider struct {
	config *OpenRouterConfig
	client *http.Client
	logger *zap.Logger
}

func NewOpenRouterProvider(logger *zap.Logger, config *OpenRouterConfig) *OpenRouterProvider {
	return &OpenRouterProvider{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

func (p *OpenRouterProvider) Initialize(ctx context.Context) error {
	if p.config.APIKey == "" {
		return fmt.Errorf("API key is required")
	}
	if p.config.Model == "" {
		return fmt.Errorf("model is required")
	}
	if p.config.Temperature < 0 || p.config.Temperature > 1 {
		return fmt.Errorf("invalid temperature: must be between 0 and 1")
	}
	if p.config.MaxTokens < 1 {
		return fmt.Errorf("invalid max tokens: must be greater than 0")
	}
	return nil
}

type generateRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type generateResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func (p *OpenRouterProvider) Generate(ctx context.Context, prompt string, opts ...types.ExecuteOption) (*types.Response, error) {
	if prompt == "" {
		return nil, fmt.Errorf("empty prompt")
	}

	// Apply rate limiting
	if p.config.RateLimiter != nil {
		err := p.config.RateLimiter.Wait(ctx)
		if err != nil {
			return nil, fmt.Errorf("rate limit exceeded: %w", err)
		}
	}

	// Apply options
	options := &types.ExecuteOptions{
		Temperature: p.config.Temperature,
		MaxTokens:   p.config.MaxTokens,
		Model:       p.config.Model,
	}
	for _, opt := range opts {
		opt(options)
	}

	// Get retries from metadata
	retries := 1
	if options.Metadata != nil {
		if r, ok := options.Metadata["retries"].(int); ok {
			retries = r
		}
	}

	// Validate options
	if options.Temperature < 0 || options.Temperature > 1 {
		return nil, fmt.Errorf("invalid temperature: must be between 0 and 1")
	}
	if options.MaxTokens < 1 {
		return nil, fmt.Errorf("invalid max tokens: must be greater than 0")
	}

	reqBody := generateRequest{
		Model: options.Model,
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
		MaxTokens:   options.MaxTokens,
		Temperature: options.Temperature,
	}

	var lastErr error
	for attempt := 0; attempt < retries; attempt++ {
		if attempt > 0 {
			p.logger.Info("retrying request", 
				zap.Int("attempt", attempt+1),
				zap.Int("max_attempts", retries))
			// Wait before retry
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		response, err := p.makeRequest(ctx, reqBody)
		if err == nil {
			return response, nil
		}
		lastErr = err
		
		// Don't retry on context cancellation or validation errors
		if ctx.Err() != nil || isValidationError(err) {
			return nil, err
		}
	}

	return nil, fmt.Errorf("all attempts failed: %w", lastErr)
}

func (p *OpenRouterProvider) makeRequest(ctx context.Context, reqBody generateRequest) (*types.Response, error) {
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set required headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.config.APIKey))
	req.Header.Set("HTTP-Referer", "https://github.com/pimentel/peppergo")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var genResp generateResponse
	if err := json.Unmarshal(body, &genResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(genResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	return &types.Response{
		Content: genResp.Choices[0].Message.Content,
		Usage: types.Usage{
			PromptTokens:     genResp.Usage.PromptTokens,
			CompletionTokens: genResp.Usage.CompletionTokens,
			TotalTokens:      genResp.Usage.TotalTokens,
		},
		FinishReason: genResp.Choices[0].FinishReason,
		Timestamp:    time.Now().Unix(),
	}, nil
}

func isValidationError(err error) bool {
	return strings.Contains(err.Error(), "invalid") ||
		strings.Contains(err.Error(), "empty prompt")
} 