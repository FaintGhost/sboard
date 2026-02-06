package api

import (
  "net/http"
  "strings"

  "github.com/gin-gonic/gin"
)

const defaultCORSOrigins = "http://localhost:5173"

func CORSMiddleware(allowOrigins string) gin.HandlerFunc {
  allowedOrigins := parseAllowedOrigins(allowOrigins)
  allowAll := len(allowedOrigins) == 1 && allowedOrigins[0] == "*"

  return func(c *gin.Context) {
    origin := c.GetHeader("Origin")
    if origin == "" {
      c.Next()
      return
    }

    originAllowed := allowAll || containsOrigin(allowedOrigins, origin)
    if originAllowed {
      if allowAll {
        c.Header("Access-Control-Allow-Origin", "*")
      } else {
        c.Header("Access-Control-Allow-Origin", origin)
      }
      c.Header("Vary", "Origin")
      c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
      c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type")
      c.Header("Access-Control-Max-Age", "86400")
    }

    if c.Request.Method == http.MethodOptions && c.GetHeader("Access-Control-Request-Method") != "" {
      if !originAllowed {
        c.AbortWithStatus(http.StatusForbidden)
        return
      }
      c.AbortWithStatus(http.StatusNoContent)
      return
    }

    c.Next()
  }
}

func parseAllowedOrigins(raw string) []string {
  value := strings.TrimSpace(raw)
  if value == "" {
    value = defaultCORSOrigins
  }
  parts := strings.Split(value, ",")
  out := make([]string, 0, len(parts))
  for _, part := range parts {
    p := strings.TrimSpace(part)
    if p == "" {
      continue
    }
    out = append(out, p)
  }
  if len(out) == 0 {
    return []string{defaultCORSOrigins}
  }
  return out
}

func containsOrigin(allowed []string, origin string) bool {
  for _, item := range allowed {
    if item == origin {
      return true
    }
  }
  return false
}
