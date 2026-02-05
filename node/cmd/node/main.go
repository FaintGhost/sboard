package main

import (
  "log"

  "sboard/node/internal/api"
  "sboard/node/internal/config"
  "sboard/node/internal/core"
  "sboard/node/internal/sync"
  "github.com/gin-gonic/gin"
)

type coreAdapter struct{ c *core.Core }

func (a *coreAdapter) ApplyConfig(ctx *gin.Context, body []byte) error {
  sbctx := sync.NewSingboxContext()
  inbounds, err := sync.ParseAndValidateInbounds(sbctx, body)
  if err != nil {
    return err
  }
  return a.c.Apply(inbounds, body)
}

func main() {
  cfg := config.Load()
  sbctx := sync.NewSingboxContext()
  c, err := core.New(sbctx, cfg.LogLevel)
  if err != nil {
    log.Fatal(err)
  }
  r := api.NewRouter(cfg.SecretKey, &coreAdapter{c: c})
  if err := r.Run(cfg.HTTPAddr); err != nil {
    log.Fatal(err)
  }
}
