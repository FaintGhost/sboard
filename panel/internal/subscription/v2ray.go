package subscription

import (
  "encoding/base64"
  "encoding/json"
  "errors"
  "fmt"
  "net/url"
  "strconv"
  "strings"
)

func BuildV2Ray(user User, items []Item) ([]byte, error) {
  if user.UUID == "" {
    return nil, errors.New("missing user uuid")
  }

  lines := make([]string, 0, len(items))
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

    tag := item.InboundTag
    if tag == "" {
      tag = fmt.Sprintf("%s-%d", item.InboundType, idx+1)
    }

    settings := map[string]any{}
    if len(item.Settings) > 0 {
      if err := json.Unmarshal(item.Settings, &settings); err != nil {
        return nil, fmt.Errorf("invalid settings: %w", err)
      }
    }

    // Mirror sing-box behavior: fill credentials from user UUID unless already present.
    injectCredentials(user, item.InboundType, settings)

    switch item.InboundType {
    case "vless":
      lines = append(lines, buildVLESS(user, item.NodePublicAddress, port, tag, settings, len(item.TLSSettings) > 0))
    case "vmess":
      line, err := buildVMess(user, item.NodePublicAddress, port, tag, len(item.TLSSettings) > 0)
      if err != nil {
        return nil, err
      }
      lines = append(lines, line)
    case "trojan":
      lines = append(lines, buildTrojan(user, item.NodePublicAddress, port, tag))
    case "shadowsocks":
      lines = append(lines, buildShadowsocks(user, item.NodePublicAddress, port, tag, settings))
    default:
      // Unknown protocol: skip for now.
    }
  }

  raw := strings.Join(lines, "\n")
  encoded := base64.StdEncoding.EncodeToString([]byte(raw))
  return []byte(encoded), nil
}

func buildVLESS(user User, host string, port int, tag string, settings map[string]any, hasTLS bool) string {
  qp := url.Values{}
  qp.Set("encryption", "none")
  if flow, ok := settings["flow"].(string); ok && strings.TrimSpace(flow) != "" {
    qp.Set("flow", strings.TrimSpace(flow))
  }
  if hasTLS {
    qp.Set("security", "tls")
  }
  u := url.URL{
    Scheme:   "vless",
    User:     url.User(user.UUID),
    Host:     fmt.Sprintf("%s:%d", host, port),
    RawQuery: qp.Encode(),
    Fragment: tag,
  }
  return u.String()
}

func buildTrojan(user User, host string, port int, tag string) string {
  return fmt.Sprintf("trojan://%s@%s:%d#%s", url.QueryEscape(user.UUID), host, port, url.QueryEscape(tag))
}

func buildShadowsocks(user User, host string, port int, tag string, settings map[string]any) string {
  method := "aes-128-gcm"
  if m, ok := settings["method"].(string); ok && strings.TrimSpace(m) != "" {
    method = strings.TrimSpace(m)
  }
  // ss://BASE64(method:password)@host:port#tag
  auth := base64.RawURLEncoding.EncodeToString([]byte(method + ":" + user.UUID))
  return fmt.Sprintf("ss://%s@%s:%d#%s", auth, host, port, url.QueryEscape(tag))
}

func buildVMess(user User, host string, port int, tag string, hasTLS bool) (string, error) {
  payload := map[string]string{
    "v":   "2",
    "ps":  tag,
    "add": host,
    "port": strconv.Itoa(port),
    "id":  user.UUID,
    "aid": "0",
    "net": "tcp",
    "type": "none",
    "host": "",
    "path": "",
    "tls":  "",
  }
  if hasTLS {
    payload["tls"] = "tls"
  }
  b, err := json.Marshal(payload)
  if err != nil {
    return "", err
  }
  return "vmess://" + base64.StdEncoding.EncodeToString(b), nil
}
