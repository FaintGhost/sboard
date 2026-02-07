package monitor

import (
	"context"
	"log"
	"time"

	"sboard/panel/internal/db"
	"sboard/panel/internal/node"
)

type NodeClient interface {
	Health(ctx context.Context, node db.Node) error
	SyncConfig(ctx context.Context, node db.Node, payload any) error
}

type NodesMonitor struct {
	store  *db.Store
	client NodeClient

	// consecutive failures per node id
	fails map[int64]int

	offlineAfter int
}

func NewNodesMonitor(store *db.Store, client NodeClient) *NodesMonitor {
	if client == nil {
		client = node.NewClient(nil)
	}
	return &NodesMonitor{
		store:        store,
		client:       client,
		fails:        map[int64]int{},
		offlineAfter: 2,
	}
}

// CheckOnce updates node online/offline status by probing each node's /api/health.
// When a node transitions from offline -> online, it triggers a sync.
func (m *NodesMonitor) CheckOnce(ctx context.Context) error {
	if m == nil || m.store == nil || m.client == nil {
		return nil
	}
	nodes, err := m.store.ListNodes(ctx, 10000, 0)
	if err != nil {
		return err
	}

	for _, n := range nodes {
		// Keep a short per-node timeout so one bad node doesn't block the whole pass.
		nctx, cancel := context.WithTimeout(ctx, 3*time.Second)
		err := m.client.Health(nctx, n)
		cancel()

		if err == nil {
			m.fails[n.ID] = 0
			// Transition detection based on DB status.
			wasOnline := n.Status == "online"
			now := m.store.Now().UTC()
			if err := m.store.MarkNodeOnline(ctx, n.ID, now); err != nil {
				log.Printf("[monitor] mark online failed node=%d err=%v", n.ID, err)
			}
			if !wasOnline {
				if n.GroupID == nil {
					continue
				}
				if err := m.syncNode(ctx, n); err != nil {
					log.Printf("[monitor] sync failed node=%d err=%v", n.ID, err)
				}
			}
			continue
		}

		m.fails[n.ID]++
		if m.fails[n.ID] < m.offlineAfter {
			continue
		}
		if n.Status != "offline" {
			if err := m.store.MarkNodeOffline(ctx, n.ID); err != nil {
				log.Printf("[monitor] mark offline failed node=%d err=%v", n.ID, err)
			}
		}
	}
	return nil
}

func (m *NodesMonitor) syncNode(ctx context.Context, n db.Node) error {
	if n.GroupID == nil {
		return nil
	}
	inbounds, err := m.store.ListInbounds(ctx, 10000, 0, n.ID)
	if err != nil {
		return err
	}
	users, err := m.store.ListActiveUsersForGroup(ctx, *n.GroupID)
	if err != nil {
		return err
	}
	payload, err := node.BuildSyncPayload(n, inbounds, users)
	if err != nil {
		return err
	}
	return m.client.SyncConfig(ctx, n, payload)
}
