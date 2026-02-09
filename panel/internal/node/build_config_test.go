package node

import (
	"encoding/base64"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"sboard/panel/internal/db"
)

func TestBuildSyncPayload_Shadowsocks2022_DerivesPasswordAndUserKeys(t *testing.T) {
	inbUUID := "11111111-1111-1111-1111-111111111111"
	userUUID := "22222222-2222-2222-2222-222222222222"

	payload, err := BuildSyncPayload(db.Node{}, []db.Inbound{
		{
			UUID:       inbUUID,
			Tag:        "ss-in",
			NodeID:     1,
			Protocol:   "shadowsocks",
			ListenPort: 443,
			Settings:   []byte(`{"method":"2022-blake3-aes-128-gcm"}`),
		},
	}, []db.User{
		{UUID: userUUID, Username: "alice"},
	})
	require.NoError(t, err)
	require.Len(t, payload.Inbounds, 1)

	item := payload.Inbounds[0]
	require.Equal(t, "shadowsocks", item["type"])
	require.Equal(t, "ss-in", item["tag"])
	require.Equal(t, "2022-blake3-aes-128-gcm", item["method"])

	inbID := uuid.MustParse(inbUUID)
	expectedPassword := base64.StdEncoding.EncodeToString(inbID[:])
	require.Equal(t, expectedPassword, item["password"])

	users, ok := item["users"].([]map[string]any)
	require.True(t, ok)
	require.Len(t, users, 1)

	u := users[0]
	require.Equal(t, "alice", u["name"])
	userID := uuid.MustParse(userUUID)
	expectedUserPassword := base64.StdEncoding.EncodeToString(userID[:])
	require.Equal(t, expectedUserPassword, u["password"])
}

func TestBuildSyncPayload_MergesGlobalConfigFromTemplate(t *testing.T) {
	payload, err := BuildSyncPayload(db.Node{}, []db.Inbound{
		{
			UUID:       "11111111-1111-1111-1111-111111111111",
			Tag:        "ss-in",
			NodeID:     1,
			Protocol:   "shadowsocks",
			ListenPort: 8388,
			Settings: []byte(`{
        "method":"2022-blake3-aes-128-gcm",
        "password":"8JCsPssfgS8tiRwiMlhARg==",
        "__config":{
          "$schema":"https://sing-box.sagernet.org/schema/config.json",
          "log":{"level":"info"},
          "outbounds":[{"type":"direct","tag":"direct"}],
          "route":{"final":"direct"}
        }
      }`),
		},
	}, nil)
	require.NoError(t, err)
	require.Equal(t, "https://sing-box.sagernet.org/schema/config.json", payload.Schema)
	require.Equal(t, "info", payload.Log["level"])
	require.Len(t, payload.Outbounds, 1)
	require.NotNil(t, payload.Route)
	require.Equal(t, "direct", payload.Route["final"])
}

func TestBuildSyncPayload_MergesGlobalConfigOnlyOnce(t *testing.T) {
	payload, err := BuildSyncPayload(db.Node{}, []db.Inbound{
		{
			UUID:       "11111111-1111-1111-1111-111111111111",
			Tag:        "vless-a",
			NodeID:     1,
			Protocol:   "vless",
			ListenPort: 443,
			Settings: []byte(`{
        "flow":"xtls-rprx-vision",
        "__config":{
          "outbounds":[{"type":"direct","tag":"direct"}],
          "route":{"final":"direct"}
        }
      }`),
		},
		{
			UUID:       "22222222-2222-2222-2222-222222222222",
			Tag:        "vless-b",
			NodeID:     1,
			Protocol:   "vless",
			ListenPort: 8443,
			Settings: []byte(`{
        "flow":"xtls-rprx-vision",
        "__config":{
          "outbounds":[{"type":"direct","tag":"direct"}],
          "route":{"final":"direct"}
        }
      }`),
		},
	}, nil)
	require.NoError(t, err)
	require.Len(t, payload.Inbounds, 2)
	require.Len(t, payload.Outbounds, 1)
	require.Equal(t, "direct", payload.Route["final"])
}
