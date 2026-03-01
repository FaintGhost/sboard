package config_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
		"PANEL_URL",
		"NODE_HEARTBEAT_INTERVAL",
		"NODE_UUID",
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
	require.Equal(t, "", cfg.PanelURL)
	require.Equal(t, 30, cfg.HeartbeatIntervalS)
	require.Equal(t, "", cfg.NodeUUID)
}

func TestLoadFromEnv(t *testing.T) {
	resetNodeEnv(t)
	t.Setenv("NODE_HTTP_ADDR", ":39000")
	t.Setenv("NODE_SECRET_KEY", "node-secret")
	t.Setenv("NODE_LOG_LEVEL", "debug")
	t.Setenv("NODE_STATE_PATH", "/tmp/node-state.json")
	t.Setenv("PANEL_URL", "https://panel.example.com")
	t.Setenv("NODE_HEARTBEAT_INTERVAL", "15")
	t.Setenv("NODE_UUID", "test-uuid-1234")

	cfg := config.Load()
	require.Equal(t, ":39000", cfg.HTTPAddr)
	require.Equal(t, "node-secret", cfg.SecretKey)
	require.Equal(t, "debug", cfg.LogLevel)
	require.Equal(t, "/tmp/node-state.json", cfg.StatePath)
	require.Equal(t, "https://panel.example.com", cfg.PanelURL)
	require.Equal(t, 15, cfg.HeartbeatIntervalS)
	require.Equal(t, "test-uuid-1234", cfg.NodeUUID)
}

func TestHeartbeatInterval_Default(t *testing.T) {
	resetNodeEnv(t)

	cfg := config.Load()
	assert.Equal(t, 30*time.Second, cfg.HeartbeatInterval())
}

func TestHeartbeatInterval_Custom(t *testing.T) {
	resetNodeEnv(t)
	t.Setenv("NODE_HEARTBEAT_INTERVAL", "60")

	cfg := config.Load()
	assert.Equal(t, 60*time.Second, cfg.HeartbeatInterval())
}

func TestHeartbeatInterval_BelowMinimum(t *testing.T) {
	resetNodeEnv(t)
	t.Setenv("NODE_HEARTBEAT_INTERVAL", "3")

	cfg := config.Load()
	// Below minimum of 5s, should fall back to default 30s.
	assert.Equal(t, 30*time.Second, cfg.HeartbeatInterval())
}

func TestHeartbeatInterval_InvalidString(t *testing.T) {
	resetNodeEnv(t)
	t.Setenv("NODE_HEARTBEAT_INTERVAL", "not-a-number")

	cfg := config.Load()
	// Invalid input is ignored, keeps default 30s.
	assert.Equal(t, 30*time.Second, cfg.HeartbeatInterval())
}
