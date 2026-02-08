package api

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSyncPayloadDebugJSONFromBody_MasksCredentialsAndKeepsNames(t *testing.T) {
	raw := []byte(`{"inbounds":[{"type":"shadowsocks","tag":"ss-in","password":"server-psk-12345678","users":[{"name":"kayson","password":"user-key-11111111"},{"name":"alice","uuid":"123e4567-e89b-12d3-a456-426614174000"}]}]}`)

	logged := syncPayloadDebugJSONFromBody(raw)

	require.Contains(t, logged, "kayson")
	require.Contains(t, logged, "alice")
	require.NotContains(t, logged, "server-psk-12345678")
	require.NotContains(t, logged, "user-key-11111111")
	require.NotContains(t, logged, "123e4567-e89b-12d3-a456-426614174000")
}
