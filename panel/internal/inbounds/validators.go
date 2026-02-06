package inbounds

import (
  "errors"
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

