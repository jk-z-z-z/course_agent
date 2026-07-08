package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"course_agent_backend/internal/bootstrap"
	"course_agent_backend/internal/config"
)

func main() {
	cfgPath := os.Getenv("APP_CONFIG")
	if cfgPath == "" {
		cfgPath = filepath.Join("configs", "config.example.yaml")
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	app, err := bootstrap.New(cfg)
	if err != nil {
		log.Fatalf("bootstrap failed: %v", err)
	}
	defer func() {
		if closeErr := app.Close(); closeErr != nil {
			log.Printf("shutdown close error: %v", closeErr)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("backend listening on %s", app.Server.Addr)
		if serveErr := app.Server.ListenAndServe(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			log.Printf("http server stopped: %v", serveErr)
			stop()
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.Server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}
}
