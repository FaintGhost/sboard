package stats

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing/common/buf"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
	"github.com/stretchr/testify/require"
)

type fakePacketConn struct{}

func (c *fakePacketConn) ReadPacket(_ *buf.Buffer) (M.Socksaddr, error) {
	return M.Socksaddr{}, errors.New("not implemented")
}

func (c *fakePacketConn) WritePacket(_ *buf.Buffer, _ M.Socksaddr) error {
	return nil
}

func (c *fakePacketConn) Close() error { return nil }

func (c *fakePacketConn) LocalAddr() net.Addr {
	return &net.UDPAddr{IP: net.IPv4zero, Port: 0}
}

func (c *fakePacketConn) SetDeadline(_ time.Time) error      { return nil }
func (c *fakePacketConn) SetReadDeadline(_ time.Time) error  { return nil }
func (c *fakePacketConn) SetWriteDeadline(_ time.Time) error { return nil }

var _ N.PacketConn = (*fakePacketConn)(nil)

func TestInboundTrafficTracker_NilSafety(t *testing.T) {
	var tracker *InboundTrafficTracker

	require.Nil(t, tracker.Snapshot(false))
	require.Equal(t, InboundTrafficMeta{}, tracker.Meta())
}

func TestInboundTrafficTracker_CounterSnapshotAndReset(t *testing.T) {
	tracker := NewInboundTrafficTracker()

	up1, down1 := tracker.counter("z-in", "bob")
	up1.Add(11)
	down1.Add(22)

	up2, down2 := tracker.counter("a-in", "alice")
	up2.Add(3)
	down2.Add(4)

	out := tracker.Snapshot(false)
	require.Len(t, out, 2)
	require.Equal(t, "a-in", out[0].Tag)
	require.Equal(t, "alice", out[0].User)
	require.Equal(t, int64(3), out[0].Uplink)
	require.Equal(t, int64(4), out[0].Downlink)

	require.Equal(t, "z-in", out[1].Tag)
	require.Equal(t, "bob", out[1].User)
	require.Equal(t, int64(11), out[1].Uplink)
	require.Equal(t, int64(22), out[1].Downlink)

	meta := tracker.Meta()
	require.Equal(t, 2, meta.TrackedTags)
	require.Equal(t, int64(0), meta.TCPConns)
	require.Equal(t, int64(0), meta.UDPConns)

	reset := tracker.Snapshot(true)
	require.Len(t, reset, 2)
	require.Equal(t, int64(3), reset[0].Uplink)
	require.Equal(t, int64(4), reset[0].Downlink)

	after := tracker.Snapshot(false)
	require.Len(t, after, 2)
	require.Equal(t, int64(0), after[0].Uplink)
	require.Equal(t, int64(0), after[0].Downlink)
}

func TestInboundTrafficTracker_RoutedConnectionAndPacketConnection(t *testing.T) {
	tracker := NewInboundTrafficTracker()

	left, right := net.Pipe()
	defer left.Close()
	defer right.Close()

	_ = left.SetDeadline(time.Now().Add(2 * time.Second))
	_ = right.SetDeadline(time.Now().Add(2 * time.Second))

	wrapped := tracker.RoutedConnection(
		context.Background(),
		left,
		adapter.InboundContext{Inbound: "", User: "  alice  "},
		nil,
		nil,
	)

	done := make(chan error, 1)
	go func() {
		buf := make([]byte, 16)
		n, err := right.Read(buf)
		if err != nil {
			done <- err
			return
		}
		_, err = right.Write(buf[:n])
		done <- err
	}()

	_, err := wrapped.Write([]byte("ping"))
	require.NoError(t, err)

	rb := make([]byte, 16)
	_, err = wrapped.Read(rb)
	require.NoError(t, err)
	require.NoError(t, <-done)

	udpWrapped := tracker.RoutedPacketConnection(
		context.Background(),
		&fakePacketConn{},
		adapter.InboundContext{Inbound: "udp-in", User: " user2 "},
		nil,
		nil,
	)
	require.NotNil(t, udpWrapped)

	rows := tracker.Snapshot(false)
	require.Len(t, rows, 2)

	require.Equal(t, "_unknown", rows[0].Tag)
	require.Equal(t, "alice", rows[0].User)
	require.Greater(t, rows[0].Uplink, int64(0))
	require.Greater(t, rows[0].Downlink, int64(0))

	require.Equal(t, "udp-in", rows[1].Tag)
	require.Equal(t, "user2", rows[1].User)
	require.Equal(t, int64(0), rows[1].Uplink)
	require.Equal(t, int64(0), rows[1].Downlink)

	meta := tracker.Meta()
	require.Equal(t, 2, meta.TrackedTags)
	require.Equal(t, int64(1), meta.TCPConns)
	require.Equal(t, int64(1), meta.UDPConns)
}
