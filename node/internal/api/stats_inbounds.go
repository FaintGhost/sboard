package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"sboard/node/internal/stats"
)

type InboundTrafficProvider interface {
	InboundTraffic(reset bool) []stats.InboundTraffic
}

func StatsInboundsGet(secret string, provider InboundTrafficProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !requireBearer(c, secret) {
			return
		}
		if provider == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "stats not ready"})
			return
		}
		reset := false
		if v := strings.TrimSpace(c.Query("reset")); v == "1" || strings.EqualFold(v, "true") {
			reset = true
		}
		items := provider.InboundTraffic(reset)
		c.JSON(http.StatusOK, gin.H{"data": items, "reset": reset})
	}
}
