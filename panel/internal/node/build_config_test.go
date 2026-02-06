package node

import (
  "encoding/base64"
  "testing"

  "sboard/panel/internal/db"
  "github.com/google/uuid"
  "github.com/stretchr/testify/require"
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

  // Password should be derived from inbound UUID bytes (16 bytes) as base64.
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

