package main

import (
	"context"
	"log"

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
	if err := r.Run(cfg.HTTPAddr); err != nil {
		log.Fatal(err)
	}
}
