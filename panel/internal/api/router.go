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
  auth := r.Group("/api")
  auth.Use(AuthMiddleware(cfg.JWTSecret))
  auth.GET("/users", func(c *gin.Context) {
    c.JSON(200, gin.H{"data": []any{}})
  })
  _ = store
  return r
}
