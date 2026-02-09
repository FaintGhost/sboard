package sync

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseAndValidateConfig_WrapsUnmarshalErrorWithMeta(t *testing.T) {
	ctx := NewSingboxContext()
	body := []byte(`{
    "inbounds": [
      {
        "type": "vless",
        "tag": "vless-in",
        "listen": "0.0.0.0",
        "listen_port": 443,
        "users": "not-an-array"
      }
    ]
  }`)
	_, err := ParseAndValidateConfig(ctx, body)
	require.Error(t, err)
	require.Contains(t, err.Error(), "inbounds[0]")
	require.Contains(t, err.Error(), "tag=vless-in")
	require.Contains(t, err.Error(), "type=vless")
}

func TestParseAndValidateConfig_ParsesFullConfig(t *testing.T) {
	ctx := NewSingboxContext()
	body := []byte(`{
    "log": {"level":"info"},
    "dns": {},
    "inbounds": [
      {
        "type": "shadowsocks",
        "tag": "ss-in",
        "listen": "0.0.0.0",
        "listen_port": 8388,
        "method": "2022-blake3-aes-128-gcm",
        "password": "8JCsPssfgS8tiRwiMlhARg==",
        "users": []
      }
    ],
    "outbounds": [
      {"type":"direct","tag":"direct"}
    ],
    "route": {"final":"direct"}
  }`)

	options, err := ParseAndValidateConfig(ctx, body)
	require.NoError(t, err)
	require.Len(t, options.Inbounds, 1)
	require.Len(t, options.Outbounds, 1)
	require.NotNil(t, options.Route)
	require.Equal(t, "direct", options.Route.Final)
}
