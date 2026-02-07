package config

import (
	"errors"
	"os"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr         string
	DBPath           string
	JWTSecret        string
	SetupToken       string
	CORSAllowOrigins string
	LogRequests      bool

	// Optional: serve the built web UI (Vite dist) from the Panel process.
	ServeWeb bool
	WebDir   string

	// Node monitor: periodically probe node health and auto-sync when nodes come online.
	NodeMonitorInterval time.Duration
}

func Load() Config {
	cfg := Config{
		HTTPAddr:            ":8080",
		DBPath:              "panel.db",
		LogRequests:         true,
		ServeWeb:            false,
		WebDir:              "web/dist",
		NodeMonitorInterval: 10 * time.Second,
	}
	if v := os.Getenv("PANEL_HTTP_ADDR"); v != "" {
		cfg.HTTPAddr = v
	}
	if v := os.Getenv("PANEL_DB_PATH"); v != "" {
		cfg.DBPath = v
	}
	if v := os.Getenv("PANEL_JWT_SECRET"); v != "" {
		cfg.JWTSecret = v
	}
	if v := os.Getenv("PANEL_SETUP_TOKEN"); v != "" {
		cfg.SetupToken = v
	}
	if v := os.Getenv("PANEL_CORS_ALLOW_ORIGINS"); v != "" {
		cfg.CORSAllowOrigins = v
	}
	if v := os.Getenv("PANEL_LOG_REQUESTS"); v != "" {
		cfg.LogRequests = parseBool(v, cfg.LogRequests)
	}
	if v := os.Getenv("PANEL_SERVE_WEB"); v != "" {
		cfg.ServeWeb = parseBool(v, cfg.ServeWeb)
	}
	if v := os.Getenv("PANEL_WEB_DIR"); v != "" {
		cfg.WebDir = v
	}
	if v := os.Getenv("PANEL_NODE_MONITOR_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(strings.TrimSpace(v)); err == nil {
			cfg.NodeMonitorInterval = d
		}
	}
	return cfg
}

func Validate(cfg Config) error {
	if cfg.JWTSecret == "" {
		return errors.New("missing jwt config")
	}
	return nil
}

func parseBool(raw string, defaultValue bool) bool {
	v := strings.TrimSpace(strings.ToLower(raw))
	if v == "" {
		return defaultValue
	}
	switch v {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return defaultValue
	}
}
