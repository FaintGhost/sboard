package subscription_test

import (
  "encoding/base64"
  "strings"
  "testing"

  "sboard/panel/internal/subscription"
  "github.com/stretchr/testify/require"
)

func TestBuildV2Ray_VLESS(t *testing.T) {
  user := subscription.User{UUID: "u-1", Username: "alice"}
  items := []subscription.Item{
    {
      InboundType:       "vless",
      InboundTag:        "vless-in",
      NodePublicAddress: "a.example.com",
      InboundListenPort: 443,
      Settings:          []byte(`{}`),
    },
  }
  out, err := subscription.BuildV2Ray(user, items)
  require.NoError(t, err)

  decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(out)))
  require.NoError(t, err)
  require.Contains(t, string(decoded), "vless://u-1@a.example.com:443")
}

