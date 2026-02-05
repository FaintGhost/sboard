package api

import (
  "sboard/panel/internal/config"
  "sboard/panel/internal/db"
  "github.com/gin-gonic/gin"
)

func NewRouter(cfg config.Config, store *db.Store) *gin.Engine {
  r := gin.New()
  r.GET("/api/health", Health)
  r.POST("/api/admin/login", AdminLogin(cfg))
  _ = store
  return r
}
