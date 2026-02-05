package api

import (
  "fmt"
  "net/http"
  "strings"
  "time"

  "sboard/panel/internal/config"
  "github.com/gin-gonic/gin"
  "github.com/golang-jwt/jwt/v5"
)

type loginReq struct {
  Username string `json:"username"`
  Password string `json:"password"`
}

type loginResp struct {
  Token     string `json:"token"`
  ExpiresAt string `json:"expires_at"`
}

func AdminLogin(cfg config.Config) gin.HandlerFunc {
  return func(c *gin.Context) {
    var req loginReq
    if err := c.ShouldBindJSON(&req); err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
      return
    }
    if req.Username != cfg.AdminUser || req.Password != cfg.AdminPass {
      c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
      return
    }
    token, exp, err := signAdminToken(cfg.JWTSecret)
    if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "sign token failed"})
      return
    }
    c.JSON(http.StatusOK, gin.H{"data": loginResp{Token: token, ExpiresAt: exp.Format(time.RFC3339)}})
  }
}

func signAdminToken(secret string) (string, time.Time, error) {
  now := time.Now()
  exp := now.Add(24 * time.Hour)
  claims := jwt.RegisteredClaims{
    Subject:   "admin",
    IssuedAt:  jwt.NewNumericDate(now),
    ExpiresAt: jwt.NewNumericDate(exp),
  }
  token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
  signed, err := token.SignedString([]byte(secret))
  if err != nil {
    return "", time.Time{}, err
  }
  return signed, exp, nil
}

func AuthMiddleware(secret string) gin.HandlerFunc {
  return func(c *gin.Context) {
    auth := c.GetHeader("Authorization")
    if !strings.HasPrefix(auth, "Bearer ") {
      c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
      c.Abort()
      return
    }
    tokenStr := strings.TrimPrefix(auth, "Bearer ")
    claims := &jwt.RegisteredClaims{}
    token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
      if t.Method != jwt.SigningMethodHS256 {
        return nil, fmt.Errorf("unexpected signing method")
      }
      return []byte(secret), nil
    })
    if err != nil || !token.Valid || claims.Subject != "admin" {
      c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
      c.Abort()
      return
    }
    c.Next()
  }
}
