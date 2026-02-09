package api

import (
	"errors"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"sboard/panel/internal/db"
)

const subscriptionBaseURLKey = "subscription_base_url"

type systemSettingsDTO struct {
	SubscriptionBaseURL string `json:"subscription_base_url"`
}

type updateSystemSettingsReq struct {
	SubscriptionBaseURL string `json:"subscription_base_url"`
}

func SystemSettingsGet(store *db.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !ensureStore(c, store) {
			return
		}

		value, err := store.GetSystemSetting(c.Request.Context(), subscriptionBaseURLKey)
		if err != nil && !errors.Is(err, db.ErrNotFound) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "load settings failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": systemSettingsDTO{SubscriptionBaseURL: value},
		})
	}
}

func SystemSettingsPut(store *db.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !ensureStore(c, store) {
			return
		}

		var req updateSystemSettingsReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}

		normalized, err := normalizeSubscriptionBaseURL(req.SubscriptionBaseURL)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if normalized == "" {
			if err := store.DeleteSystemSetting(c.Request.Context(), subscriptionBaseURLKey); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "save settings failed"})
				return
			}
		} else {
			if err := store.UpsertSystemSetting(c.Request.Context(), subscriptionBaseURLKey, normalized); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "save settings failed"})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"data": systemSettingsDTO{SubscriptionBaseURL: normalized},
		})
	}
}

func normalizeSubscriptionBaseURL(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", nil
	}

	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", errors.New("invalid subscription_base_url")
	}

	scheme := strings.ToLower(strings.TrimSpace(parsed.Scheme))
	if scheme != "http" && scheme != "https" {
		return "", errors.New("subscription_base_url must use http or https")
	}

	if parsed.User != nil || (parsed.Path != "" && parsed.Path != "/") || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", errors.New("subscription_base_url must be protocol + ip:port")
	}

	host := strings.TrimSpace(parsed.Hostname())
	if net.ParseIP(host) == nil {
		return "", errors.New("subscription_base_url must use a valid IP")
	}

	portStr := strings.TrimSpace(parsed.Port())
	if portStr == "" {
		return "", errors.New("subscription_base_url must include port")
	}

	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		return "", errors.New("subscription_base_url has invalid port")
	}

	return scheme + "://" + net.JoinHostPort(host, strconv.Itoa(port)), nil
}
