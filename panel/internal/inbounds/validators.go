package inbounds

import (
  "errors"
  "fmt"
  "strings"
  "sync"
)

// SettingsValidator validates inbound "settings" JSON for a specific protocol.
// This is intentionally pluggable so future protocols can register their own validation
// without touching the API layer.
type SettingsValidator func(settings map[string]any) error

var (
  mu         sync.RWMutex
  validators = map[string]SettingsValidator{
    "shadowsocks": validateShadowsocksSettings,
    "vless":       validateVLESSettings,
    "vmess":       validateVmessSettings,
    "trojan":      validateTrojanSettings,
    "socks":       validateSOCKSSettings,
    "http":        validateSOCKSSettings,
    "mixed":       validateSOCKSSettings,
    "hysteria2":   validateHysteria2Settings,
    "tuic":        validateTUICSettings,
    "naive":       validateNaiveSettings,
    "shadowtls":    validateShadowTLSSettings,
    "anytls":      validateAnyTLSSettings,
  }
)

func RegisterSettingsValidator(protocol string, v SettingsValidator) {
  key := strings.TrimSpace(strings.ToLower(protocol))
  if key == "" || v == nil {
    return
  }
  mu.Lock()
  defer mu.Unlock()
  validators[key] = v
}

func ValidateSettings(protocol string, settings map[string]any) error {
  key := strings.TrimSpace(strings.ToLower(protocol))
  if key == "" {
    return nil
  }
  mu.RLock()
  v := validators[key]
  mu.RUnlock()
  if v == nil {
    return nil
  }
  return v(settings)
}

// ss2022MultiUserMethods lists the SS2022 methods that support multi-user mode.
// sing-box's shadowaead_2022.NewMultiService only supports aes-128-gcm and aes-256-gcm;
// chacha20-poly1305 is single-user only.
var ss2022MultiUserMethods = map[string]struct{}{
  "2022-blake3-aes-128-gcm": {},
  "2022-blake3-aes-256-gcm": {},
}

func validateShadowsocksSettings(settings map[string]any) error {
  method, _ := settings["method"].(string)
  method = strings.TrimSpace(method)
  if method == "" {
    return errors.New("shadowsocks settings.method required")
  }
  // Our system always injects a users list (multi-user mode).
  // chacha20-poly1305 does not support multi-user in sing-box.
  if strings.HasPrefix(method, "2022-") {
    if _, ok := ss2022MultiUserMethods[method]; !ok {
      return errors.New("shadowsocks method " + method + " does not support multi-user mode; use 2022-blake3-aes-128-gcm or 2022-blake3-aes-256-gcm")
    }
  }
  return nil
}

func validateVLESSettings(settings map[string]any) error {
  users, _ := settings["users"].([]any)
  if len(users) == 0 {
    return errors.New("vless settings.users is required and must not be empty")
  }
  for i, u := range users {
    uMap, ok := u.(map[string]any)
    if !ok {
      return errors.New("vless settings.users[" + fmt.Sprintf("%d", i) + "] must be an object")
    }
    uuid, _ := uMap["uuid"].(string)
    if strings.TrimSpace(uuid) == "" {
      return errors.New("vless settings.users[" + fmt.Sprintf("%d", i) + "].uuid is required")
    }
    flow, _ := uMap["flow"].(string)
    flow = strings.TrimSpace(flow)
    if flow != "" && flow != "xtls-rprx-vision" {
      return errors.New("vless settings.users[" + fmt.Sprintf("%d", i) + "].flow must be empty or \"xtls-rprx-vision\"")
    }
  }
  return nil
}

func validateVmessSettings(settings map[string]any) error {
  users, _ := settings["users"].([]any)
  if len(users) == 0 {
    return errors.New("vmess settings.users is required and must not be empty")
  }
  for i, u := range users {
    uMap, ok := u.(map[string]any)
    if !ok {
      return errors.New("vmess settings.users[" + fmt.Sprintf("%d", i) + "] must be an object")
    }
    uuid, _ := uMap["uuid"].(string)
    if strings.TrimSpace(uuid) == "" {
      return errors.New("vmess settings.users[" + fmt.Sprintf("%d", i) + "].uuid is required")
    }
    if alterId, ok := uMap["alterId"].(float64); ok && alterId < 0 {
      return errors.New("vmess settings.users[" + fmt.Sprintf("%d", i) + "].alterId must be >= 0")
    }
  }
  return nil
}

func validateTrojanSettings(settings map[string]any) error {
  users, _ := settings["users"].([]any)
  if len(users) == 0 {
    return errors.New("trojan settings.users is required and must not be empty")
  }
  for i, u := range users {
    uMap, ok := u.(map[string]any)
    if !ok {
      return errors.New("trojan settings.users[" + fmt.Sprintf("%d", i) + "] must be an object")
    }
    password, _ := uMap["password"].(string)
    if strings.TrimSpace(password) == "" {
      return errors.New("trojan settings.users[" + fmt.Sprintf("%d", i) + "].password is required")
    }
  }
  return nil
}

func validateSOCKSSettings(settings map[string]any) error {
  users, _ := settings["users"].([]any)
  for i, u := range users {
    uMap, ok := u.(map[string]any)
    if !ok {
      return errors.New("socks settings.users[" + fmt.Sprintf("%d", i) + "] must be an object")
    }
    username, _ := uMap["username"].(string)
    password, _ := uMap["password"].(string)
    if strings.TrimSpace(username) == "" {
      return errors.New("socks settings.users[" + fmt.Sprintf("%d", i) + "].username is required when user is specified")
    }
    if strings.TrimSpace(password) == "" {
      return errors.New("socks settings.users[" + fmt.Sprintf("%d", i) + "].password is required when user is specified")
    }
  }
  return nil
}

func validateHysteria2Settings(settings map[string]any) error {
  users, _ := settings["users"].([]any)
  if len(users) == 0 {
    return errors.New("hysteria2 settings.users is required and must not be empty")
  }
  for i, u := range users {
    uMap, ok := u.(map[string]any)
    if !ok {
      return errors.New("hysteria2 settings.users[" + fmt.Sprintf("%d", i) + "] must be an object")
    }
    password, _ := uMap["password"].(string)
    if strings.TrimSpace(password) == "" {
      return errors.New("hysteria2 settings.users[" + fmt.Sprintf("%d", i) + "].password is required")
    }
  }
  return nil
}

func validateTUICSettings(settings map[string]any) error {
  users, _ := settings["users"].([]any)
  if len(users) == 0 {
    return errors.New("tuic settings.users is required and must not be empty")
  }
  for i, u := range users {
    uMap, ok := u.(map[string]any)
    if !ok {
      return errors.New("tuic settings.users[" + fmt.Sprintf("%d", i) + "] must be an object")
    }
    uuid, _ := uMap["uuid"].(string)
    if strings.TrimSpace(uuid) == "" {
      return errors.New("tuic settings.users[" + fmt.Sprintf("%d", i) + "].uuid is required")
    }
    password, _ := uMap["password"].(string)
    if strings.TrimSpace(password) == "" {
      return errors.New("tuic settings.users[" + fmt.Sprintf("%d", i) + "].password is required")
    }
  }
  return nil
}

func validateNaiveSettings(settings map[string]any) error {
  users, _ := settings["users"].([]any)
  if len(users) == 0 {
    return errors.New("naive settings.users is required and must not be empty")
  }
  for i, u := range users {
    uMap, ok := u.(map[string]any)
    if !ok {
      return errors.New("naive settings.users[" + fmt.Sprintf("%d", i) + "] must be an object")
    }
    username, _ := uMap["username"].(string)
    if strings.TrimSpace(username) == "" {
      return errors.New("naive settings.users[" + fmt.Sprintf("%d", i) + "].username is required")
    }
    password, _ := uMap["password"].(string)
    if strings.TrimSpace(password) == "" {
      return errors.New("naive settings.users[" + fmt.Sprintf("%d", i) + "].password is required")
    }
  }
  return nil
}

func validateShadowTLSSettings(settings map[string]any) error {
  users, _ := settings["users"].([]any)
  if len(users) == 0 {
    return errors.New("shadowtls settings.users is required and must not be empty")
  }
  for i, u := range users {
    uMap, ok := u.(map[string]any)
    if !ok {
      return errors.New("shadowtls settings.users[" + fmt.Sprintf("%d", i) + "] must be an object")
    }
    password, _ := uMap["password"].(string)
    if strings.TrimSpace(password) == "" {
      return errors.New("shadowtls settings.users[" + fmt.Sprintf("%d", i) + "].password is required")
    }
  }
  handshake, _ := settings["handshake"].(map[string]any)
  if handshake == nil {
    return errors.New("shadowtls settings.handshake is required")
  }
  server, _ := handshake["server"].(string)
  if strings.TrimSpace(server) == "" {
    return errors.New("shadowtls settings.handshake.server is required")
  }
  return nil
}

func validateAnyTLSSettings(settings map[string]any) error {
  users, _ := settings["users"].([]any)
  if len(users) == 0 {
    return errors.New("anytls settings.users is required and must not be empty")
  }
  for i, u := range users {
    uMap, ok := u.(map[string]any)
    if !ok {
      return errors.New("anytls settings.users[" + fmt.Sprintf("%d", i) + "] must be an object")
    }
    password, _ := uMap["password"].(string)
    if strings.TrimSpace(password) == "" {
      return errors.New("anytls settings.users[" + fmt.Sprintf("%d", i) + "].password is required")
    }
  }
  return nil
}

