package subscription

import (
  "encoding/json"
  "errors"
  "fmt"

  "sboard/panel/internal/sskey"
)

type User struct {
  UUID     string
  Username string
}

type Item struct {
  InboundUUID       string
  InboundType       string
  InboundTag        string
  NodePublicAddress string
  InboundListenPort int
  InboundPublicPort int
  Settings          json.RawMessage
  TLSSettings       json.RawMessage
  TransportSettings json.RawMessage
}

func BuildSingbox(user User, items []Item) ([]byte, error) {
  if user.UUID == "" {
    return nil, errors.New("missing user uuid")
  }
  outbounds := make([]map[string]any, 0, len(items))
  for idx, item := range items {
    if item.InboundType == "" {
      return nil, errors.New("missing inbound type")
    }
    if item.NodePublicAddress == "" {
      return nil, errors.New("missing node public address")
    }
    port := item.InboundListenPort
    if item.InboundPublicPort > 0 {
      port = item.InboundPublicPort
    }
    if port <= 0 {
      return nil, errors.New("invalid inbound port")
    }

    settings := map[string]any{}
    if len(item.Settings) > 0 {
      if err := json.Unmarshal(item.Settings, &settings); err != nil {
        return nil, fmt.Errorf("invalid settings: %w", err)
      }
    }

    outbound := map[string]any{
      "type":        item.InboundType,
      "server":      item.NodePublicAddress,
      "server_port": port,
    }
    tag := item.InboundTag
    if tag == "" {
      tag = fmt.Sprintf("%s-%d", item.InboundType, idx+1)
    }
    outbound["tag"] = tag

    injectCredentials(user, item.InboundUUID, item.InboundType, settings)
    for k, v := range settings {
      outbound[k] = v
    }

    if len(item.TLSSettings) > 0 {
      tls := map[string]any{}
      if err := json.Unmarshal(item.TLSSettings, &tls); err != nil {
        return nil, fmt.Errorf("invalid tls settings: %w", err)
      }
      outbound["tls"] = tls
    }
    if len(item.TransportSettings) > 0 {
      transport := map[string]any{}
      if err := json.Unmarshal(item.TransportSettings, &transport); err != nil {
        return nil, fmt.Errorf("invalid transport settings: %w", err)
      }
      outbound["transport"] = transport
    }

    outbounds = append(outbounds, outbound)
  }

  payload := map[string]any{"outbounds": outbounds}
  return json.Marshal(payload)
}

func injectCredentials(user User, inboundUUID, inboundType string, settings map[string]any) {
  switch inboundType {
  case "vless", "vmess":
    if _, ok := settings["uuid"]; !ok {
      settings["uuid"] = user.UUID
    }
  case "trojan":
    if _, ok := settings["password"]; !ok {
      settings["password"] = user.UUID
    }
  case "shadowsocks":
    method, _ := settings["method"].(string)
    serverPSK, _ := settings["password"].(string)

    if sskey.Is2022Method(method) {
      // SS2022 multi-user: client password = <server_psk>:<user_key>
      // Derive user key from user UUID
      userKey, err := sskey.DerivePassword(user.UUID, method)
      if err != nil {
        userKey = user.UUID
      }
      // If server PSK is missing in settings, derive it from inbound UUID
      if serverPSK == "" {
        serverPSK, _ = sskey.DerivePassword(inboundUUID, method)
      }
      if serverPSK != "" {
        settings["password"] = serverPSK + ":" + userKey
      } else {
        settings["password"] = userKey
      }
    } else {
      // Classic methods: just use user UUID as password
      if serverPSK == "" {
        settings["password"] = user.UUID
      }
    }
  }
}
