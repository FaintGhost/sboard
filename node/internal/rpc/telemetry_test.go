package rpc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	nodev1 "sboard/node/internal/rpc/gen/sboard/node/v1"
	"sboard/node/internal/stats"
)

func TestNodeControlTelemetry(t *testing.T) {
	t.Run("traffic sample", func(t *testing.T) {
		iface, err := stats.DetectDefaultInterface()
		if err != nil {
			t.Skipf("skip traffic sample test: %v", err)
		}

		s := &Server{}
		resp, err := s.GetTraffic(context.Background(), &nodev1.GetTrafficRequest{Interface: iface})
		require.NoError(t, err)
		require.Equal(t, iface, resp.GetInterface())
		require.NotEmpty(t, resp.GetAt())
		parsedAt, err := time.Parse(time.RFC3339, resp.GetAt())
		require.NoError(t, err)
		require.False(t, parsedAt.IsZero())
	})

	t.Run("inbound traffic reset and meta", func(t *testing.T) {
		s := &Server{inbound: inboundProviderStub{}}
		resp, err := s.GetInboundTraffic(context.Background(), &nodev1.GetInboundTrafficRequest{Reset_: true})
		require.NoError(t, err)
		require.True(t, resp.GetReset_())
		require.Len(t, resp.GetData(), 1)
		require.Equal(t, "ss-in", resp.GetData()[0].GetTag())
		require.Equal(t, "alice", resp.GetData()[0].GetUser())
		require.Equal(t, int64(1), resp.GetData()[0].GetUplink())
		require.Equal(t, int64(2), resp.GetData()[0].GetDownlink())
		require.NotNil(t, resp.GetMeta())
		require.Equal(t, int32(1), resp.GetMeta().GetTrackedTags())
	})
}
