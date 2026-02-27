package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sboard/panel/internal/api"
	"sboard/panel/internal/config"
	"sboard/panel/internal/db"
	"sboard/panel/internal/monitor"
	"sboard/panel/internal/rpc"
)

func ensureSetupTokenIfNeeded(cfg *config.Config, store *db.Store) (string, error) {
	n, err := db.AdminCount(store)
	if err != nil {
		return "", err
	}
	if n > 0 {
		return "", nil
	}
	if cfg.SetupToken == "" {
		token, err := api.GenerateSetupToken()
		if err != nil {
			return "", err
		}
		cfg.SetupToken = token
	}
	return cfg.SetupToken, nil
}

func main() {
	cfg := config.Load()
	if err := config.Validate(cfg); err != nil {
		log.Fatal(err)
	}
	database, err := db.Open(cfg.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	if err := db.MigrateUp(database, ""); err != nil {
		log.Fatal(err)
	}
	store := db.NewStore(database)

	if err := api.InitSystemTimezone(context.Background(), store); err != nil {
		log.Fatalf("init timezone failed: %v", err)
	}

	if token, err := ensureSetupTokenIfNeeded(&cfg, store); err != nil {
		log.Fatal(err)
	} else if token != "" {
		log.Printf("[setup] no admin found. setup token: %s", token)
	}

	monitorCtx, monitorCancel := context.WithCancel(context.Background())

	if cfg.NodeMonitorInterval > 0 {
		m := monitor.NewNodesMonitor(store, nil)
		go func() {
			if err := m.CheckOnce(monitorCtx); err != nil {
				log.Printf("[monitor] check failed: %v", err)
			}
			ticker := time.NewTicker(cfg.NodeMonitorInterval)
			defer ticker.Stop()
			for {
				select {
				case <-monitorCtx.Done():
					return
				case <-ticker.C:
					if err := m.CheckOnce(monitorCtx); err != nil {
						log.Printf("[monitor] check failed: %v", err)
					}
				}
			}
		}()
		log.Printf("[monitor] nodes enabled interval=%s", cfg.NodeMonitorInterval)
	}

	tm := monitor.NewTrafficMonitor(store, nil)
	go tm.Run(monitorCtx, cfg.TrafficMonitorInterval)
	log.Printf("[monitor] traffic enabled interval=%s", cfg.TrafficMonitorInterval)

	rpcHandler := rpc.NewHandler(cfg, store)
	r := api.NewRouterWithRPC(cfg, store, rpcHandler)
	srv := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: r,
	}

	go func() {
		log.Printf("[http] starting server on %s", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[http] listen error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Printf("[shutdown] received signal: %v", sig)

	monitorCancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("[shutdown] http server shutdown error: %v", err)
	} else {
		log.Printf("[shutdown] http server stopped")
	}

	if err := database.Close(); err != nil {
		log.Printf("[shutdown] db close error: %v", err)
	} else {
		log.Printf("[shutdown] db closed")
	}

	log.Printf("[shutdown] graceful shutdown complete")
}
