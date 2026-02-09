package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"sboard/panel/internal/singboxcli"
)

type SingBoxTools interface {
	Format(ctx context.Context, config string) (string, error)
	Check(ctx context.Context, config string) (string, error)
	Generate(ctx context.Context, kind string) (string, error)
}

var singBoxToolsFactory = func() SingBoxTools {
	return singboxcli.New()
}

func SetSingBoxToolsFactoryForTest(f func() SingBoxTools) (restore func()) {
	old := singBoxToolsFactory
	singBoxToolsFactory = f
	return func() { singBoxToolsFactory = old }
}

type singBoxConfigReq struct {
	Config string `json:"config"`
	Mode   string `json:"mode"`
}

type singBoxGenerateReq struct {
	Command string `json:"command"`
}

func SingBoxFormat(tools SingBoxTools) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req singBoxConfigReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}

		wrapped, err := wrapConfigIfNeeded(req.Config, req.Mode)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 8*time.Second)
		defer cancel()

		formatted, err := tools.Format(ctx, wrapped)
		if err != nil {
			writeSingBoxError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": gin.H{"formatted": formatted}})
	}
}

func SingBoxCheck(tools SingBoxTools) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req singBoxConfigReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}

		wrapped, err := wrapConfigIfNeeded(req.Config, req.Mode)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 8*time.Second)
		defer cancel()

		output, err := tools.Check(ctx, wrapped)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"data": gin.H{"ok": false, "output": strings.TrimSpace(err.Error())}})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": gin.H{"ok": true, "output": output}})
	}
}

func SingBoxGenerate(tools SingBoxTools) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req singBoxGenerateReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 8*time.Second)
		defer cancel()

		output, err := tools.Generate(ctx, req.Command)
		if err != nil {
			writeSingBoxError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": gin.H{"output": output}})
	}
}

func writeSingBoxError(c *gin.Context, err error) {
	if errors.Is(err, singboxcli.ErrInvalidGenerateKind) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid generate command"})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": strings.TrimSpace(err.Error())})
}

func wrapConfigIfNeeded(configText string, mode string) (string, error) {
	trimmed := strings.TrimSpace(configText)
	if trimmed == "" {
		return "", errors.New("config required")
	}

	var obj map[string]json.RawMessage
	if err := json.Unmarshal([]byte(trimmed), &obj); err != nil {
		return "", errors.New("invalid json")
	}

	normalizedMode := strings.ToLower(strings.TrimSpace(mode))
	isFullConfig := hasAnyKey(obj, "inbounds", "outbounds", "route", "dns", "log")

	if normalizedMode == "inbound" {
		if isFullConfig {
			return normalizeInboundEditorConfig(trimmed)
		}
		wrapped, err := wrapInboundAsConfig(trimmed)
		if err != nil {
			return "", err
		}
		return wrapped, nil
	}

	if normalizedMode == "config" {
		return trimmed, nil
	}

	if normalizedMode == "" || normalizedMode == "auto" {
		if isFullConfig {
			return normalizeInboundEditorConfig(trimmed)
		}
		wrapped, err := wrapInboundAsConfig(trimmed)
		if err != nil {
			return "", err
		}
		return wrapped, nil
	}

	return trimmed, nil
}

func hasAnyKey(obj map[string]json.RawMessage, keys ...string) bool {
	for _, key := range keys {
		if _, ok := obj[key]; ok {
			return true
		}
	}
	return false
}

func wrapInboundAsConfig(inboundText string) (string, error) {
	var inbound map[string]any
	if err := json.Unmarshal([]byte(inboundText), &inbound); err != nil {
		return "", errors.New("invalid inbound json")
	}

	// panel metadata field, not a sing-box inbound field.
	delete(inbound, "public_port")

	wrapped := map[string]any{
		"log": map[string]any{
			"level": "error",
		},
		"inbounds": []any{inbound},
		"outbounds": []map[string]any{{
			"type": "direct",
			"tag":  "direct",
		}},
	}

	raw, err := json.Marshal(wrapped)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func normalizeInboundEditorConfig(configText string) (string, error) {
	var config map[string]any
	if err := json.Unmarshal([]byte(configText), &config); err != nil {
		return "", errors.New("invalid config json")
	}

	rawInbounds, ok := config["inbounds"]
	if !ok {
		return configText, nil
	}

	inbounds, ok := rawInbounds.([]any)
	if !ok {
		return "", errors.New("inbounds must be an array")
	}

	for idx, item := range inbounds {
		inbound, ok := item.(map[string]any)
		if !ok {
			return "", errors.New("inbounds[" + strconv.Itoa(idx) + "] must be an object")
		}
		delete(inbound, "public_port")
	}

	out, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	return string(out), nil
}
