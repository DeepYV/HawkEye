// HawkEye — Real-Time Frustration Detection Engine
//
// Single binary, single port. Run with --dev for local development
// with in-memory storage and no external dependencies.
//
// Usage:
//
//	go run ./cmd/hawkeye --dev
//	go run ./cmd/hawkeye --port 8080 --api-key my-key
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/your-org/frustration-engine/internal/app"
	"github.com/your-org/frustration-engine/internal/config"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cfg := config.Load()
	cfg.Print()

	application := app.New(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	application.Start(ctx)

	// Graceful shutdown on SIGINT / SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigCh
		fmt.Println("\n[hawkeye] shutting down...")
		application.Stop()
		cancel()
	}()

	addr := ":" + cfg.Port
	if err := application.Server.ListenAndServe(addr); err != nil {
		log.Printf("[hawkeye] server stopped: %v", err)
	}

	fmt.Println("[hawkeye] stopped.")
}
