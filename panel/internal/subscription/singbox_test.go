package subscription_test

import (
  "encoding/json"
  "testing"

  "sboard/panel/internal/subscription"
  "github.com/stretchr/testify/require"
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
