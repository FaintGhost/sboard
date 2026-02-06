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

func validateShadowsocksSettings(settings map[string]any) error {
  method, _ := settings["method"].(string)
  if strings.TrimSpace(method) == "" {
    return errors.New("shadowsocks settings.method required")
  }
  return nil
}

