package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"sboard/panel/internal/db"
)

type nodeTrafficSampleDTO struct {
	ID         int64   `json:"id"`
	InboundTag *string `json:"inbound_tag"`
	Upload     int64   `json:"upload"`
	Download   int64   `json:"download"`
	RecordedAt string  `json:"recorded_at"`
}

func NodeTrafficList(store *db.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil || id <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		limit := 50
		if v := c.Query("limit"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n >= 1 && n <= 500 {
				limit = n
			}
		}
		offset := 0
		if v := c.Query("offset"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n >= 0 && n <= 1_000_000 {
				offset = n
			}
		}

		items, err := store.ListNodeTrafficSamples(c.Request.Context(), id, limit, offset)
		if err != nil {
			if err == db.ErrNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		out := make([]nodeTrafficSampleDTO, 0, len(items))
		for _, it := range items {
			out = append(out, nodeTrafficSampleDTO{
				ID:         it.ID,
				InboundTag: it.InboundTag,
				Upload:     it.Upload,
				Download:   it.Download,
				RecordedAt: timeRFC3339OrEmpty(it.RecordedAt),
			})
		}
		c.JSON(http.StatusOK, gin.H{"data": out})
	}
}

func timeRFC3339OrEmpty(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}
