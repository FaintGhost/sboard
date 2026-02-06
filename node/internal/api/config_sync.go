package api

import (
  "io"
  "net/http"
  "log"
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
  log.Printf("[sync] apply request from=%s bytes=%d", c.ClientIP(), len(body))
  if err := core.ApplyConfig(c, body); err != nil {
    // Attach error to gin context so middleware can log it too.
    _ = c.Error(err)
    log.Printf("[sync] apply failed: %v", err)
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }
  c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
