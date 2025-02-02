package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/pimentel/peppergo/internal/api"
	"github.com/pimentel/peppergo/internal/proxy"
	"github.com/pimentel/peppergo/pkg/types"
	"github.com/stretchr/testify/suite"
	"bufio"
)

type ProxyTestSuite struct {
	suite.Suite
	server *httptest.Server
	proxy  *proxy.Service
}

func TestProxySuite(t *testing.T) {
	suite.Run(t, new(ProxyTestSuite))
}

func (s *ProxyTestSuite) SetupSuite() {
	// Create proxy service
	s.proxy = proxy.NewService()

	// Register test provider
	mockProvider := &MockProvider{
		name:    "mock",
		models:  []string{"test-model"},
		handler: s.mockCompletionHandler,
	}
	err := s.proxy.RegisterProvider(mockProvider)
	s.Require().NoError(err)

	// Create API handler
	handler := api.NewHandler(s.proxy)

	// Create test server
	s.server = httptest.NewServer(handler.Router())
}

func (s *ProxyTestSuite) TearDownSuite() {
	s.server.Close()
}

func (s *ProxyTestSuite) TestChatCompletion() {
	// Prepare request
	reqBody := types.ChatRequest{
		Model: "test-model",
		Messages: []types.Message{
			{
				Role:    "user",
				Content: "Hello!",
			},
		},
		Temperature: 0.7,
	}

	body, err := json.Marshal(reqBody)
	s.Require().NoError(err)

	// Create request
	req, err := http.NewRequest(http.MethodPost, s.server.URL+"/v1/chat/completions", bytes.NewBuffer(body))
	s.Require().NoError(err)

	// Set provider header
	req.Header.Set("X-Provider", "mock")
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	// Check response status
	s.Equal(http.StatusOK, resp.StatusCode)

	// Parse response
	var chatResp types.ChatResponse
	err = json.NewDecoder(resp.Body).Decode(&chatResp)
	s.Require().NoError(err)

	// Verify response
	s.Equal("mock", chatResp.Model)
	s.Len(chatResp.Choices, 1)
	s.Equal("Hello! I am a mock response.", chatResp.Choices[0].Message.Content)
}

func (s *ProxyTestSuite) TestStreamChatCompletion() {
	// Prepare request
	reqBody := types.ChatRequest{
		Model: "test-model",
		Messages: []types.Message{
			{
				Role:    "user",
				Content: "Hello!",
			},
		},
		Temperature: 0.7,
		Stream:      true,
	}

	body, err := json.Marshal(reqBody)
	s.Require().NoError(err)

	// Create request
	req, err := http.NewRequest(http.MethodPost, s.server.URL+"/v1/chat/completions", bytes.NewBuffer(body))
	s.Require().NoError(err)

	// Set provider header
	req.Header.Set("X-Provider", "mock")
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	// Check response status
	s.Equal(http.StatusOK, resp.StatusCode)
	s.Equal("text/event-stream", resp.Header.Get("Content-Type"))

	// Read SSE events
	scanner := bufio.NewScanner(resp.Body)
	var events []types.ChatResponse
	var fullMessage strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var event types.ChatResponse
		err = json.Unmarshal([]byte(data), &event)
		s.Require().NoError(err)
		events = append(events, event)
		
		// Accumulate the message content
		fullMessage.WriteString(event.Choices[0].Message.Content)
	}
	s.Require().NoError(scanner.Err())

	// Verify events
	s.NotEmpty(events)
	s.Equal("mock", events[0].Model)
	s.Len(events[0].Choices, 1)
	
	// Verify the complete accumulated message
	s.Equal("Hello! I am a mock response. ", fullMessage.String())
}

func (s *ProxyTestSuite) TestListProviders() {
	// Create request
	req, err := http.NewRequest(http.MethodGet, s.server.URL+"/v1/providers", nil)
	s.Require().NoError(err)

	// Send request
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	// Check response status
	s.Equal(http.StatusOK, resp.StatusCode)

	// Parse response
	var result map[string][]string
	err = json.NewDecoder(resp.Body).Decode(&result)
	s.Require().NoError(err)

	// Verify response
	s.Contains(result["providers"], "mock")
}

// MockProvider implements the types.Provider interface for testing
type MockProvider struct {
	name    string
	models  []string
	handler func(context.Context, *types.ChatRequest) (*types.ChatResponse, error)
}

func (p *MockProvider) Name() string {
	return p.name
}

func (p *MockProvider) AvailableModels() []string {
	return p.models
}

func (p *MockProvider) Chat(ctx context.Context, req *types.ChatRequest) (*types.ChatResponse, error) {
	return p.handler(ctx, req)
}

func (p *MockProvider) StreamChat(ctx context.Context, req *types.ChatRequest) (<-chan *types.ChatResponse, error) {
	responses := make(chan *types.ChatResponse)
	go func() {
		defer close(responses)

		// Get the base response
		resp, err := p.handler(ctx, req)
		if err != nil {
			return
		}

		// Split the response into chunks to simulate streaming
		message := resp.Choices[0].Message.Content
		words := strings.Split(message, " ")
		
		for i, word := range words {
			select {
			case <-ctx.Done():
				return
			default:
				// Create a streaming response for each word
				streamResp := &types.ChatResponse{
					ID:      fmt.Sprintf("%s-%d", resp.ID, i),
					Object:  "chat.completion.chunk",
					Created: resp.Created,
					Model:   resp.Model,
					Choices: []types.Choice{
						{
							Index: 0,
							Message: types.Message{
								Role:    "assistant",
								Content: word + " ",
							},
							FinishReason: "",
						},
					},
				}

				// Set finish reason for last chunk
				if i == len(words)-1 {
					streamResp.Choices[0].FinishReason = "stop"
				}

				responses <- streamResp
				time.Sleep(50 * time.Millisecond) // Simulate realistic streaming delay
			}
		}
	}()

	return responses, nil
}

func (s *ProxyTestSuite) mockCompletionHandler(ctx context.Context, req *types.ChatRequest) (*types.ChatResponse, error) {
	return &types.ChatResponse{
		ID:      "mock-response-id",
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   "mock",
		Choices: []types.Choice{
			{
				Index: 0,
				Message: types.Message{
					Role:    "assistant",
					Content: "Hello! I am a mock response.",
				},
				FinishReason: "stop",
			},
		},
		Usage: types.Usage{
			PromptTokens:     10,
			CompletionTokens: 10,
			TotalTokens:      20,
		},
	}, nil
} 