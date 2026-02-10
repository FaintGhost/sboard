package config_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/config"
)

func resetPanelEnv(t *testing.T) {
	t.Helper()
	keys := []string{
		"PANEL_HTTP_ADDR",
		"PANEL_DB_PATH",
		"PANEL_JWT_SECRET",
		"PANEL_SETUP_TOKEN",
		"PANEL_CORS_ALLOW_ORIGINS",
		"PANEL_LOG_REQUESTS",
		"PANEL_SERVE_WEB",
		"PANEL_WEB_DIR",
		"PANEL_NODE_MONITOR_INTERVAL",
		"PANEL_TRAFFIC_MONITOR_INTERVAL",
	}
	for _, key := range keys {
		t.Setenv(key, "")
	}
}

func TestValidateConfig(t *testing.T) {
	cfg := config.Config{}
	err := config.Validate(cfg)
	require.Error(t, err)

	cfg = config.Config{JWTSecret: "secret"}
	require.NoError(t, config.Validate(cfg))
}

func TestLoadDefaults(t *testing.T) {
	resetPanelEnv(t)

	cfg := config.Load()
	require.Equal(t, ":8080", cfg.HTTPAddr)
	require.Equal(t, "panel.db", cfg.DBPath)
	require.Equal(t, "", cfg.JWTSecret)
	require.Equal(t, "", cfg.SetupToken)
	require.Equal(t, "", cfg.CORSAllowOrigins)
	require.True(t, cfg.LogRequests)
	require.False(t, cfg.ServeWeb)
	require.Equal(t, "web/dist", cfg.WebDir)
	require.Equal(t, 10*time.Second, cfg.NodeMonitorInterval)
	require.Equal(t, 30*time.Second, cfg.TrafficMonitorInterval)
}

func TestLoadFromEnv(t *testing.T) {
	resetPanelEnv(t)
	t.Setenv("PANEL_HTTP_ADDR", ":18080")
	t.Setenv("PANEL_DB_PATH", "/tmp/panel.db")
	t.Setenv("PANEL_JWT_SECRET", "jwt-secret")
	t.Setenv("PANEL_SETUP_TOKEN", "setup-token")
	t.Setenv("PANEL_CORS_ALLOW_ORIGINS", "https://a.example,https://b.example")
	t.Setenv("PANEL_LOG_REQUESTS", "off")
	t.Setenv("PANEL_SERVE_WEB", "yes")
	t.Setenv("PANEL_WEB_DIR", "/app/web/dist")
	t.Setenv("PANEL_NODE_MONITOR_INTERVAL", "25s")
	t.Setenv("PANEL_TRAFFIC_MONITOR_INTERVAL", "45s")

	cfg := config.Load()
	require.Equal(t, ":18080", cfg.HTTPAddr)
	require.Equal(t, "/tmp/panel.db", cfg.DBPath)
	require.Equal(t, "jwt-secret", cfg.JWTSecret)
	require.Equal(t, "setup-token", cfg.SetupToken)
	require.Equal(t, "https://a.example,https://b.example", cfg.CORSAllowOrigins)
	require.False(t, cfg.LogRequests)
	require.True(t, cfg.ServeWeb)
	require.Equal(t, "/app/web/dist", cfg.WebDir)
	require.Equal(t, 25*time.Second, cfg.NodeMonitorInterval)
	require.Equal(t, 45*time.Second, cfg.TrafficMonitorInterval)
}

func TestLoadInvalidValuesFallbackAndClamp(t *testing.T) {
	resetPanelEnv(t)
	t.Setenv("PANEL_LOG_REQUESTS", "not-bool")
	t.Setenv("PANEL_SERVE_WEB", "not-bool")
	t.Setenv("PANEL_NODE_MONITOR_INTERVAL", "bad-duration")
	t.Setenv("PANEL_TRAFFIC_MONITOR_INTERVAL", "1s")

	cfg := config.Load()

	// invalid bool keeps default
	require.True(t, cfg.LogRequests)
	require.False(t, cfg.ServeWeb)
	// invalid duration keeps default
	require.Equal(t, 10*time.Second, cfg.NodeMonitorInterval)
	// traffic interval clamps to 5s minimum
	require.Equal(t, 5*time.Second, cfg.TrafficMonitorInterval)
}
