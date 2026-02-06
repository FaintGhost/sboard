package api

import (
  "log"
  "time"

  "github.com/gin-gonic/gin"
)

func RequestLogger(enabled bool) gin.HandlerFunc {
  if !enabled {
    return func(c *gin.Context) { c.Next() }
  }

  return func(c *gin.Context) {
    start := time.Now()
    method := c.Request.Method
    path := c.Request.URL.Path

    c.Next()

    status := c.Writer.Status()
    cost := time.Since(start)
    log.Printf("[http] %s %s -> %d (%s)", method, path, status, cost.Truncate(time.Microsecond))
  }
}

