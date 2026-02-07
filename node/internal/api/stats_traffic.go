package api

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"sboard/node/internal/stats"
)

func StatsTrafficGet(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !requireBearer(c, secret) {
			return
		}
		iface := strings.TrimSpace(c.Query("interface"))
		if iface == "" {
			iface = strings.TrimSpace(os.Getenv("NODE_TRAFFIC_INTERFACE"))
		}
		sample, err := stats.CurrentSample(iface)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, sample)
	}
}
