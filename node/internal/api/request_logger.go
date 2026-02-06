package api

import (
  "log"
  "time"

  "github.com/gin-gonic/gin"
)

func RequestLogger() gin.HandlerFunc {
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

