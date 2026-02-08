package stats

import (
	"context"
	"net"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sagernet/sing-box/adapter"
	sbufio "github.com/sagernet/sing/common/bufio"
	N "github.com/sagernet/sing/common/network"
)

type InboundTraffic struct {
	Tag      string    `json:"tag"`
	User     string    `json:"user,omitempty"`
	Uplink   int64     `json:"uplink"`
	Downlink int64     `json:"downlink"`
	At       time.Time `json:"at"`
}

type InboundTrafficMeta struct {
	TrackedTags int   `json:"tracked_tags"`
	TCPConns    int64 `json:"tcp_conns"`
	UDPConns    int64 `json:"udp_conns"`
}

type inboundUserKey struct {
	Tag  string
	User string
}

// InboundTrafficTracker tracks inbound-level traffic by wrapping routed connections.
//
// Notes:
//   - This only counts traffic that passes through the sing-box router.
//   - It tracks by metadata.Inbound + metadata.User.
//   - Snapshot(reset=true) returns delta since last reset and resets internal counters.
type InboundTrafficTracker struct {
	createdAt time.Time
	mu        sync.Mutex
	uplink    map[inboundUserKey]*atomic.Int64
	downlink  map[inboundUserKey]*atomic.Int64

	tcpConns atomic.Int64
	udpConns atomic.Int64
}

func NewInboundTrafficTracker() *InboundTrafficTracker {
	return &InboundTrafficTracker{
		createdAt: time.Now().UTC(),
		uplink:    make(map[inboundUserKey]*atomic.Int64),
		downlink:  make(map[inboundUserKey]*atomic.Int64),
	}
}

func (t *InboundTrafficTracker) RoutedConnection(ctx context.Context, conn net.Conn, metadata adapter.InboundContext, _ adapter.Rule, _ adapter.Outbound) net.Conn {
	t.tcpConns.Add(1)
	tag := metadata.Inbound
	if tag == "" {
		// Keep an explicit bucket for debugging; metadata.Inbound should normally be set by inbound implementations.
		tag = "_unknown"
	}
	user := strings.TrimSpace(metadata.User)
	up, down := t.counter(tag, user)
	return sbufio.NewInt64CounterConn(conn, []*atomic.Int64{up}, []*atomic.Int64{down})
}

func (t *InboundTrafficTracker) RoutedPacketConnection(ctx context.Context, conn N.PacketConn, metadata adapter.InboundContext, _ adapter.Rule, _ adapter.Outbound) N.PacketConn {
	t.udpConns.Add(1)
	tag := metadata.Inbound
	if tag == "" {
		tag = "_unknown"
	}
	user := strings.TrimSpace(metadata.User)
	up, down := t.counter(tag, user)
	return sbufio.NewInt64CounterPacketConn(conn, []*atomic.Int64{up}, nil, []*atomic.Int64{down}, nil)
}

func (t *InboundTrafficTracker) counter(tag, user string) (uplink *atomic.Int64, downlink *atomic.Int64) {
	key := inboundUserKey{Tag: tag, User: user}

	t.mu.Lock()
	defer t.mu.Unlock()

	up := t.uplink[key]
	if up == nil {
		up = &atomic.Int64{}
		t.uplink[key] = up
	}

	down := t.downlink[key]
	if down == nil {
		down = &atomic.Int64{}
		t.downlink[key] = down
	}

	return up, down
}

func (t *InboundTrafficTracker) Snapshot(reset bool) []InboundTraffic {
	if t == nil {
		return nil
	}

	now := time.Now().UTC()

	t.mu.Lock()
	keys := make([]inboundUserKey, 0, len(t.uplink))
	for key := range t.uplink {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].Tag == keys[j].Tag {
			return keys[i].User < keys[j].User
		}
		return keys[i].Tag < keys[j].Tag
	})

	out := make([]InboundTraffic, 0, len(keys))
	for _, key := range keys {
		up := t.uplink[key]
		down := t.downlink[key]
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
			Tag:      key.Tag,
			User:     key.User,
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
	tags := make(map[string]struct{}, len(t.uplink))
	for key := range t.uplink {
		tags[key.Tag] = struct{}{}
	}
	tracked := len(tags)
	t.mu.Unlock()

	return InboundTrafficMeta{
		TrackedTags: tracked,
		TCPConns:    t.tcpConns.Load(),
		UDPConns:    t.udpConns.Load(),
	}
}
