package sync

import (
  "testing"

  "github.com/stretchr/testify/require"
)

func TestParseAndValidateInbounds_WrapsUnmarshalErrorWithMeta(t *testing.T) {
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
  _, err := ParseAndValidateInbounds(ctx, body)
  require.Error(t, err)
  require.Contains(t, err.Error(), "inbounds[0]")
  require.Contains(t, err.Error(), "tag=vless-in")
  require.Contains(t, err.Error(), "type=vless")
}
