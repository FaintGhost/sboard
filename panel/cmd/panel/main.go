package main

import (
	"context"
	"log"
	"time"

	"sboard/panel/internal/api"
	"sboard/panel/internal/config"
	"sboard/panel/internal/db"
	"sboard/panel/internal/monitor"
)

func main() {
	cfg := config.Load()
	if err := config.Validate(cfg); err != nil {
		log.Fatal(err)
	}
	database, err := db.Open(cfg.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	if err := db.MigrateUp(database, "internal/db/migrations"); err != nil {
		log.Fatal(err)
	}
	store := db.NewStore(database)

	// If the panel is not initialized yet, require a setup token for secure onboarding.
	// If not provided via env, generate one and print it once at startup.
	if n, err := db.AdminCount(store); err == nil && n == 0 {
		if cfg.SetupToken == "" {
			token, err := api.GenerateSetupToken()
			if err != nil {
				log.Fatal(err)
			}
			cfg.SetupToken = token
		}
		log.Printf("[setup] no admin found. setup token: %s", cfg.SetupToken)
	}

	if cfg.NodeMonitorInterval > 0 {
		m := monitor.NewNodesMonitor(store, nil)
		ctx := context.Background()
		// initial pass for fast feedback in UI
		go func() {
			if err := m.CheckOnce(ctx); err != nil {
				log.Printf("[monitor] check failed: %v", err)
			}
			ticker := time.NewTicker(cfg.NodeMonitorInterval)
			defer ticker.Stop()
			for range ticker.C {
				if err := m.CheckOnce(ctx); err != nil {
					log.Printf("[monitor] check failed: %v", err)
				}
			}
		}()
		log.Printf("[monitor] nodes enabled interval=%s", cfg.NodeMonitorInterval)
	}

	r := api.NewRouter(cfg, store)
	if err := r.Run(cfg.HTTPAddr); err != nil {
		log.Fatal(err)
	}
}
