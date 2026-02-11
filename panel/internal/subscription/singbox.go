package subscription

import (
  "crypto/ecdh"
  "encoding/base64"
  "encoding/json"
  "errors"
  "fmt"
  "strings"

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
    sanitizeSettingsForClient(settings)

    var tls map[string]any
    if len(item.TLSSettings) > 0 {
      tls = map[string]any{}
      if err := json.Unmarshal(item.TLSSettings, &tls); err != nil {
        return nil, fmt.Errorf("invalid tls settings: %w", err)
      }
      normalizeTLSForClient(item.InboundType, tls)
    }

    for k, v := range settings {
      // Skip internal fields that shouldn't be in the output
      if k == "__config" {
        continue
      }
      outbound[k] = v
    }

    if tls != nil {
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

func sanitizeSettingsForClient(settings map[string]any) {
  if len(settings) == 0 {
    return
  }

  // Inbound-only sniff fields must not appear in outbound subscriptions.
  delete(settings, "sniff")
  delete(settings, "sniff_override_destination")
  delete(settings, "sniff_timeout")
}

func normalizeTLSForClient(inboundType string, tls map[string]any) {
  if inboundType != "vless" && inboundType != "vmess" && inboundType != "trojan" {
    return
  }

  reality, ok := tls["reality"].(map[string]any)
  if !ok || len(reality) == 0 {
    return
  }

  // Inbound uses private_key, outbound requires public_key.
  if pk, _ := reality["public_key"].(string); strings.TrimSpace(pk) == "" {
    if sk, _ := reality["private_key"].(string); strings.TrimSpace(sk) != "" {
      if derived, err := deriveRealityPublicKey(sk); err == nil && derived != "" {
        reality["public_key"] = derived
      }
    }
  }
  delete(reality, "private_key")
  delete(reality, "handshake")

  // Outbound short_id expects a string; template may carry array form.
  if sidArr, ok := reality["short_id"].([]any); ok {
    for _, item := range sidArr {
      if text, ok := item.(string); ok && strings.TrimSpace(text) != "" {
        reality["short_id"] = strings.TrimSpace(text)
        break
      }
    }
  }
  if sidArr, ok := reality["short_id"].([]string); ok && len(sidArr) > 0 {
    reality["short_id"] = strings.TrimSpace(sidArr[0])
  }

  // For reality outbound, enable a sane uTLS default if absent.
  if _, ok := tls["utls"].(map[string]any); !ok {
    tls["utls"] = map[string]any{
      "enabled":     true,
      "fingerprint": "chrome",
    }
  }
}

func deriveRealityPublicKey(privateKey string) (string, error) {
  keyText := strings.TrimSpace(privateKey)
  if keyText == "" {
    return "", errors.New("empty private key")
  }

  var raw []byte
  var err error
  for _, dec := range []*base64.Encoding{base64.RawURLEncoding, base64.URLEncoding, base64.RawStdEncoding, base64.StdEncoding} {
    raw, err = dec.DecodeString(keyText)
    if err == nil {
      break
    }
  }
  if err != nil {
    return "", err
  }

  key, err := ecdh.X25519().NewPrivateKey(raw)
  if err != nil {
    return "", err
  }
  return base64.RawURLEncoding.EncodeToString(key.PublicKey().Bytes()), nil
}
