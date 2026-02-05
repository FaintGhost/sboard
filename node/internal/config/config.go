package config

import "os"

type Config struct {
  HTTPAddr  string
  SecretKey string
  LogLevel  string
}

func Load() Config {
  cfg := Config{
    HTTPAddr:  ":3000",
    SecretKey: "",
    LogLevel:  "info",
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
  return cfg
}
