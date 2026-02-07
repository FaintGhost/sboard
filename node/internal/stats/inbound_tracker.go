package stats

import (
	"context"
	"net"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sagernet/sing-box/adapter"
	sbufio "github.com/sagernet/sing/common/bufio"
	N "github.com/sagernet/sing/common/network"
)

type InboundTraffic struct {
	Tag      string    `json:"tag"`
	Uplink   int64     `json:"uplink"`
	Downlink int64     `json:"downlink"`
	At       time.Time `json:"at"`
}

type InboundTrafficMeta struct {
	TrackedTags int   `json:"tracked_tags"`
	TCPConns    int64 `json:"tcp_conns"`
	UDPConns    int64 `json:"udp_conns"`
}

// InboundTrafficTracker tracks inbound-level traffic by wrapping routed connections.
//
// Notes:
//   - This only counts traffic that passes through the sing-box router.
//   - It tracks by metadata.Inbound (inbound tag).
//   - Snapshot(reset=true) returns delta since last reset and resets internal counters.
type InboundTrafficTracker struct {
	createdAt time.Time
	mu        sync.Mutex
	uplink    map[string]*atomic.Int64
	downlink  map[string]*atomic.Int64

	tcpConns atomic.Int64
	udpConns atomic.Int64
}

func NewInboundTrafficTracker() *InboundTrafficTracker {
	return &InboundTrafficTracker{
		createdAt: time.Now().UTC(),
		uplink:    make(map[string]*atomic.Int64),
		downlink:  make(map[string]*atomic.Int64),
	}
}

func (t *InboundTrafficTracker) RoutedConnection(ctx context.Context, conn net.Conn, metadata adapter.InboundContext, _ adapter.Rule, _ adapter.Outbound) net.Conn {
	t.tcpConns.Add(1)
	tag := metadata.Inbound
	if tag == "" {
		// Keep an explicit bucket for debugging; metadata.Inbound should normally be set by inbound implementations.
		tag = "_unknown"
	}
	up, down := t.counter(tag)
	return sbufio.NewInt64CounterConn(conn, []*atomic.Int64{up}, []*atomic.Int64{down})
}

func (t *InboundTrafficTracker) RoutedPacketConnection(ctx context.Context, conn N.PacketConn, metadata adapter.InboundContext, _ adapter.Rule, _ adapter.Outbound) N.PacketConn {
	t.udpConns.Add(1)
	tag := metadata.Inbound
	if tag == "" {
		tag = "_unknown"
	}
	up, down := t.counter(tag)
	return sbufio.NewInt64CounterPacketConn(conn, []*atomic.Int64{up}, nil, []*atomic.Int64{down}, nil)
}

func (t *InboundTrafficTracker) counter(tag string) (uplink *atomic.Int64, downlink *atomic.Int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	up := t.uplink[tag]
	if up == nil {
		up = &atomic.Int64{}
		t.uplink[tag] = up
	}
	down := t.downlink[tag]
	if down == nil {
		down = &atomic.Int64{}
		t.downlink[tag] = down
	}
	return up, down
}

func (t *InboundTrafficTracker) Snapshot(reset bool) []InboundTraffic {
	if t == nil {
		return nil
	}
	now := time.Now().UTC()
	t.mu.Lock()
	tags := make([]string, 0, len(t.uplink))
	for tag := range t.uplink {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	out := make([]InboundTraffic, 0, len(tags))
	for _, tag := range tags {
		up := t.uplink[tag]
		down := t.downlink[tag]
		if up == nil || down == nil {
			continue
		}
		var u, d int64
		if reset {
			u = up.Swap(0)
			d = down.Swap(0)
		} else {
			u = up.Load()
			d = down.Load()
		}
		out = append(out, InboundTraffic{
			Tag:      tag,
			Uplink:   u,
			Downlink: d,
			At:       now,
		})
	}
	t.mu.Unlock()
	return out
}

func (t *InboundTrafficTracker) Meta() InboundTrafficMeta {
	if t == nil {
		return InboundTrafficMeta{}
	}
	t.mu.Lock()
	tracked := len(t.uplink)
	t.mu.Unlock()
	return InboundTrafficMeta{
		TrackedTags: tracked,
		TCPConns:    t.tcpConns.Load(),
		UDPConns:    t.udpConns.Load(),
	}
}
