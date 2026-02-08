package api

import (
	"testing"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/node"
)

func TestSyncPayloadDebugJSON_MasksCredentialsAndKeepsUsers(t *testing.T) {
	payload := node.SyncPayload{
		Inbounds: []map[string]any{
			{
				"type":     "shadowsocks",
				"tag":      "ss-in",
				"method":   "2022-blake3-aes-128-gcm",
				"password": "server-psk-12345678",
				"users": []map[string]any{
					{"name": "kayson", "password": "user-key-11111111"},
					{"name": "alice", "uuid": "123e4567-e89b-12d3-a456-426614174000"},
				},
			},
		},
	}

	logged := syncPayloadDebugJSON(payload)

	require.Contains(t, logged, "kayson")
	require.Contains(t, logged, "alice")
	require.NotContains(t, logged, "server-psk-12345678")
	require.NotContains(t, logged, "user-key-11111111")
	require.NotContains(t, logged, "123e4567-e89b-12d3-a456-426614174000")
}
