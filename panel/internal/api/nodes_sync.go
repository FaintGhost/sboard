package api

import (
  "errors"
  "net/http"
  "net/http/httputil"
  "strings"

  "sboard/panel/internal/db"
  "sboard/panel/internal/node"
  "github.com/gin-gonic/gin"
)

var nodeClientFactory = func() *node.Client {
  return node.NewClient(nil)
}

// SetNodeClientFactoryForTest allows api tests to inject a fake node client without opening sockets.
func SetNodeClientFactoryForTest(f func() *node.Client) (restore func()) {
  old := nodeClientFactory
  nodeClientFactory = f
  return func() { nodeClientFactory = old }
}

func NodeHealth(store *db.Store) gin.HandlerFunc {
  return func(c *gin.Context) {
    if !ensureStore(c, store) {
      return
    }
    id, err := parseID(c.Param("id"))
    if err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
      return
    }
    n, err := store.GetNodeByID(c.Request.Context(), id)
    if err != nil {
      if errors.Is(err, db.ErrNotFound) {
        c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
        return
      }
      c.JSON(http.StatusInternalServerError, gin.H{"error": "get node failed"})
      return
    }
    client := nodeClientFactory()
    if err := client.Health(c.Request.Context(), n); err != nil {
      c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
      return
    }
    c.JSON(http.StatusOK, gin.H{"status": "ok"})
  }
}

func NodeSync(store *db.Store) gin.HandlerFunc {
  return func(c *gin.Context) {
    if !ensureStore(c, store) {
      return
    }
    id, err := parseID(c.Param("id"))
    if err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
      return
    }
    n, err := store.GetNodeByID(c.Request.Context(), id)
    if err != nil {
      if errors.Is(err, db.ErrNotFound) {
        c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
        return
      }
      c.JSON(http.StatusInternalServerError, gin.H{"error": "get node failed"})
      return
    }
    if n.GroupID == nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": "node group_id not set"})
      return
    }

    inbounds, err := store.ListInbounds(c.Request.Context(), 10000, 0, n.ID)
    if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "list inbounds failed"})
      return
    }
    users, err := store.ListActiveUsersForGroup(c.Request.Context(), *n.GroupID)
    if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "list users failed"})
      return
    }

    payload, err := node.BuildSyncPayload(n, inbounds, users)
    if err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }

    client := nodeClientFactory()
    if err := client.SyncConfig(c.Request.Context(), n, payload); err != nil {
      // include a short dump for debugging when node returns 4xx/5xx
      if strings.Contains(err.Error(), "node sync status") {
        // Best-effort: show current request context only, not secrets.
        dump, _ := httputil.DumpRequest(c.Request, false)
        _ = dump
      }
      c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
      return
    }

    c.JSON(http.StatusOK, gin.H{"status": "ok"})
  }
}

