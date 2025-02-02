package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pimentel/peppergo/internal/api"
	"github.com/pimentel/peppergo/internal/provider"
	"github.com/pimentel/peppergo/internal/proxy"
)

func main() {
	// Create proxy service
	proxyService := proxy.NewService()

	// Register providers
	openRouterProvider := provider.NewOpenRouter()
	if err := proxyService.RegisterProvider(openRouterProvider); err != nil {
		log.Fatalf("Failed to register OpenRouter provider: %v", err)
	}

	// Create API handler
	handler := api.NewHandler(proxyService)

	// Create HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: handler.Router(),
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
} 