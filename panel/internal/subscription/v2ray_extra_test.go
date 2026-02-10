package subscription_test

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/subscription"
)

func TestBuildV2Ray_MultiProtocols(t *testing.T) {
	user := subscription.User{UUID: "user-uuid-1234", Username: "alice"}
	items := []subscription.Item{
		{
			InboundType:       "vmess",
			InboundTag:        "vmess-in",
			NodePublicAddress: "vmess.example.com",
			InboundListenPort: 443,
		},
		{
			InboundType:       "trojan",
			InboundTag:        "trojan-in",
			NodePublicAddress: "trojan.example.com",
			InboundListenPort: 8443,
		},
		{
			InboundType:       "shadowsocks",
			InboundTag:        "ss-in",
			NodePublicAddress: "ss.example.com",
			InboundListenPort: 8388,
			Settings:          []byte(`{"method":"aes-256-gcm"}`),
		},
	}

	out, err := subscription.BuildV2Ray(user, items)
	require.NoError(t, err)

	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(out)))
	require.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	require.Len(t, lines, 3)

	require.True(t, strings.HasPrefix(lines[0], "vmess://"))
	vmessRaw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(lines[0], "vmess://"))
	require.NoError(t, err)
	vmess := map[string]string{}
	require.NoError(t, json.Unmarshal(vmessRaw, &vmess))
	require.Equal(t, user.UUID, vmess["id"])
	require.Equal(t, "vmess.example.com", vmess["add"])
	require.Equal(t, "vmess-in", vmess["ps"])

	require.Contains(t, lines[1], "trojan://")
	require.Contains(t, lines[1], user.UUID+"@trojan.example.com:8443")

	require.True(t, strings.HasPrefix(lines[2], "ss://"))
	authPart := strings.TrimPrefix(lines[2], "ss://")
	idx := strings.Index(authPart, "@")
	require.Greater(t, idx, 0)
	authRaw, err := base64.RawURLEncoding.DecodeString(authPart[:idx])
	require.NoError(t, err)
	require.Equal(t, "aes-256-gcm:"+user.UUID, string(authRaw))
}

func TestBuildV2Ray_ValidationAndUnknownSkip(t *testing.T) {
	_, err := subscription.BuildV2Ray(subscription.User{}, nil)
	require.ErrorContains(t, err, "missing user uuid")

	user := subscription.User{UUID: "u-1"}

	_, err = subscription.BuildV2Ray(user, []subscription.Item{{
		InboundType:       "vless",
		NodePublicAddress: "",
		InboundListenPort: 443,
	}})
	require.ErrorContains(t, err, "missing node public address")

	_, err = subscription.BuildV2Ray(user, []subscription.Item{{
		InboundType:       "vless",
		NodePublicAddress: "a.example.com",
		InboundListenPort: 0,
	}})
	require.ErrorContains(t, err, "invalid inbound port")

	_, err = subscription.BuildV2Ray(user, []subscription.Item{{
		InboundType:       "vless",
		NodePublicAddress: "a.example.com",
		InboundListenPort: 443,
		Settings:          []byte(`{"x":`),
	}})
	require.ErrorContains(t, err, "invalid settings")

	out, err := subscription.BuildV2Ray(user, []subscription.Item{{
		InboundType:       "unknown",
		NodePublicAddress: "a.example.com",
		InboundListenPort: 443,
	}})
	require.NoError(t, err)
	raw, err := base64.StdEncoding.DecodeString(string(out))
	require.NoError(t, err)
	require.Equal(t, "", string(raw))
}
