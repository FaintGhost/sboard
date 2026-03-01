package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Core interface {
	ApplyConfig(ctx *gin.Context, body []byte) error
}

func NewRouter(secret string, core Core, traffic InboundTrafficProvider) *gin.Engine {
	return NewRouterWithRPC(secret, core, traffic, nil)
}

func NewRouterWithRPC(secret string, core Core, traffic InboundTrafficProvider, rpcHandler http.Handler) *gin.Engine {
	r := gin.New()
	r.Use(RequestLogger())
	r.Use(gin.Recovery())

	if rpcHandler == nil {
		r.GET("/api/health", Health)
		r.GET("/api/stats/traffic", StatsTrafficGet(secret))
		r.GET("/api/stats/inbounds", StatsInboundsGet(secret, traffic))
		r.POST("/api/config/sync", func(c *gin.Context) {
			ConfigSync(c, secret, core)
		})
	}
	if rpcHandler != nil {
		r.Any("/rpc", gin.WrapH(http.StripPrefix("/rpc", rpcHandler)))
		r.Any("/rpc/*path", gin.WrapH(http.StripPrefix("/rpc", rpcHandler)))
	}
	return r
}
