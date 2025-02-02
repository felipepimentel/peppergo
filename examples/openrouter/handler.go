package main

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/yourusername/peppergo/internal/provider"
	"github.com/yourusername/peppergo/pkg/types"
)

type CompletionRequest struct {
	Prompt      string  `json:"prompt"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
}

type CompletionResponse struct {
	Content      string `json:"content"`
	TotalTokens  int    `json:"total_tokens"`
	FinishReason string `json:"finish_reason"`
}

type Handler struct {
	provider *provider.OpenRouterProvider
	logger   *zap.Logger
}

func NewHandler(p *provider.OpenRouterProvider, logger *zap.Logger) *Handler {
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

	var req CompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Prompt == "" {
		http.Error(w, "Prompt is required", http.StatusBadRequest)
		return
	}

	// Prepare options
	opts := []types.ExecuteOption{}
	if req.MaxTokens > 0 {
		opts = append(opts, types.WithMaxTokens(req.MaxTokens))
	}
	if req.Temperature >= 0 && req.Temperature <= 1 {
		opts = append(opts, types.WithTemperature(req.Temperature))
	}

	// Generate completion
	response, err := h.provider.Generate(r.Context(), req.Prompt, opts...)
	if err != nil {
		h.logger.Error("failed to generate completion", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Prepare response
	resp := CompletionResponse{
		Content:      response.Content,
		TotalTokens:  response.Usage.TotalTokens,
		FinishReason: response.FinishReason,
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
} 