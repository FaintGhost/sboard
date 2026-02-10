package api

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"sboard/panel/internal/db"
	"sboard/panel/internal/node"
)

type nodeDTO struct {
	ID            int64      `json:"id"`
	UUID          string     `json:"uuid"`
	Name          string     `json:"name"`
	APIAddress    string     `json:"api_address"`
	APIPort       int        `json:"api_port"`
	SecretKey     string     `json:"secret_key"`
	PublicAddress string     `json:"public_address"`
	GroupID       *int64     `json:"group_id"`
	Status        string     `json:"status"`
	LastSeenAt    *time.Time `json:"last_seen_at"`
}

type createNodeReq struct {
	Name          string `json:"name"`
	APIAddress    string `json:"api_address"`
	APIPort       int    `json:"api_port"`
	SecretKey     string `json:"secret_key"`
	PublicAddress string `json:"public_address"`
	GroupID       *int64 `json:"group_id"`
}

type updateNodeReq struct {
	Name          *string `json:"name"`
	APIAddress    *string `json:"api_address"`
	APIPort       *int    `json:"api_port"`
	SecretKey     *string `json:"secret_key"`
	PublicAddress *string `json:"public_address"`
	GroupID       *int64  `json:"group_id"`
	GroupIDSet    bool    `json:"-"`
}

func NodesCreate(store *db.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !ensureStore(c, store) {
			return
		}
		var req createNodeReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		name := strings.TrimSpace(req.Name)
		apiAddr := strings.TrimSpace(req.APIAddress)
		pubAddr := strings.TrimSpace(req.PublicAddress)
		if name == "" || apiAddr == "" || req.APIPort <= 0 || strings.TrimSpace(req.SecretKey) == "" || pubAddr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid node"})
			return
		}
		n, err := store.CreateNode(c.Request.Context(), db.NodeCreate{
			Name:          name,
			APIAddress:    apiAddr,
			APIPort:       req.APIPort,
			SecretKey:     req.SecretKey,
			PublicAddress: pubAddr,
			GroupID:       req.GroupID,
		})
		if err != nil {
			if errors.Is(err, db.ErrConflict) {
				c.JSON(http.StatusConflict, gin.H{"error": "conflict"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "create node failed"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": toNodeDTO(n)})
	}
}

func NodesList(store *db.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !ensureStore(c, store) {
			return
		}
		limit, offset, err := parseLimitOffset(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pagination"})
			return
		}
		items, err := store.ListNodes(c.Request.Context(), limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "list nodes failed"})
			return
		}
		out := make([]nodeDTO, 0, len(items))
		for _, n := range items {
			out = append(out, toNodeDTO(n))
		}
		c.JSON(http.StatusOK, gin.H{"data": out})
	}
}

func NodesGet(store *db.Store) gin.HandlerFunc {
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
		c.JSON(http.StatusOK, gin.H{"data": toNodeDTO(n)})
	}
}

func NodesUpdate(store *db.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !ensureStore(c, store) {
			return
		}
		id, err := parseID(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		var req updateNodeReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		upd := db.NodeUpdate{}
		if req.Name != nil {
			name := strings.TrimSpace(*req.Name)
			if name == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid name"})
				return
			}
			upd.Name = &name
		}
		if req.APIAddress != nil {
			addr := strings.TrimSpace(*req.APIAddress)
			if addr == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid api_address"})
				return
			}
			upd.APIAddress = &addr
		}
		if req.APIPort != nil {
			if *req.APIPort <= 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid api_port"})
				return
			}
			upd.APIPort = req.APIPort
		}
		if req.SecretKey != nil {
			if strings.TrimSpace(*req.SecretKey) == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid secret_key"})
				return
			}
			upd.SecretKey = req.SecretKey
		}
		if req.PublicAddress != nil {
			addr := strings.TrimSpace(*req.PublicAddress)
			if addr == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid public_address"})
				return
			}
			upd.PublicAddress = &addr
		}
		if req.GroupID != nil {
			upd.GroupIDSet = true
			upd.GroupID = req.GroupID
		}
		n, err := store.UpdateNode(c.Request.Context(), id, upd)
		if err != nil {
			if errors.Is(err, db.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
				return
			}
			if errors.Is(err, db.ErrConflict) {
				c.JSON(http.StatusConflict, gin.H{"error": "conflict"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "update node failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": toNodeDTO(n)})
	}
}

func NodesDelete(store *db.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !ensureStore(c, store) {
			return
		}
		id, err := parseID(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		force := parseBoolQuery(c.Query("force"))
		if !force {
			if err := store.DeleteNode(c.Request.Context(), id); err != nil {
				if errors.Is(err, db.ErrNotFound) {
					c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
					return
				}
				if errors.Is(err, db.ErrConflict) {
					c.JSON(http.StatusConflict, gin.H{"error": "node is in use"})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": "delete node failed"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
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

		inbounds, err := store.ListInbounds(c.Request.Context(), 10000, 0, n.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "list inbounds failed"})
			return
		}

		if len(inbounds) > 0 {
			lock := nodeLock(n.ID)
			lock.Lock()
			defer lock.Unlock()

			client := nodeClientFactory()
			emptyPayload := node.SyncPayload{Inbounds: []map[string]any{}}
			if err := client.SyncConfig(c.Request.Context(), n, emptyPayload); err != nil {
				c.JSON(http.StatusBadGateway, gin.H{"error": "force drain failed: " + err.Error()})
				return
			}
			_ = store.MarkNodeOnline(c.Request.Context(), n.ID, store.NowUTC())
		}

		deletedInbounds, err := store.DeleteInboundsByNode(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "delete node inbounds failed"})
			return
		}
		if err := store.DeleteNode(c.Request.Context(), id); err != nil {
			if errors.Is(err, db.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
				return
			}
			if errors.Is(err, db.ErrConflict) {
				c.JSON(http.StatusConflict, gin.H{"error": "node is in use"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "delete node failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "force": true, "deleted_inbounds": deletedInbounds})
	}
}

func parseBoolQuery(raw string) bool {
	v := strings.TrimSpace(strings.ToLower(raw))
	switch v {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

func toNodeDTO(n db.Node) nodeDTO {
	return nodeDTO{
		ID:            n.ID,
		UUID:          n.UUID,
		Name:          n.Name,
		APIAddress:    n.APIAddress,
		APIPort:       n.APIPort,
		SecretKey:     n.SecretKey,
		PublicAddress: n.PublicAddress,
		GroupID:       n.GroupID,
		Status:        n.Status,
		LastSeenAt:    timeInSystemTimezonePtr(n.LastSeenAt),
	}
}
