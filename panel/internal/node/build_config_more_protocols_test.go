package node

import (
  "testing"

  "github.com/stretchr/testify/require"
  "sboard/panel/internal/db"
)

func TestBuildSyncPayload_AuthUsersForSocksHttpMixed(t *testing.T) {
  users := []db.User{
    {UUID: "11111111-1111-1111-1111-111111111111", Username: "alice"},
  }

  protocols := []string{"socks", "http", "mixed"}
  for _, protocol := range protocols {
    t.Run(protocol, func(t *testing.T) {
      payload, err := BuildSyncPayload(db.Node{}, []db.Inbound{
        {
          UUID:       "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
          Tag:        protocol + "-in",
          NodeID:     1,
          Protocol:   protocol,
          ListenPort: 1080,
          Settings:   []byte(`{}`),
        },
      }, users)
      require.NoError(t, err)
      require.Len(t, payload.Inbounds, 1)

      item := payload.Inbounds[0]
      require.Equal(t, protocol, item["type"])

      gotUsers, ok := item["users"].([]map[string]any)
      require.True(t, ok)
      require.Len(t, gotUsers, 1)
      require.Equal(t, "alice", gotUsers[0]["username"])
      require.Equal(t, users[0].UUID, gotUsers[0]["password"])
    })
  }
}
