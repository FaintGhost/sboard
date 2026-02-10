package config_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"sboard/node/internal/config"
)

func resetNodeEnv(t *testing.T) {
	t.Helper()
	keys := []string{
		"NODE_HTTP_ADDR",
		"NODE_SECRET_KEY",
		"NODE_LOG_LEVEL",
		"NODE_STATE_PATH",
	}
	for _, key := range keys {
		t.Setenv(key, "")
	}
}

func TestLoadDefaults(t *testing.T) {
	resetNodeEnv(t)

	cfg := config.Load()
	require.Equal(t, ":3000", cfg.HTTPAddr)
	require.Equal(t, "", cfg.SecretKey)
	require.Equal(t, "info", cfg.LogLevel)
	require.Equal(t, "/data/last_sync.json", cfg.StatePath)
}

func TestLoadFromEnv(t *testing.T) {
	resetNodeEnv(t)
	t.Setenv("NODE_HTTP_ADDR", ":39000")
	t.Setenv("NODE_SECRET_KEY", "node-secret")
	t.Setenv("NODE_LOG_LEVEL", "debug")
	t.Setenv("NODE_STATE_PATH", "/tmp/node-state.json")

	cfg := config.Load()
	require.Equal(t, ":39000", cfg.HTTPAddr)
	require.Equal(t, "node-secret", cfg.SecretKey)
	require.Equal(t, "debug", cfg.LogLevel)
	require.Equal(t, "/tmp/node-state.json", cfg.StatePath)
}
