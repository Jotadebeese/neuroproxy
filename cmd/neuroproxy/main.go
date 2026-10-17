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

	"github.com/Jotadebeese/neuroproxy/internal/ai"
	"github.com/Jotadebeese/neuroproxy/internal/cache"
	"github.com/Jotadebeese/neuroproxy/internal/config"
	"github.com/Jotadebeese/neuroproxy/internal/contracts"
	"github.com/Jotadebeese/neuroproxy/internal/engine"
	"github.com/Jotadebeese/neuroproxy/internal/gateway"
)

func main() {
	cfg := config.Load()

	// Initialize Storage (Google Cloud Memorystore Redis or in-memory LRU)
	storage := cache.NewRuleStorage(cfg.RedisURL)

	// Initialize Circuit Breaker & Gemini Healer
	circuitBreaker := ai.NewCircuitBreaker(0.5, 10, 5, 15*time.Second)
	healer := ai.NewGeminiHealer(circuitBreaker)

	// Initialize Engine
	eng := engine.NewEngine(storage, healer)

	// Seed demo traffic for instant Mission Control dashboard data
	seedDemoTraffic(eng)

	// Initialize Gateway Router
	gw := gateway.NewGateway(cfg, eng)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      gw.Handler(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 25 * time.Second,
	}

	fmt.Printf("\n⚡ NeuroProxy v2.0 Self-Healing API Gateway running on http://localhost:%s\n", cfg.Port)
	fmt.Printf("   ├─ Mission Control UI : http://localhost:%s/\n", cfg.Port)
	fmt.Printf("   ├─ Proxy Ingestion    : POST http://localhost:%s/api/proxy/{checkout|inventory|telemetry}\n", cfg.Port)
	fmt.Printf("   ├─ Middleware Export  : GET  http://localhost:%s/api/export/middleware.go\n", cfg.Port)
	fmt.Printf("   ├─ Cloud Project      : %s\n", cfg.GCPProject)
	fmt.Printf("   └─ Gemini API Key     : %v\n\n", cfg.GeminiAPIKey != "")

	// Graceful shutdown handling
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-stop
	fmt.Println("\nShutting down NeuroProxy gateway gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
	fmt.Println("NeuroProxy stopped.")
}

func seedDemoTraffic(eng *engine.Engine) {
	ctx := context.Background()
	presets := contracts.GetChaosPresets()
	if len(presets) > 0 {
		_, _ = eng.ProcessRequest(ctx, presets[0].ServiceID, presets[0].Payload, true)
		_, _ = eng.ProcessRequest(ctx, presets[0].ServiceID, presets[0].Payload, true)
	}
}
