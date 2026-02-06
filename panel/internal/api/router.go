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
  auth.GET("/users/:id/groups", UserGroupsGet(store))
  auth.PUT("/users/:id/groups", UserGroupsPut(store))
  auth.GET("/groups", GroupsList(store))
  auth.POST("/groups", GroupsCreate(store))
  auth.GET("/groups/:id", GroupsGet(store))
  auth.PUT("/groups/:id", GroupsUpdate(store))
  auth.DELETE("/groups/:id", GroupsDelete(store))
  auth.GET("/nodes", NodesList(store))
  auth.POST("/nodes", NodesCreate(store))
  auth.GET("/nodes/:id", NodesGet(store))
  auth.PUT("/nodes/:id", NodesUpdate(store))
  auth.DELETE("/nodes/:id", NodesDelete(store))
  auth.GET("/nodes/:id/health", NodeHealth(store))
  auth.POST("/nodes/:id/sync", NodeSync(store))
  auth.GET("/inbounds", InboundsList(store))
  auth.POST("/inbounds", InboundsCreate(store))
  auth.GET("/inbounds/:id", InboundsGet(store))
  auth.PUT("/inbounds/:id", InboundsUpdate(store))
  auth.DELETE("/inbounds/:id", InboundsDelete(store))

  if cfg.ServeWeb {
    ServeWebUI(r, cfg.WebDir)
  }
  return r
}
