package db_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/db"
)

func TestUpdateInbound_FullAndNoop(t *testing.T) {
	store := setupStore(t)
	ctx := context.Background()

	g, err := store.CreateGroup(ctx, "g1", "")
	require.NoError(t, err)

	n, err := store.CreateNode(ctx, db.NodeCreate{
		Name:          "n1",
		APIAddress:    "api.local",
		APIPort:       2222,
		SecretKey:     "secret",
		PublicAddress: "a.example.com",
		GroupID:       &g.ID,
	})
	require.NoError(t, err)

	inb, err := store.CreateInbound(ctx, db.InboundCreate{
		Tag:               "in-1",
		NodeID:            n.ID,
		Protocol:          "vless",
		ListenPort:        443,
		PublicPort:        0,
		Settings:          json.RawMessage(`{"flow":"xtls-rprx-vision"}`),
		TLSSettings:       json.RawMessage(`{"enabled":true}`),
		TransportSettings: json.RawMessage(`{"type":"ws"}`),
	})
	require.NoError(t, err)

	newTag := "in-1-updated"
	newProto := "trojan"
	newListenPort := 8443
	newPublicPort := 9443
	newSettings := json.RawMessage(`{"password":"p@ss"}`)
	emptyJSON := json.RawMessage{}
	newTransport := json.RawMessage(`{"type":"http"}`)

	updated, err := store.UpdateInbound(ctx, inb.ID, db.InboundUpdate{
		Tag:               &newTag,
		Protocol:          &newProto,
		ListenPort:        &newListenPort,
		PublicPort:        &newPublicPort,
		Settings:          &newSettings,
		TLSSettings:       &emptyJSON,
		TransportSettings: &newTransport,
	})
	require.NoError(t, err)
	require.Equal(t, newTag, updated.Tag)
	require.Equal(t, newProto, updated.Protocol)
	require.Equal(t, newListenPort, updated.ListenPort)
	require.Equal(t, newPublicPort, updated.PublicPort)
	require.JSONEq(t, string(newSettings), string(updated.Settings))
	require.Nil(t, updated.TLSSettings)
	require.JSONEq(t, string(newTransport), string(updated.TransportSettings))

	noop, err := store.UpdateInbound(ctx, inb.ID, db.InboundUpdate{})
	require.NoError(t, err)
	require.Equal(t, updated.ID, noop.ID)
	require.Equal(t, updated.Tag, noop.Tag)
	require.Equal(t, updated.Protocol, noop.Protocol)
}

func TestUpdateInbound_NotFoundAndConflict(t *testing.T) {
	store := setupStore(t)
	ctx := context.Background()

	g, err := store.CreateGroup(ctx, "g1", "")
	require.NoError(t, err)

	n, err := store.CreateNode(ctx, db.NodeCreate{
		Name:          "n1",
		APIAddress:    "api.local",
		APIPort:       2222,
		SecretKey:     "secret",
		PublicAddress: "a.example.com",
		GroupID:       &g.ID,
	})
	require.NoError(t, err)

	first, err := store.CreateInbound(ctx, db.InboundCreate{
		Tag:        "in-1",
		NodeID:     n.ID,
		Protocol:   "vless",
		ListenPort: 443,
		Settings:   json.RawMessage(`{}`),
	})
	require.NoError(t, err)

	_, err = store.CreateInbound(ctx, db.InboundCreate{
		Tag:        "in-2",
		NodeID:     n.ID,
		Protocol:   "trojan",
		ListenPort: 8443,
		Settings:   json.RawMessage(`{}`),
	})
	require.NoError(t, err)

	newTag := "in-2"
	_, err = store.UpdateInbound(ctx, first.ID, db.InboundUpdate{Tag: &newTag})
	require.Error(t, err)
	require.ErrorIs(t, err, db.ErrConflict)

	newPort := 9999
	_, err = store.UpdateInbound(ctx, 99999, db.InboundUpdate{ListenPort: &newPort})
	require.Error(t, err)
	require.ErrorIs(t, err, db.ErrNotFound)
}
