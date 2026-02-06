package api

import (
  "encoding/json"
  "errors"
  "net/http"
  "strconv"
  "strings"

  "sboard/panel/internal/db"
  "github.com/gin-gonic/gin"
)

type inboundDTO struct {
  ID                int64           `json:"id"`
  UUID              string          `json:"uuid"`
  Tag               string          `json:"tag"`
  NodeID            int64           `json:"node_id"`
  Protocol          string          `json:"protocol"`
  ListenPort        int             `json:"listen_port"`
  PublicPort        int             `json:"public_port"`
  Settings          json.RawMessage `json:"settings"`
  TLSSettings       json.RawMessage `json:"tls_settings"`
  TransportSettings json.RawMessage `json:"transport_settings"`
}

type createInboundReq struct {
  Tag               string          `json:"tag"`
  NodeID            int64           `json:"node_id"`
  Protocol          string          `json:"protocol"`
  ListenPort        int             `json:"listen_port"`
  PublicPort        int             `json:"public_port"`
  Settings          json.RawMessage `json:"settings"`
  TLSSettings       json.RawMessage `json:"tls_settings"`
  TransportSettings json.RawMessage `json:"transport_settings"`
}

type updateInboundReq struct {
  Tag               *string          `json:"tag"`
  Protocol          *string          `json:"protocol"`
  ListenPort        *int             `json:"listen_port"`
  PublicPort        *int             `json:"public_port"`
  Settings          *json.RawMessage `json:"settings"`
  TLSSettings       *json.RawMessage `json:"tls_settings"`
  TransportSettings *json.RawMessage `json:"transport_settings"`
}

func InboundsCreate(store *db.Store) gin.HandlerFunc {
  return func(c *gin.Context) {
    if !ensureStore(c, store) {
      return
    }
    var req createInboundReq
    if err := c.ShouldBindJSON(&req); err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
      return
    }
    tag := strings.TrimSpace(req.Tag)
    proto := strings.TrimSpace(req.Protocol)
    if tag == "" || proto == "" || req.NodeID <= 0 || req.ListenPort <= 0 {
      c.JSON(http.StatusBadRequest, gin.H{"error": "invalid inbound"})
      return
    }
    if len(req.Settings) == 0 || !json.Valid(req.Settings) {
      c.JSON(http.StatusBadRequest, gin.H{"error": "invalid settings"})
      return
    }
    if len(req.TLSSettings) > 0 && !json.Valid(req.TLSSettings) {
      c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tls_settings"})
      return
    }
    if len(req.TransportSettings) > 0 && !json.Valid(req.TransportSettings) {
      c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transport_settings"})
      return
    }

    inb, err := store.CreateInbound(c.Request.Context(), db.InboundCreate{
      Tag:               tag,
      NodeID:            req.NodeID,
      Protocol:          proto,
      ListenPort:        req.ListenPort,
      PublicPort:        req.PublicPort,
      Settings:          req.Settings,
      TLSSettings:       req.TLSSettings,
      TransportSettings: req.TransportSettings,
    })
    if err != nil {
      if errors.Is(err, db.ErrConflict) {
        c.JSON(http.StatusConflict, gin.H{"error": "conflict"})
        return
      }
      if errors.Is(err, db.ErrNotFound) {
        c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
        return
      }
      c.JSON(http.StatusInternalServerError, gin.H{"error": "create inbound failed"})
      return
    }
    c.JSON(http.StatusCreated, gin.H{"data": toInboundDTO(inb)})
  }
}

func InboundsList(store *db.Store) gin.HandlerFunc {
  return func(c *gin.Context) {
    if !ensureStore(c, store) {
      return
    }
    limit, offset, err := parseLimitOffset(c)
    if err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pagination"})
      return
    }
    nodeID := int64(0)
    if raw := strings.TrimSpace(c.Query("node_id")); raw != "" {
      v, err := strconv.ParseInt(raw, 10, 64)
      if err != nil || v < 0 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid node_id"})
        return
      }
      nodeID = v
    }
    items, err := store.ListInbounds(c.Request.Context(), limit, offset, nodeID)
    if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "list inbounds failed"})
      return
    }
    out := make([]inboundDTO, 0, len(items))
    for _, inb := range items {
      out = append(out, toInboundDTO(inb))
    }
    c.JSON(http.StatusOK, gin.H{"data": out})
  }
}

func InboundsGet(store *db.Store) gin.HandlerFunc {
  return func(c *gin.Context) {
    if !ensureStore(c, store) {
      return
    }
    id, err := parseID(c.Param("id"))
    if err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
      return
    }
    inb, err := store.GetInboundByID(c.Request.Context(), id)
    if err != nil {
      if errors.Is(err, db.ErrNotFound) {
        c.JSON(http.StatusNotFound, gin.H{"error": "inbound not found"})
        return
      }
      c.JSON(http.StatusInternalServerError, gin.H{"error": "get inbound failed"})
      return
    }
    c.JSON(http.StatusOK, gin.H{"data": toInboundDTO(inb)})
  }
}

func InboundsUpdate(store *db.Store) gin.HandlerFunc {
  return func(c *gin.Context) {
    if !ensureStore(c, store) {
      return
    }
    id, err := parseID(c.Param("id"))
    if err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
      return
    }
    var req updateInboundReq
    if err := c.ShouldBindJSON(&req); err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
      return
    }
    upd := db.InboundUpdate{}
    if req.Tag != nil {
      tag := strings.TrimSpace(*req.Tag)
      if tag == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag"})
        return
      }
      upd.Tag = &tag
    }
    if req.Protocol != nil {
      p := strings.TrimSpace(*req.Protocol)
      if p == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid protocol"})
        return
      }
      upd.Protocol = &p
    }
    if req.ListenPort != nil {
      if *req.ListenPort <= 0 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid listen_port"})
        return
      }
      upd.ListenPort = req.ListenPort
    }
    if req.PublicPort != nil {
      if *req.PublicPort < 0 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid public_port"})
        return
      }
      upd.PublicPort = req.PublicPort
    }
    if req.Settings != nil {
      if len(*req.Settings) == 0 || !json.Valid(*req.Settings) {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid settings"})
        return
      }
      upd.Settings = req.Settings
    }
    if req.TLSSettings != nil {
      if len(*req.TLSSettings) > 0 && !json.Valid(*req.TLSSettings) {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tls_settings"})
        return
      }
      upd.TLSSettings = req.TLSSettings
    }
    if req.TransportSettings != nil {
      if len(*req.TransportSettings) > 0 && !json.Valid(*req.TransportSettings) {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transport_settings"})
        return
      }
      upd.TransportSettings = req.TransportSettings
    }
    inb, err := store.UpdateInbound(c.Request.Context(), id, upd)
    if err != nil {
      if errors.Is(err, db.ErrNotFound) {
        c.JSON(http.StatusNotFound, gin.H{"error": "inbound not found"})
        return
      }
      if errors.Is(err, db.ErrConflict) {
        c.JSON(http.StatusConflict, gin.H{"error": "conflict"})
        return
      }
      c.JSON(http.StatusInternalServerError, gin.H{"error": "update inbound failed"})
      return
    }
    c.JSON(http.StatusOK, gin.H{"data": toInboundDTO(inb)})
  }
}

func InboundsDelete(store *db.Store) gin.HandlerFunc {
  return func(c *gin.Context) {
    if !ensureStore(c, store) {
      return
    }
    id, err := parseID(c.Param("id"))
    if err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
      return
    }
    if err := store.DeleteInbound(c.Request.Context(), id); err != nil {
      if errors.Is(err, db.ErrNotFound) {
        c.JSON(http.StatusNotFound, gin.H{"error": "inbound not found"})
        return
      }
      c.JSON(http.StatusInternalServerError, gin.H{"error": "delete inbound failed"})
      return
    }
    c.JSON(http.StatusOK, gin.H{"status": "ok"})
  }
}

func toInboundDTO(inb db.Inbound) inboundDTO {
  return inboundDTO{
    ID:                inb.ID,
    UUID:              inb.UUID,
    Tag:               inb.Tag,
    NodeID:            inb.NodeID,
    Protocol:          inb.Protocol,
    ListenPort:        inb.ListenPort,
    PublicPort:        inb.PublicPort,
    Settings:          inb.Settings,
    TLSSettings:       inb.TLSSettings,
    TransportSettings: inb.TransportSettings,
  }
}

