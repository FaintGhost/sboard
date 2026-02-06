package api

import (
  "sboard/panel/internal/config"
  "sboard/panel/internal/db"
  "github.com/gin-gonic/gin"
)

func NewRouter(cfg config.Config, store *db.Store) *gin.Engine {
  r := gin.New()
  r.Use(RequestLogger(cfg.LogRequests))
  r.Use(gin.Recovery())
  r.Use(CORSMiddleware(cfg.CORSAllowOrigins))
  r.GET("/api/health", Health)
  r.POST("/api/admin/login", AdminLogin(cfg))
  r.GET("/api/sub/:user_uuid", SubscriptionGet(store))
  auth := r.Group("/api")
  auth.Use(AuthMiddleware(cfg.JWTSecret))
  auth.GET("/users", UsersList(store))
  auth.POST("/users", UsersCreate(store))
  auth.GET("/users/:id", UsersGet(store))
  auth.PUT("/users/:id", UsersUpdate(store))
  auth.DELETE("/users/:id", UsersDelete(store))
  return r
}
