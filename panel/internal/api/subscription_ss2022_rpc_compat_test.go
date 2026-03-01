package api_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/api"
	"sboard/panel/internal/config"
	"sboard/panel/internal/db"
)

func TestSubscriptionSS2022RPCCompatibility(t *testing.T) {
	store := setupSubscriptionStore(t)
	ctx := t.Context()

	user, err := store.CreateUser(ctx, "alice")
	require.NoError(t, err)

	group, err := store.CreateGroup(ctx, "g-ss2022", "")
	require.NoError(t, err)
	require.NoError(t, store.ReplaceGroupUsers(ctx, group.ID, []int64{user.ID}))

	nodeItem, err := store.CreateNode(ctx, db.NodeCreate{
		Name:          "node-ss2022",
		APIAddress:    "node",
		APIPort:       3000,
		SecretKey:     "secret",
		PublicAddress: "node.example.com",
		GroupID:       &group.ID,
	})
	require.NoError(t, err)

	_, err = store.CreateInbound(ctx, db.InboundCreate{
		NodeID:     nodeItem.ID,
		Tag:        "ss2022-in",
		Protocol:   "shadowsocks",
		ListenPort: 8388,
		PublicPort: 8388,
		Settings:   []byte(`{"method":"2022-blake3-aes-128-gcm"}`),
	})
	require.NoError(t, err)

	r := api.NewRouter(config.Config{}, store)
	req := httptest.NewRequest(http.MethodGet, "/api/sub/"+user.UUID+"?format=singbox", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Header().Get("Content-Type"), "application/json")

	var payload struct {
		Outbounds []map[string]any `json:"outbounds"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &payload))
	require.NotEmpty(t, payload.Outbounds)

	var ssOutbound map[string]any
	for _, outbound := range payload.Outbounds {
		if strings.EqualFold(fmt.Sprintf("%v", outbound["type"]), "shadowsocks") {
			ssOutbound = outbound
			break
		}
	}
	require.NotNil(t, ssOutbound)

	password, ok := ssOutbound["password"].(string)
	require.True(t, ok)
	parts := strings.SplitN(password, ":", 2)
	require.Len(t, parts, 2)

	serverKey, err := base64.StdEncoding.DecodeString(parts[0])
	require.NoError(t, err)
	userKey, err := base64.StdEncoding.DecodeString(parts[1])
	require.NoError(t, err)
	require.NotEmpty(t, serverKey)
	require.NotEmpty(t, userKey)
}
