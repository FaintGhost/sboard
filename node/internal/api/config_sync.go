package api

import (
  "io"
  "net/http"
  "strings"

  "github.com/gin-gonic/gin"
)

func ConfigSync(c *gin.Context, secret string, core Core) {
  auth := c.GetHeader("Authorization")
  if !strings.HasPrefix(auth, "Bearer ") || strings.TrimPrefix(auth, "Bearer ") != secret {
    c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
    return
  }
  body, err := io.ReadAll(c.Request.Body)
  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
    return
  }
  if core == nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "core not ready"})
    return
  }
  if err := core.ApplyConfig(c, body); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }
  c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
