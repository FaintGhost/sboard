package api

import (
  "context"
  "fmt"
  "net/http"
  "strings"

  "github.com/gin-gonic/gin"
  "github.com/golang-jwt/jwt/v5"
  "sboard/panel/internal/config"
  "sboard/panel/internal/db"
)

// publicOperations lists operations that do NOT require authentication.
var publicOperations = map[string]bool{
  "GetHealth":          true,
  "GetBootstrapStatus": true,
  "Bootstrap":          true,
  "Login":              true,
  "GetSubscription":    true,
}

// authStrictMiddleware returns a StrictMiddlewareFunc that enforces JWT
// authentication for all operations except those in publicOperations.
func authStrictMiddleware(secret string) StrictMiddlewareFunc {
  return func(f StrictHandlerFunc, operationID string) StrictHandlerFunc {
    if publicOperations[operationID] {
      return f
    }
    return func(ctx *gin.Context, request any) (any, error) {
      auth := ctx.GetHeader("Authorization")
      if !strings.HasPrefix(auth, "Bearer ") {
        ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        ctx.Abort()
        return nil, nil
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
        ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        ctx.Abort()
        return nil, nil
      }
      return f(ctx, request)
    }
  }
}

func NewRouter(cfg config.Config, store *db.Store) *gin.Engine {
  if store != nil {
    _ = InitSystemTimezone(context.Background(), store)
  }

  r := gin.New()
  r.Use(RequestLogger(cfg.LogRequests))
  r.Use(gin.Recovery())
  r.Use(CORSMiddleware(cfg.CORSAllowOrigins))

  singBoxTools := singBoxToolsFactory()

  server := NewServer(store, cfg, singBoxTools)
  strictHandler := NewStrictHandler(server, []StrictMiddlewareFunc{
    authStrictMiddleware(cfg.JWTSecret),
  })

  RegisterHandlersWithOptions(r, strictHandler, GinServerOptions{
    BaseURL: "/api",
    ErrorHandler: func(c *gin.Context, err error, statusCode int) {
      msg := err.Error()
      // Normalize generated parameter validation errors.
      if statusCode == http.StatusBadRequest && strings.Contains(msg, "Invalid format for parameter") {
        if strings.Contains(msg, "limit") || strings.Contains(msg, "offset") {
          msg = "invalid pagination"
        } else {
          msg = "invalid parameter"
        }
      }
      c.JSON(statusCode, gin.H{"error": msg})
    },
  })

  if cfg.ServeWeb {
    ServeWebUI(r, cfg.WebDir)
  }
  return r
}
