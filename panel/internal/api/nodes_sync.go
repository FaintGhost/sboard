package api

import (
	"errors"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/gin-gonic/gin"
	"sboard/panel/internal/db"
	"sboard/panel/internal/node"
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
			_ = store.MarkNodeOffline(c.Request.Context(), n.ID)
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		_ = store.MarkNodeOnline(c.Request.Context(), n.ID, store.Now().UTC())
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
		res := trySyncNode(c.Request.Context(), store, n)
		if res.Status != "ok" {
			// include a short dump for debugging when node returns 4xx/5xx
			if strings.Contains(res.Error, "node sync status") {
				// Best-effort: show current request context only, not secrets.
				dump, _ := httputil.DumpRequest(c.Request, false)
				_ = dump
			}
			c.JSON(http.StatusBadGateway, gin.H{"error": res.Error})
			return
		}
		_ = store.MarkNodeOnline(c.Request.Context(), n.ID, store.Now().UTC())
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}
