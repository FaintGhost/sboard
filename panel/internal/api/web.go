package api

import (
  "net/http"
  "path/filepath"
  "strings"

  "github.com/gin-gonic/gin"
)

// ServeWebUI serves a Vite-built SPA from webDir.
//
// Routes:
// - /assets/*: static assets
// - /: index.html
// - other GET paths (except /api/*): fallback to index.html (SPA routing)
func ServeWebUI(r *gin.Engine, webDir string) {
  dir := strings.TrimSpace(webDir)
  if dir == "" {
    return
  }

  indexPath := filepath.Join(dir, "index.html")
  assetsDir := filepath.Join(dir, "assets")

  // Vite build output.
  r.Static("/assets", assetsDir)
  r.GET("/", func(c *gin.Context) {
    c.File(indexPath)
  })

  // SPA fallback (keep /api/* as API-only).
  r.NoRoute(func(c *gin.Context) {
    if strings.HasPrefix(c.Request.URL.Path, "/api/") || c.Request.URL.Path == "/api" {
      c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
      return
    }
    if c.Request.Method != http.MethodGet {
      c.Status(http.StatusNotFound)
      return
    }
    c.File(indexPath)
  })
}

