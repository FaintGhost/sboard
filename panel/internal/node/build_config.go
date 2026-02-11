package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"sboard/panel/internal/db"
	"sboard/panel/internal/sskey"
)

const extraConfigKey = "__config"

type SyncPayload struct {
	Schema       string           `json:"$schema,omitempty"`
	Log          map[string]any   `json:"log,omitempty"`
	DNS          map[string]any   `json:"dns,omitempty"`
	NTP          map[string]any   `json:"ntp,omitempty"`
	Certificate  map[string]any   `json:"certificate,omitempty"`
	Endpoints    []any            `json:"endpoints,omitempty"`
	Inbounds     []map[string]any `json:"inbounds"`
	Outbounds    []any            `json:"outbounds,omitempty"`
	Route        map[string]any   `json:"route,omitempty"`
	Services     []any            `json:"services,omitempty"`
	Experimental map[string]any   `json:"experimental,omitempty"`
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

		mergeGlobalConfigFromSettings(settings, &out)

		if typ == "shadowsocks" {
			method, _ := settings["method"].(string)
			method = strings.TrimSpace(method)
			if method != "" {
				if method != "none" {
					if pw, ok := settings["password"].(string); !ok || strings.TrimSpace(pw) == "" {
						derived, err := sskey.DerivePassword(inb.UUID, method)
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

		item["users"] = buildUsersForProtocol(typ, users, settings)

		for k, v := range settings {
			if _, exists := item[k]; exists {
				continue
			}
			if shouldSkipSettingKeyInInbound(typ, k) {
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

func mergeGlobalConfigFromSettings(settings map[string]any, out *SyncPayload) {
	if settings == nil || out == nil {
		return
	}

	raw, ok := settings[extraConfigKey]
	if !ok {
		return
	}
	delete(settings, extraConfigKey)

	cfg, ok := raw.(map[string]any)
	if !ok {
		return
	}

	if value, ok := asMap(cfg["log"]); ok {
		if out.Log == nil {
			out.Log = value
		}
	}
	if value, ok := asString(cfg["$schema"]); ok {
		if out.Schema == "" {
			out.Schema = value
		}
	}
	if value, ok := asMap(cfg["dns"]); ok {
		if out.DNS == nil {
			out.DNS = value
		}
	}
	if value, ok := asMap(cfg["ntp"]); ok {
		if out.NTP == nil {
			out.NTP = value
		}
	}
	if value, ok := asMap(cfg["certificate"]); ok {
		if out.Certificate == nil {
			out.Certificate = value
		}
	}
	if value, ok := asArray(cfg["endpoints"]); ok {
		if len(out.Endpoints) == 0 {
			out.Endpoints = append(out.Endpoints, value...)
		}
	}
	if value, ok := asArray(cfg["outbounds"]); ok {
		if len(out.Outbounds) == 0 {
			out.Outbounds = append(out.Outbounds, value...)
		}
	}
	if value, ok := asMap(cfg["route"]); ok {
		if out.Route == nil {
			out.Route = value
		}
	}
	if value, ok := asArray(cfg["services"]); ok {
		if len(out.Services) == 0 {
			out.Services = append(out.Services, value...)
		}
	}
	if value, ok := asMap(cfg["experimental"]); ok {
		if out.Experimental == nil {
			out.Experimental = value
		}
	}
}

func asString(value any) (string, bool) {
	text, ok := value.(string)
	if !ok {
		return "", false
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return "", false
	}
	return text, true
}

func asMap(value any) (map[string]any, bool) {
	mapped, ok := value.(map[string]any)
	if !ok {
		return nil, false
	}
	return mapped, true
}

func asArray(value any) ([]any, bool) {
	arr, ok := value.([]any)
	if !ok {
		return nil, false
	}
	return arr, true
}

func shouldSkipSettingKeyInInbound(protocol string, key string) bool {
	if key == extraConfigKey {
		return true
	}
	if protocol == "vless" && key == "flow" {
		// flow is a vless user-level field, not inbound top-level field.
		return true
	}
	return false
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
			if protocol == "shadowsocks" && sskey.Is2022Method(method) {
				pw, err := sskey.DerivePassword(u.UUID, method)
				if err == nil && pw != "" {
					out = append(out, map[string]any{"name": name, "password": pw})
					continue
				}
			}
			out = append(out, map[string]any{"name": name, "password": u.UUID})
		case "socks", "http", "mixed":
			out = append(out, map[string]any{"username": name, "password": u.UUID})
		default:
			out = append(out, map[string]any{"name": name, "uuid": u.UUID})
		}
	}
	return out
}
