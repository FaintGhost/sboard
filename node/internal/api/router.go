package api

import "github.com/gin-gonic/gin"

type Core interface {
	ApplyConfig(ctx *gin.Context, body []byte) error
}

func NewRouter(secret string, core Core) *gin.Engine {
	r := gin.New()
	r.Use(RequestLogger())
	r.Use(gin.Recovery())
	r.GET("/api/health", Health)
	r.GET("/api/stats/traffic", StatsTrafficGet(secret))
	r.POST("/api/config/sync", func(c *gin.Context) {
		ConfigSync(c, secret, core)
	})
	return r
}
