package api

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"sboard/node/internal/sync"
)

func ConfigSync(c *gin.Context, secret string, core Core) {
	if !requireBearer(c, secret) {
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
	// Best-effort debug log: show whether password/method is present without leaking full secrets.
	var meta struct {
		Inbounds []struct {
			Tag      string `json:"tag"`
			Type     string `json:"type"`
			Method   string `json:"method"`
			Password string `json:"password"`
			Users    []any  `json:"users"`
		} `json:"inbounds"`
	}
	if err := json.Unmarshal(body, &meta); err == nil {
		for _, inb := range meta.Inbounds {
			log.Printf(
				"[sync] inbound tag=%s type=%s method=%s password_len=%d users=%d",
				inb.Tag,
				inb.Type,
				inb.Method,
				len(inb.Password),
				len(inb.Users),
			)
		}
	}
	if err := core.ApplyConfig(c, body); err != nil {
		// Attach error to gin context so middleware can log it too.
		_ = c.Error(err)
		log.Printf("[sync] apply failed: %v", err)
		var bre sync.BadRequestError
		if errors.As(err, &bre) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
