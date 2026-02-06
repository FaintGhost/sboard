package api

import (
  "crypto/rand"
  "encoding/base64"
  "net/http"
  "strings"

  "github.com/gin-gonic/gin"

  "sboard/panel/internal/config"
  "sboard/panel/internal/db"
  "sboard/panel/internal/password"
)

type bootstrapStatusResp struct {
  NeedsSetup bool `json:"needs_setup"`
}

type bootstrapReq struct {
  Username        string `json:"username"`
  Password        string `json:"password"`
  ConfirmPassword string `json:"confirm_password"`
  SetupToken      string `json:"setup_token"`
}

type bootstrapResp struct {
  OK bool `json:"ok"`
}

func AdminBootstrapGet(store *db.Store) gin.HandlerFunc {
  return func(c *gin.Context) {
    n, err := db.AdminCount(store)
    if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
      return
    }
    c.JSON(http.StatusOK, gin.H{"data": bootstrapStatusResp{NeedsSetup: n == 0}})
  }
}

func AdminBootstrapPost(cfg config.Config, store *db.Store) gin.HandlerFunc {
  return func(c *gin.Context) {
    var req bootstrapReq
    if err := c.ShouldBindJSON(&req); err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
      return
    }

    setupToken := strings.TrimSpace(req.SetupToken)
    if setupToken == "" {
      setupToken = strings.TrimSpace(c.GetHeader("X-Setup-Token"))
    }
    if strings.TrimSpace(cfg.SetupToken) == "" || setupToken != strings.TrimSpace(cfg.SetupToken) {
      c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
      return
    }

    username := strings.TrimSpace(req.Username)
    if username == "" {
      c.JSON(http.StatusBadRequest, gin.H{"error": "missing username"})
      return
    }
    if strings.TrimSpace(req.Password) == "" {
      c.JSON(http.StatusBadRequest, gin.H{"error": "missing password"})
      return
    }
    if req.Password != req.ConfirmPassword {
      c.JSON(http.StatusBadRequest, gin.H{"error": "passwords do not match"})
      return
    }
    if len(req.Password) < 8 {
      c.JSON(http.StatusBadRequest, gin.H{"error": "password too short"})
      return
    }

    h, err := password.Hash(req.Password)
    if err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": "invalid password"})
      return
    }
    created, err := db.AdminCreateIfNone(store, username, h)
    if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
      return
    }
    if !created {
      c.JSON(http.StatusConflict, gin.H{"error": "already initialized"})
      return
    }
    c.JSON(http.StatusOK, gin.H{"data": bootstrapResp{OK: true}})
  }
}

func GenerateSetupToken() (string, error) {
  b := make([]byte, 32)
  if _, err := rand.Read(b); err != nil {
    return "", err
  }
  return base64.RawURLEncoding.EncodeToString(b), nil
}

