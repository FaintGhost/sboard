package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"sboard/node/internal/api"
	"sboard/node/internal/config"
	"sboard/node/internal/core"
	"sboard/node/internal/state"
	"sboard/node/internal/sync"
)

type coreAdapter struct {
	c         *core.Core
	sbctx     context.Context
	statePath string
}

func (a *coreAdapter) ApplyConfig(ctx *gin.Context, body []byte) error {
	options, err := sync.ParseAndValidateConfig(a.sbctx, body)
	if err != nil {
		return err
	}
	if err := a.c.ApplyOptions(options, body); err != nil {
		return err
	}
	if err := state.Persist(a.statePath, body); err != nil {
		// Don't fail the sync if persistence fails; but log it for debugging.
		log.Printf("[state] persist failed path=%s err=%v", a.statePath, err)
	}
	return nil
}

func main() {
	cfg := config.Load()
	sbctx := sync.NewSingboxContext()
	c, err := core.New(sbctx, cfg.LogLevel)
	if err != nil {
		log.Fatal(err)
	}
	adapter := &coreAdapter{c: c, sbctx: sbctx, statePath: cfg.StatePath}

	if applied, err := state.Restore(cfg.StatePath, func(raw []byte) error {
		// ApplyConfig expects a gin.Context for logging hooks; startup restore has no HTTP context.
		// We still validate and apply the same payload to restore config after restart.
		options, err := sync.ParseAndValidateConfig(sbctx, raw)
		if err != nil {
			return err
		}
		return c.ApplyOptions(options, raw)
	}); err != nil {
		log.Printf("[state] restore failed path=%s err=%v", cfg.StatePath, err)
	} else if applied {
		log.Printf("[state] restored from path=%s", cfg.StatePath)
	}

	r := api.NewRouter(cfg.SecretKey, adapter, c)

	// Create HTTP server with explicit address
	srv := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: r,
	}

	// Start server in goroutine
	go func() {
		log.Printf("[http] starting server on %s", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[http] listen error: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Printf("[shutdown] received signal: %v", sig)

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[shutdown] http server shutdown error: %v", err)
	} else {
		log.Printf("[shutdown] http server stopped")
	}

	// Close sing-box core
	if err := c.Close(); err != nil {
		log.Printf("[shutdown] core close error: %v", err)
	} else {
		log.Printf("[shutdown] core stopped")
	}

	log.Printf("[shutdown] graceful shutdown complete")
}
