package config

import (
  "errors"
  "os"
)

type Config struct {
  HTTPAddr  string
  DBPath    string
  AdminUser string
  AdminPass string
  JWTSecret string
}

func Load() Config {
  cfg := Config{
    HTTPAddr: ":8080",
    DBPath:   "panel.db",
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
  return cfg
}

func Validate(cfg Config) error {
  if cfg.AdminUser == "" || cfg.AdminPass == "" || cfg.JWTSecret == "" {
    return errors.New("missing admin or jwt config")
  }
  return nil
}
