package traffic

import (
	"context"
	"errors"
	"time"

	"sboard/panel/internal/db"
)

// Provider is a pluggable traffic analytics backend.
//
// Today it is backed by SQLite via db.Store, but keeping this interface makes it
// straightforward to move aggregation to another store without touching API/UI.
type Provider interface {
	NodesSummary(ctx context.Context, window time.Duration) ([]NodeSummary, error)
	TotalSummary(ctx context.Context, window time.Duration) (TotalSummary, error)
	Timeseries(ctx context.Context, q TimeseriesQuery) ([]TimeseriesPoint, error)
}

type NodeSummary struct {
	NodeID         int64
	Upload         int64
	Download       int64
	LastRecordedAt time.Time
	Samples        int64
	Inbounds       int64
}

type TotalSummary struct {
	Upload         int64
	Download       int64
	LastRecordedAt time.Time
	Samples        int64
	Nodes          int64
	Inbounds       int64
}

type TimeseriesQuery struct {
	Window time.Duration
	Bucket Bucket
	// NodeID == 0 means global.
	NodeID int64
}

type Bucket string

const (
	BucketMinute Bucket = "minute"
	BucketHour   Bucket = "hour"
	BucketDay    Bucket = "day"
)

type TimeseriesPoint struct {
	BucketStart time.Time
	Upload      int64
	Download    int64
}

func NewSQLiteProvider(store *db.Store) Provider {
	return &sqliteProvider{store: store}
}

type sqliteProvider struct {
	store *db.Store
}

func (p *sqliteProvider) NodesSummary(ctx context.Context, window time.Duration) ([]NodeSummary, error) {
	if p == nil || p.store == nil {
		return nil, errors.New("missing store")
	}
	if window < 0 {
		return nil, errors.New("invalid window")
	}

	var since time.Time
	if window > 0 {
		since = p.store.NowUTC().Add(-window)
	}

	items, err := p.store.ListNodeTrafficSummaries(ctx, since)
	if err != nil {
		return nil, err
	}
	out := make([]NodeSummary, 0, len(items))
	for _, it := range items {
		out = append(out, NodeSummary{
			NodeID:         it.NodeID,
			Upload:         it.Upload,
			Download:       it.Download,
			LastRecordedAt: it.LastRecordedAt,
			Samples:        it.Samples,
			Inbounds:       it.Inbounds,
		})
	}
	return out, nil
}

func (p *sqliteProvider) TotalSummary(ctx context.Context, window time.Duration) (TotalSummary, error) {
	if p == nil || p.store == nil {
		return TotalSummary{}, errors.New("missing store")
	}
	if window < 0 {
		return TotalSummary{}, errors.New("invalid window")
	}

	var since time.Time
	if window > 0 {
		since = p.store.NowUTC().Add(-window)
	}

	it, err := p.store.GetTrafficTotalSummary(ctx, since)
	if err != nil {
		return TotalSummary{}, err
	}
	return TotalSummary{
		Upload:         it.Upload,
		Download:       it.Download,
		LastRecordedAt: it.LastRecordedAt,
		Samples:        it.Samples,
		Nodes:          it.Nodes,
		Inbounds:       it.Inbounds,
	}, nil
}

func (p *sqliteProvider) Timeseries(ctx context.Context, q TimeseriesQuery) ([]TimeseriesPoint, error) {
	if p == nil || p.store == nil {
		return nil, errors.New("missing store")
	}
	if q.Window < 0 {
		return nil, errors.New("invalid window")
	}
	if q.Bucket != BucketMinute && q.Bucket != BucketHour && q.Bucket != BucketDay {
		return nil, errors.New("invalid bucket")
	}

	var since time.Time
	if q.Window > 0 {
		since = p.store.NowUTC().Add(-q.Window)
	}

	items, err := p.store.ListTrafficTimeseries(ctx, since, q.NodeID, string(q.Bucket))
	if err != nil {
		return nil, err
	}

	out := make([]TimeseriesPoint, 0, len(items))
	for _, it := range items {
		out = append(out, TimeseriesPoint{
			BucketStart: it.BucketStart,
			Upload:      it.Upload,
			Download:    it.Download,
		})
	}
	return out, nil
}
