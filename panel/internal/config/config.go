package config

import (
  "errors"
  "os"
  "strings"
)

type Config struct {
  HTTPAddr  string
  DBPath    string
  AdminUser string
  AdminPass string
  JWTSecret string
  CORSAllowOrigins string
  LogRequests bool
}

func Load() Config {
  cfg := Config{
    HTTPAddr: ":8080",
    DBPath:   "panel.db",
    LogRequests: true,
  }
  if v := os.Getenv("PANEL_HTTP_ADDR"); v != "" {
    cfg.HTTPAddr = v
  }
  if v := os.Getenv("PANEL_DB_PATH"); v != "" {
    cfg.DBPath = v
  }
  if v := os.Getenv("ADMIN_USER"); v != "" {
    cfg.AdminUser = v
  }
  if v := os.Getenv("ADMIN_PASS"); v != "" {
    cfg.AdminPass = v
  }
  if v := os.Getenv("PANEL_JWT_SECRET"); v != "" {
    cfg.JWTSecret = v
  }
  if v := os.Getenv("PANEL_CORS_ALLOW_ORIGINS"); v != "" {
    cfg.CORSAllowOrigins = v
  }
  if v := os.Getenv("PANEL_LOG_REQUESTS"); v != "" {
    cfg.LogRequests = parseBool(v, cfg.LogRequests)
  }
  return cfg
}

func Validate(cfg Config) error {
  if cfg.AdminUser == "" || cfg.AdminPass == "" || cfg.JWTSecret == "" {
    return errors.New("missing admin or jwt config")
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
