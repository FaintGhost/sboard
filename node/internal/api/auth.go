package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func requireBearer(c *gin.Context, secret string) bool {
	auth := c.GetHeader("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") || strings.TrimPrefix(auth, "Bearer ") != secret {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return false
	}
	return true
}
