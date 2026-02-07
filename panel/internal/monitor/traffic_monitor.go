package monitor

import (
	"context"
	"log"
	"time"

	"sboard/panel/internal/db"
	"sboard/panel/internal/node"
)

// TrafficMonitor periodically samples node traffic (host interface counters) and stores it.
// It is intentionally isolated from sync/subscription flows so it can be disabled without risk.
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
		sample, err := m.client.Traffic(ctx, n)
		if err != nil {
			// Sampling should never block core functionality; keep it best-effort.
			log.Printf("[traffic] node id=%d name=%s sample failed: %v", n.ID, n.Name, err)
			continue
		}
		_, err = m.store.InsertNodeTrafficSample(ctx, n.ID, int64(sample.TxBytes), int64(sample.RxBytes), sample.At)
		if err != nil {
			log.Printf("[traffic] node id=%d name=%s insert failed: %v", n.ID, n.Name, err)
			continue
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
