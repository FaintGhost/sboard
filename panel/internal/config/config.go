package config

import "os"

type Config struct {
  HTTPAddr string
  DBPath   string
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
  return cfg
}
