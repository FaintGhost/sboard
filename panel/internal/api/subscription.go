package api

import (
  "net/http"
  "strings"

  "sboard/panel/internal/db"
  "sboard/panel/internal/subscription"
  "github.com/gin-gonic/gin"
)

func SubscriptionGet(store *db.Store) gin.HandlerFunc {
  return func(c *gin.Context) {
    if !ensureStore(c, store) {
      return
    }
    userUUID := strings.TrimSpace(c.Param("user_uuid"))
    if userUUID == "" {
      c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
      return
    }
    user, err := store.GetUserByUUID(c.Request.Context(), userUUID)
    if err != nil {
      if err == db.ErrNotFound {
        c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
        return
      }
      c.JSON(http.StatusInternalServerError, gin.H{"error": "get user failed"})
      return
    }
    if user.Status != "active" {
      c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
      return
    }

    inbounds, err := store.ListUserInbounds(c.Request.Context(), user.ID)
    if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "list inbounds failed"})
      return
    }
    items := make([]subscription.Item, 0, len(inbounds))
    for _, inb := range inbounds {
      items = append(items, subscription.Item{
        InboundUUID:       inb.InboundUUID,
        InboundType:       inb.InboundType,
        InboundTag:        inb.InboundTag,
        NodePublicAddress: inb.NodePublicAddress,
        InboundListenPort: inb.InboundListenPort,
        InboundPublicPort: inb.InboundPublicPort,
        Settings:          inb.Settings,
        TLSSettings:       inb.TLSSettings,
        TransportSettings: inb.TransportSettings,
      })
    }

    if format := strings.TrimSpace(c.Query("format")); format != "" {
      if format != "singbox" && format != "v2ray" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid format"})
        return
      }
      if format == "singbox" {
        payload, err := subscription.BuildSingbox(subscription.User{
          UUID:     user.UUID,
          Username: user.Username,
        }, items)
        if err != nil {
          c.JSON(http.StatusInternalServerError, gin.H{"error": "build subscription failed"})
          return
        }
        c.Data(http.StatusOK, "application/json; charset=utf-8", payload)
        return
      }
      payload, err := subscription.BuildV2Ray(subscription.User{
        UUID:     user.UUID,
        Username: user.Username,
      }, items)
      if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "build subscription failed"})
        return
      }
      c.Data(http.StatusOK, "text/plain; charset=utf-8", payload)
      return
    }

    if isSingboxUA(c.GetHeader("User-Agent")) {
      payload, err := subscription.BuildSingbox(subscription.User{
        UUID:     user.UUID,
        Username: user.Username,
      }, items)
      if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "build subscription failed"})
        return
      }
      c.Data(http.StatusOK, "application/json; charset=utf-8", payload)
      return
    }

    payload, err := subscription.BuildV2Ray(subscription.User{
      UUID:     user.UUID,
      Username: user.Username,
    }, items)
    if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "build subscription failed"})
      return
    }
    c.Data(http.StatusOK, "text/plain; charset=utf-8", payload)
  }
}

func isSingboxUA(ua string) bool {
  ua = strings.ToLower(ua)
  return strings.Contains(ua, "sing-box") ||
    strings.Contains(ua, "sfa") ||
    strings.Contains(ua, "sfi")
}
