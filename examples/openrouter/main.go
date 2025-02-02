package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"github.com/pimentel/peppergo/internal/provider"
	"github.com/pimentel/peppergo/pkg/types"
)

type Handler struct {
	provider provider.Provider
	logger   *zap.Logger
}

func NewHandler(p provider.Provider, logger *zap.Logger) *Handler {
	return &Handler{
		provider: p,
		logger:   logger,
	}
}

func (h *Handler) HandleCompletion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Simple example - in production you'd want to parse the request body
	prompt := "What is the capital of France?"
	
	ctx := r.Context()
	resp, err := h.provider.Generate(ctx, prompt, 
		types.WithTemperature(0.7),
		types.WithMaxTokens(100),
		types.WithRetries(3),
	)
	if err != nil {
		h.logger.Error("Failed to generate response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// In production you'd want to format this as JSON
	fmt.Fprintf(w, "Prompt: %s\nResponse: %s\nTokens Used: %d\nFinish Reason: %s\n",
		prompt, resp.Content, resp.Usage.TotalTokens, resp.FinishReason)
}

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Error loading .env file: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Printf("Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Get API key from environment
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		logger.Fatal("OPENROUTER_API_KEY environment variable is required")
	}

	// Get model from environment or use default
	model := os.Getenv("OPENROUTER_MODEL")
	if model == "" {
		model = "openai/gpt-3.5-turbo"
	}

	// Create rate limiter (3 requests per minute)
	limiter := rate.NewLimiter(rate.Every(20*time.Second), 1)

	// Configure provider
	config := &provider.OpenRouterConfig{
		APIKey:      apiKey,
		Model:       model,
		MaxTokens:   2000,
		Temperature: 0.7,
		RateLimiter: limiter,
	}

	// Create provider
	p := provider.NewOpenRouterProvider(logger, config)

	// Initialize provider
	if err := p.Initialize(context.Background()); err != nil {
		logger.Fatal("Failed to initialize provider", zap.Error(err))
	}

	// Create handler
	handler := NewHandler(p, logger)

	// Set up HTTP server
	http.HandleFunc("/v1/completions", handler.HandleCompletion)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	addr := fmt.Sprintf(":%s", port)
	logger.Info("Starting server", zap.String("address", addr))
	if err := http.ListenAndServe(addr, nil); err != nil {
		logger.Fatal("Server failed", zap.Error(err))
	}
} 