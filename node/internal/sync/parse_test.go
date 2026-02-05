package sync_test

import (
  "testing"

  "sboard/node/internal/sync"
  "github.com/stretchr/testify/require"
)

func TestParseAndValidateInbounds(t *testing.T) {
  ctx := sync.NewSingboxContext()
  body := []byte(`{
    "inbounds": [
      {"type":"mixed","tag":"m1","listen":"0.0.0.0","listen_port":1080}
    ]
  }`)

  inbounds, err := sync.ParseAndValidateInbounds(ctx, body)
  require.NoError(t, err)
  require.Len(t, inbounds, 1)
  require.Equal(t, "mixed", inbounds[0].Type)
  require.Equal(t, "m1", inbounds[0].Tag)
}
