package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pimentel/peppergo/internal/proxy"
	"github.com/pimentel/peppergo/pkg/types"
)

// Handler represents the HTTP API handler
type Handler struct {
	service *proxy.Service
}

// NewHandler creates a new API handler
func NewHandler(service *proxy.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// Router returns the HTTP router for the API
func (h *Handler) Router() http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// Routes
	r.Route("/v1", func(r chi.Router) {
		// Chat completion endpoint
		r.Post("/chat/completions", h.handleChat)
		
		// Provider management
		r.Get("/providers", h.handleListProviders)
	})

	return r
}

func (h *Handler) handleChat(w http.ResponseWriter, r *http.Request) {
	var req types.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get provider from header or query param
	provider := r.Header.Get("X-Provider")
	if provider == "" {
		provider = r.URL.Query().Get("provider")
	}
	if provider == "" {
		http.Error(w, "Provider not specified", http.StatusBadRequest)
		return
	}

	if req.Stream {
		h.handleStreamChat(w, r, provider, &req)
		return
	}

	resp, err := h.service.Chat(r.Context(), provider, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) handleStreamChat(w http.ResponseWriter, r *http.Request, provider string, req *types.ChatRequest) {
	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	respChan, err := h.service.StreamChat(r.Context(), provider, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for resp := range respChan {
		data, err := json.Marshal(resp)
		if err != nil {
			continue
		}
		
		// Write SSE format
		_, _ = w.Write([]byte("data: "))
		_, _ = w.Write(data)
		_, _ = w.Write([]byte("\n\n"))
		flusher.Flush()
	}
}

func (h *Handler) handleListProviders(w http.ResponseWriter, r *http.Request) {
	providers := h.service.ListProviders()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]string{
		"providers": providers,
	})
} 