package config

import "os"

type Config struct {
	HTTPAddr  string
	SecretKey string
	LogLevel  string
	StatePath string
}

func Load() Config {
	cfg := Config{
		HTTPAddr:  ":3000",
		SecretKey: "",
		LogLevel:  "info",
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
	return cfg
}
