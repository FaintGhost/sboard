package monitor

import (
	"context"
	"log"
	"time"

	"sboard/panel/internal/db"
	"sboard/panel/internal/node"
)

// TrafficMonitor periodically pulls inbound-level traffic deltas from nodes and stores them.
// It is intentionally isolated from sync/subscription flows.
type TrafficMonitor struct {
	store  *db.Store
	client *node.Client
}

func NewTrafficMonitor(store *db.Store, client *node.Client) *TrafficMonitor {
	if client == nil {
		client = node.NewClient(nil)
	}
	return &TrafficMonitor{store: store, client: client}
}

func (m *TrafficMonitor) SampleOnce(ctx context.Context) error {
	if m == nil || m.store == nil {
		return nil
	}
	nodes, err := m.store.ListNodes(ctx, 1000, 0)
	if err != nil {
		return err
	}
	for _, n := range nodes {
		// Skip nodes without API.
		if n.APIPort <= 0 || n.SecretKey == "" {
			continue
		}
		items, meta, err := m.client.InboundTrafficWithMeta(ctx, n, true)
		if err != nil {
			// Sampling should never block core functionality; keep it best-effort.
			log.Printf("[traffic] node id=%d name=%s pull failed: %v", n.ID, n.Name, err)
			continue
		}
		var sumUp, sumDown int64
		for _, it := range items {
			sumUp += it.Uplink
			sumDown += it.Downlink
		}
		// Log only when we have something actionable: either no tracked tags, or non-zero deltas.
		if len(items) == 0 || sumUp > 0 || sumDown > 0 {
			if meta != nil {
				log.Printf("[traffic] node id=%d name=%s inbounds=%d uplink=%d downlink=%d meta={tags:%d tcp:%d udp:%d}",
					n.ID, n.Name, len(items), sumUp, sumDown, meta.TrackedTags, meta.TCPConns, meta.UDPConns)
			} else {
				log.Printf("[traffic] node id=%d name=%s inbounds=%d uplink=%d downlink=%d", n.ID, n.Name, len(items), sumUp, sumDown)
			}
		}
		for _, it := range items {
			// In sing-box stats naming, uplink = client->server (read), downlink = server->client (write).
			if _, err := m.store.InsertInboundTrafficDelta(ctx, n.ID, it.Tag, it.Uplink, it.Downlink, it.At); err != nil {
				log.Printf("[traffic] node id=%d name=%s inbound=%s insert failed: %v", n.ID, n.Name, it.Tag, err)
				continue
			}
		}
	}
	return nil
}

func (m *TrafficMonitor) Run(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		return
	}
	// Initial pass for quick feedback.
	if err := m.SampleOnce(ctx); err != nil {
		log.Printf("[traffic] initial sample failed: %v", err)
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := m.SampleOnce(ctx); err != nil {
				log.Printf("[traffic] sample failed: %v", err)
			}
		}
	}
}
