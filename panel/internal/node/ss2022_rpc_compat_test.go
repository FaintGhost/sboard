package node

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/db"
	nodev1 "sboard/panel/internal/rpc/gen/sboard/node/v1"
)

func TestSS2022RPCSyncPayloadCompatibility(t *testing.T) {
	inboundUUID := "11111111-1111-1111-1111-111111111111"
	userUUID := "22222222-2222-2222-2222-222222222222"

	payload, err := BuildSyncPayload(db.Node{}, []db.Inbound{{
		UUID:       inboundUUID,
		Tag:        "ss-in",
		NodeID:     1,
		Protocol:   "shadowsocks",
		ListenPort: 8388,
		PublicPort: 8388,
		Settings:   []byte(`{"method":"2022-blake3-aes-128-gcm"}`),
	}}, []db.User{{
		UUID:     userUUID,
		Username: "alice",
	}})
	require.NoError(t, err)

	var captured []byte
	srv := newNodeRPCServer(t, nodeControlServiceTestServer{
		syncConfig: func(ctx context.Context, req *nodev1.SyncConfigRequest) (*nodev1.SyncConfigResponse, error) {
			captured = append([]byte(nil), req.GetPayloadJson()...)
			return &nodev1.SyncConfigResponse{Status: "ok"}, nil
		},
	}, nil)
	defer srv.Close()

	client := NewClient(srv.Client())
	require.NoError(t, client.SyncConfig(context.Background(), nodeFromServerURL(t, srv.URL), payload))
	require.NotEmpty(t, captured)

	var body map[string]any
	require.NoError(t, json.Unmarshal(captured, &body))
	inbounds, ok := body["inbounds"].([]any)
	require.True(t, ok)
	require.Len(t, inbounds, 1)

	inbound, ok := inbounds[0].(map[string]any)
	require.True(t, ok)

	serverPassword, ok := inbound["password"].(string)
	require.True(t, ok)
	serverKey, err := base64.StdEncoding.DecodeString(serverPassword)
	require.NoError(t, err)
	require.NotEmpty(t, serverKey)

	users, ok := inbound["users"].([]any)
	require.True(t, ok)
	require.Len(t, users, 1)
	user, ok := users[0].(map[string]any)
	require.True(t, ok)

	userPassword, ok := user["password"].(string)
	require.True(t, ok)
	userKey, err := base64.StdEncoding.DecodeString(userPassword)
	require.NoError(t, err)
	require.NotEmpty(t, userKey)
}
