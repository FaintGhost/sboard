package node

import (
  "crypto/sha256"
  "encoding/base64"
  "encoding/json"
  "errors"
  "fmt"
  "strings"

  "sboard/panel/internal/db"
  "github.com/google/uuid"
)

type SyncPayload struct {
  Inbounds []map[string]any `json:"inbounds"`
}

func BuildSyncPayload(node db.Node, inbounds []db.Inbound, users []db.User) (SyncPayload, error) {
  out := SyncPayload{Inbounds: make([]map[string]any, 0, len(inbounds))}
  for _, inb := range inbounds {
    typ := strings.TrimSpace(inb.Protocol)
    if typ == "" {
      return SyncPayload{}, errors.New("missing inbound protocol")
    }
    tag := strings.TrimSpace(inb.Tag)
    if tag == "" {
      return SyncPayload{}, errors.New("missing inbound tag")
    }
    if inb.ListenPort <= 0 {
      return SyncPayload{}, fmt.Errorf("invalid listen port for %s", tag)
    }

    settings := map[string]any{}
    if len(inb.Settings) > 0 {
      if err := json.Unmarshal(inb.Settings, &settings); err != nil {
        return SyncPayload{}, fmt.Errorf("invalid settings for %s: %w", tag, err)
      }
    }

    // Shadowsocks: sing-box requires `method` and (for most methods) `password`.
    // To keep the Panel DB simple, we allow omitting password in DB and derive
    // deterministic values during sync so both node config and subscriptions can match.
    if typ == "shadowsocks" {
      method, _ := settings["method"].(string)
      method = strings.TrimSpace(method)
      if method != "" {
        // Fill top-level password if missing (required for 2022 methods; also required
        // for classic methods per sing-box docs; `none` is the only exception).
        if method != "none" {
          if pw, ok := settings["password"].(string); !ok || strings.TrimSpace(pw) == "" {
            derived, err := deriveShadowsocksPasswordFromUUID(inb.UUID, method)
            if err != nil {
              return SyncPayload{}, fmt.Errorf("invalid shadowsocks password seed for %s: %w", tag, err)
            }
            settings["password"] = derived
          }
        }
      }
    }

    item := map[string]any{
      "type":        typ,
      "tag":         tag,
      "listen":      "0.0.0.0",
      "listen_port": inb.ListenPort,
    }

    // Protocol-specific users list.
    item["users"] = buildUsersForProtocol(typ, users, settings)

    // Merge settings to top-level, letting explicit core keys win.
    for k, v := range settings {
      if _, exists := item[k]; exists {
        continue
      }
      item[k] = v
    }

    if len(inb.TLSSettings) > 0 {
      tls := map[string]any{}
      if err := json.Unmarshal(inb.TLSSettings, &tls); err != nil {
        return SyncPayload{}, fmt.Errorf("invalid tls_settings for %s: %w", tag, err)
      }
      item["tls"] = tls
    }
    if len(inb.TransportSettings) > 0 {
      transport := map[string]any{}
      if err := json.Unmarshal(inb.TransportSettings, &transport); err != nil {
        return SyncPayload{}, fmt.Errorf("invalid transport_settings for %s: %w", tag, err)
      }
      item["transport"] = transport
    }

    out.Inbounds = append(out.Inbounds, item)
  }
  return out, nil
}

func buildUsersForProtocol(protocol string, users []db.User, settings map[string]any) []map[string]any {
  out := make([]map[string]any, 0, len(users))
  flow, _ := settings["flow"].(string)
  flow = strings.TrimSpace(flow)
  method, _ := settings["method"].(string)
  method = strings.TrimSpace(method)
  for _, u := range users {
    name := u.Username
    if name == "" {
      name = u.UUID
    }
    switch protocol {
    case "vless", "vmess":
      item := map[string]any{"name": name, "uuid": u.UUID}
      if protocol == "vless" && flow != "" {
        item["flow"] = flow
      }
      out = append(out, item)
    case "trojan", "shadowsocks":
      if protocol == "shadowsocks" && strings.HasPrefix(method, "2022-") {
        // 2022 methods require base64 keys (16/32 bytes).
        pw, err := deriveShadowsocksUserPasswordFromUUID(u.UUID, method)
        if err == nil && pw != "" {
          out = append(out, map[string]any{"name": name, "password": pw})
          continue
        }
      }
      out = append(out, map[string]any{"name": name, "password": u.UUID})
    default:
      // unknown: still provide a uuid to keep behavior predictable
      out = append(out, map[string]any{"name": name, "uuid": u.UUID})
    }
  }
  return out
}

func ss2022KeyLength(method string) int {
  switch strings.TrimSpace(method) {
  case "2022-blake3-aes-128-gcm":
    return 16
  case "2022-blake3-aes-256-gcm", "2022-blake3-chacha20-poly1305":
    return 32
  default:
    return 0
  }
}

func deriveShadowsocksPasswordFromUUID(uuidStr, method string) (string, error) {
  // For non-2022 methods, any string works; we still prefer a deterministic value.
  keyLen := ss2022KeyLength(method)
  if keyLen == 0 {
    // Classic methods still require password; use UUID string.
    return uuidStr, nil
  }
  return deriveBase64KeyFromUUID(uuidStr, keyLen)
}

func deriveShadowsocksUserPasswordFromUUID(uuidStr, method string) (string, error) {
  keyLen := ss2022KeyLength(method)
  if keyLen == 0 {
    return "", nil
  }
  return deriveBase64KeyFromUUID(uuidStr, keyLen)
}

func deriveBase64KeyFromUUID(uuidStr string, keyLen int) (string, error) {
  id, err := uuid.Parse(strings.TrimSpace(uuidStr))
  if err != nil {
    return "", err
  }
  b := id[:]
  if keyLen == 16 {
    return base64.StdEncoding.EncodeToString(b), nil
  }
  if keyLen == 32 {
    sum := sha256.Sum256(b)
    return base64.StdEncoding.EncodeToString(sum[:]), nil
  }
  return "", errors.New("unsupported key length")
}
