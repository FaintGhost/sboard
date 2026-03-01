package rpc

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/require"
	nodev1 "sboard/node/internal/rpc/gen/sboard/node/v1"
	"sboard/node/internal/stats"
	syncsvc "sboard/node/internal/sync"
)

type syncApplierStub struct {
	err  error
	body []byte
}

func (s *syncApplierStub) ApplySyncPayload(ctx context.Context, body []byte) error {
	s.body = append([]byte(nil), body...)
	return s.err
}

type inboundProviderStub struct{}

func (inboundProviderStub) InboundTraffic(reset bool) []stats.InboundTraffic {
	return []stats.InboundTraffic{{Tag: "ss-in", User: "alice", Uplink: 1, Downlink: 2}}
}

func (inboundProviderStub) InboundTrafficMeta() stats.InboundTrafficMeta {
	return stats.InboundTrafficMeta{TrackedTags: 1, TCPConns: 2, UDPConns: 3}
}

func TestNodeControlSyncConfigSuccess(t *testing.T) {
	applier := &syncApplierStub{}
	s := &Server{applier: applier}

	resp, err := s.SyncConfig(context.Background(), &nodev1.SyncConfigRequest{PayloadJson: []byte(`{"inbounds":[]}`)})
	require.NoError(t, err)
	require.Equal(t, "ok", resp.GetStatus())
	require.JSONEq(t, `{"inbounds":[]}`, string(applier.body))
}

func TestNodeControlSyncConfigBadRequest(t *testing.T) {
	applier := &syncApplierStub{err: syncsvc.BadRequestError{Message: "invalid config"}}
	s := &Server{applier: applier}

	_, err := s.SyncConfig(context.Background(), &nodev1.SyncConfigRequest{PayloadJson: []byte(`{"inbounds":[]}`)})
	require.Error(t, err)

	var connErr *connect.Error
	require.True(t, errors.As(err, &connErr))
	require.Equal(t, connect.CodeInvalidArgument, connErr.Code())
}

func TestNodeControlInvalidArgument(t *testing.T) {
	s := &Server{applier: &syncApplierStub{}}

	_, err := s.SyncConfig(context.Background(), &nodev1.SyncConfigRequest{})
	require.Error(t, err)

	var connErr *connect.Error
	require.True(t, errors.As(err, &connErr))
	require.Equal(t, connect.CodeInvalidArgument, connErr.Code())
}

func TestNodeControlHealth(t *testing.T) {
	s := &Server{}
	resp, err := s.Health(context.Background(), &nodev1.HealthRequest{})
	require.NoError(t, err)
	require.Equal(t, "ok", resp.GetStatus())
}
