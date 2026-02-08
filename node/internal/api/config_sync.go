package api

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"sboard/node/internal/sync"
)

const nodeSyncDebugPayloadEnv = "NODE_SYNC_DEBUG_PAYLOAD"

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
	if shouldDebugNodeSyncPayload() {
		log.Printf("[sync] payload=%s", syncPayloadDebugJSONFromBody(body))
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

func shouldDebugNodeSyncPayload() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(nodeSyncDebugPayloadEnv)))
	switch v {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

func syncPayloadDebugJSONFromBody(body []byte) string {
	var data any
	if err := json.Unmarshal(body, &data); err != nil {
		return `{"error":"decode payload failed"}`
	}
	sanitized := sanitizeNodeSyncPayloadForLog(data, "")
	out, err := json.Marshal(sanitized)
	if err != nil {
		return `{"error":"encode sanitized payload failed"}`
	}
	if len(out) > 65535 {
		return string(out[:65535]) + "...(truncated)"
	}
	return string(out)
}

func sanitizeNodeSyncPayloadForLog(value any, key string) any {
	key = strings.ToLower(strings.TrimSpace(key))
	switch v := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for k, val := range v {
			out[k] = sanitizeNodeSyncPayloadForLog(val, k)
		}
		return out
	case []any:
		out := make([]any, 0, len(v))
		for _, item := range v {
			out = append(out, sanitizeNodeSyncPayloadForLog(item, key))
		}
		return out
	case string:
		if key == "password" || key == "uuid" {
			return maskNodeSyncCredential(v)
		}
		return v
	default:
		return value
	}
}

func maskNodeSyncCredential(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len(value) <= 8 {
		return "***"
	}
	return value[:4] + "..." + value[len(value)-4:]
}
