package api

import "github.com/gin-gonic/gin"

func NewRouter() *gin.Engine {
  r := gin.New()
  r.GET("/api/health", Health)
  return r
}
