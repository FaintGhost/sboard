package subscription_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/subscription"
)

func TestSingboxGenerateOutbounds(t *testing.T) {
	user := subscription.User{UUID: "u-1", Username: "alice"}
	items := []subscription.Item{
		{
			InboundType:       "vless",
			NodePublicAddress: "a.example.com",
			InboundListenPort: 443,
			InboundPublicPort: 0,
			Settings:          json.RawMessage(`{"flow":"xtls-rprx-vision"}`),
		},
	}

	out, err := subscription.BuildSingbox(user, items)
	require.NoError(t, err)
	require.Contains(t, string(out), "a.example.com")
	require.Contains(t, string(out), "vless")
}

func TestSingboxFiltersInternalConfigField(t *testing.T) {
	user := subscription.User{UUID: "u-1", Username: "alice"}
	items := []subscription.Item{
		{
			InboundType:       "shadowsocks",
			NodePublicAddress: "a.example.com",
			InboundListenPort: 443,
			// Settings contains __config which should be filtered out
			Settings: json.RawMessage(`{"method":"2022-blake3-aes-256-gcm","__config":{"log":{},"dns":{}}}`),
		},
	}

	out, err := subscription.BuildSingbox(user, items)
	require.NoError(t, err)

	// __config should not appear in output
	require.NotContains(t, string(out), "__config")

	// Verify it's valid JSON and has expected fields
	var payload struct {
		Outbounds []map[string]any `json:"outbounds"`
	}
	require.NoError(t, json.Unmarshal(out, &payload))
	require.Len(t, payload.Outbounds, 1)

	outbound := payload.Outbounds[0]
	_, hasConfig := outbound["__config"]
	require.False(t, hasConfig, "__config should be filtered from outbound")
	require.Equal(t, "2022-blake3-aes-256-gcm", outbound["method"])
}

func TestSingboxVlessRealityTLSConvertsToOutboundSyntax(t *testing.T) {
	user := subscription.User{UUID: "6fd47678-1f45-48f1-8051-fdbcdc2a3ccb", Username: "alice"}
	items := []subscription.Item{
		{
			InboundType:       "vless",
			InboundTag:        "VLESS_TCP_REALITY",
			NodePublicAddress: "174.136.204.13",
			InboundListenPort: 443,
			Settings:          json.RawMessage(`{"flow":"xtls-rprx-vision","sniff":true,"sniff_override_destination":true}`),
			TLSSettings: json.RawMessage(`{
				"enabled": true,
				"server_name": "aws.amazon.com",
				"reality": {
					"enabled": true,
					"handshake": {"server":"aws.amazon.com","server_port":443},
					"private_key": "j07u-NRZcenXO3VD8u5QnZdpO2Dd-usrntJ6vh2UbLs",
					"short_id": ["f02ea9614ec73e47"]
				}
			}`),
		},
	}

	out, err := subscription.BuildSingbox(user, items)
	require.NoError(t, err)

	var payload struct {
		Outbounds []map[string]any `json:"outbounds"`
	}
	require.NoError(t, json.Unmarshal(out, &payload))
	require.Len(t, payload.Outbounds, 1)

	ob := payload.Outbounds[0]
	require.Equal(t, "xtls-rprx-vision", ob["flow"])
	_, hasSniff := ob["sniff"]
	require.False(t, hasSniff, "outbound should not contain inbound-only field sniff")
	_, hasSniffOverride := ob["sniff_override_destination"]
	require.False(t, hasSniffOverride, "outbound should not contain inbound-only field sniff_override_destination")
	tls, ok := ob["tls"].(map[string]any)
	require.True(t, ok)
	reality, ok := tls["reality"].(map[string]any)
	require.True(t, ok)

	_, hasPrivate := reality["private_key"]
	require.False(t, hasPrivate, "outbound reality must not contain private_key")
	_, hasHandshake := reality["handshake"]
	require.False(t, hasHandshake, "outbound reality should not contain handshake")

	publicKey, _ := reality["public_key"].(string)
	require.NotEmpty(t, publicKey, "outbound reality should contain derived public_key")

	sid, sidOK := reality["short_id"].(string)
	require.True(t, sidOK)
	require.Equal(t, "f02ea9614ec73e47", sid)

	utls, utlsOK := tls["utls"].(map[string]any)
	require.True(t, utlsOK)
	require.Equal(t, true, utls["enabled"])
	require.Equal(t, "chrome", utls["fingerprint"])
}

func TestSingboxVlessWithoutFlowShouldNotInjectFlow(t *testing.T) {
	user := subscription.User{UUID: "u-1", Username: "alice"}
	items := []subscription.Item{
		{
			InboundType:       "vless",
			NodePublicAddress: "a.example.com",
			InboundListenPort: 443,
			Settings:          json.RawMessage(`{"sniff":true}`),
			TLSSettings:       json.RawMessage(`{"enabled":true,"reality":{"enabled":true,"public_key":"abc"}}`),
		},
	}

	out, err := subscription.BuildSingbox(user, items)
	require.NoError(t, err)

	var payload struct {
		Outbounds []map[string]any `json:"outbounds"`
	}
	require.NoError(t, json.Unmarshal(out, &payload))
	require.Len(t, payload.Outbounds, 1)
	_, hasFlow := payload.Outbounds[0]["flow"]
	require.False(t, hasFlow)
}

func TestSingboxRemovesInboundOnlySniffFields(t *testing.T) {
	user := subscription.User{UUID: "u-1", Username: "alice"}
	items := []subscription.Item{
		{
			InboundType:       "vmess",
			NodePublicAddress: "a.example.com",
			InboundListenPort: 443,
			Settings:          json.RawMessage(`{"sniff":true,"sniff_override_destination":true,"sniff_timeout":"300ms"}`),
		},
	}

	out, err := subscription.BuildSingbox(user, items)
	require.NoError(t, err)

	var payload struct {
		Outbounds []map[string]any `json:"outbounds"`
	}
	require.NoError(t, json.Unmarshal(out, &payload))
	require.Len(t, payload.Outbounds, 1)

	ob := payload.Outbounds[0]
	_, hasSniff := ob["sniff"]
	require.False(t, hasSniff)
	_, hasSniffOverride := ob["sniff_override_destination"]
	require.False(t, hasSniffOverride)
	_, hasSniffTimeout := ob["sniff_timeout"]
	require.False(t, hasSniffTimeout)
}

func TestSingboxShadowsocks2022UsesPerUserPassword(t *testing.T) {
	inboundUUID := "11111111-1111-4111-8111-111111111111"
	userA := subscription.User{UUID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", Username: "alice"}
	userB := subscription.User{UUID: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb", Username: "bob"}
	items := []subscription.Item{
		{
			InboundUUID:       inboundUUID,
			InboundType:       "shadowsocks",
			NodePublicAddress: "a.example.com",
			InboundListenPort: 8388,
			Settings:          json.RawMessage(`{"method":"2022-blake3-aes-128-gcm"}`),
		},
	}

	outA, err := subscription.BuildSingbox(userA, items)
	require.NoError(t, err)
	outB, err := subscription.BuildSingbox(userB, items)
	require.NoError(t, err)

	readPassword := func(raw []byte) string {
		var payload struct {
			Outbounds []map[string]any `json:"outbounds"`
		}
		require.NoError(t, json.Unmarshal(raw, &payload))
		require.Len(t, payload.Outbounds, 1)
		password, _ := payload.Outbounds[0]["password"].(string)
		return password
	}

	passwordA := readPassword(outA)
	passwordB := readPassword(outB)
	require.NotEmpty(t, passwordA)
	require.NotEmpty(t, passwordB)
	require.NotEqual(t, passwordA, passwordB)

	partsA := strings.Split(passwordA, ":")
	partsB := strings.Split(passwordB, ":")
	require.Len(t, partsA, 2)
	require.Len(t, partsB, 2)
	// server psk should be stable per inbound; user key should differ by user.
	require.Equal(t, partsA[0], partsB[0])
	require.NotEqual(t, partsA[1], partsB[1])
}
