package node

import (
  "encoding/json"
  "errors"
  "fmt"
  "strings"

  "sboard/panel/internal/db"
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
      out = append(out, map[string]any{"name": name, "password": u.UUID})
    default:
      // unknown: still provide a uuid to keep behavior predictable
      out = append(out, map[string]any{"name": name, "uuid": u.UUID})
    }
  }
  return out
}

