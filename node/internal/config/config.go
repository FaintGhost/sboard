package config

import (
	"os"
	"strconv"
	"time"
)

const (
	defaultHeartbeatIntervalS = 30
	minHeartbeatIntervalS     = 5
	defaultUUIDPath           = "/data/node-uuid"
)

type Config struct {
	HTTPAddr           string
	SecretKey          string
	LogLevel           string
	StatePath          string
	PanelURL           string
	HeartbeatIntervalS int
	NodeUUID           string
}

// HeartbeatInterval returns HeartbeatIntervalS as a time.Duration.
// It defaults to 30s and enforces a minimum of 5s.
func (c Config) HeartbeatInterval() time.Duration {
	s := c.HeartbeatIntervalS
	if s < minHeartbeatIntervalS {
		s = defaultHeartbeatIntervalS
	}
	return time.Duration(s) * time.Second
}

func Load() Config {
	cfg := Config{
		HTTPAddr:           ":3000",
		SecretKey:          "",
		LogLevel:           "info",
		HeartbeatIntervalS: defaultHeartbeatIntervalS,
		// Persist the last successful sync payload so the node can restore config after restart.
		// Mount /data for durable storage in Docker.
		StatePath: "/data/last_sync.json",
	}
	if v := os.Getenv("NODE_HTTP_ADDR"); v != "" {
		cfg.HTTPAddr = v
	}
	if v := os.Getenv("NODE_SECRET_KEY"); v != "" {
		cfg.SecretKey = v
	}
	if v := os.Getenv("NODE_LOG_LEVEL"); v != "" {
		cfg.LogLevel = v
	}
	if v := os.Getenv("NODE_STATE_PATH"); v != "" {
		cfg.StatePath = v
	}
	if v := os.Getenv("PANEL_URL"); v != "" {
		cfg.PanelURL = v
	}
	if v := os.Getenv("NODE_HEARTBEAT_INTERVAL"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.HeartbeatIntervalS = n
		}
	}
	if v := os.Getenv("NODE_UUID"); v != "" {
		cfg.NodeUUID = v
	}

	// Auto-generate UUID when PANEL_URL is set but no explicit UUID provided.
	if cfg.NodeUUID == "" && cfg.PanelURL != "" {
		id, err := LoadOrGenerateUUID(defaultUUIDPath)
		if err == nil {
			cfg.NodeUUID = id
		}
	}

	return cfg
}
