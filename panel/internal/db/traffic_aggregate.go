package db

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type NodeTrafficSummary struct {
	NodeID         int64
	Upload         int64
	Download       int64
	LastRecordedAt time.Time
	Samples        int64
	Inbounds       int64
}

type TrafficTotalSummary struct {
	Upload         int64
	Download       int64
	LastRecordedAt time.Time
	Samples        int64
	Nodes          int64
	Inbounds       int64
}

type TrafficTimeseriesPoint struct {
	BucketStart time.Time
	Upload      int64
	Download    int64
}

func (s *Store) ListNodeTrafficSummaries(ctx context.Context, since time.Time) ([]NodeTrafficSummary, error) {
	if s == nil || s.DB == nil {
		return nil, errors.New("missing db")
	}

	q := `
    SELECT
      n.id AS node_id,
      COALESCE(SUM(ts.upload), 0) AS upload,
      COALESCE(SUM(ts.download), 0) AS download,
      COALESCE(MAX(ts.recorded_at), '') AS last_recorded_at,
      COUNT(ts.id) AS samples,
      COUNT(DISTINCT ts.inbound_tag) AS inbounds
    FROM nodes n
    LEFT JOIN traffic_stats ts
      ON ts.node_id = n.id
     AND ts.user_id IS NULL
  `
	args := []any{}
	if !since.IsZero() {
		q += " AND ts.recorded_at >= ?\n"
		args = append(args, since.UTC().Format(time.RFC3339))
	}
	q += `
    GROUP BY n.id
    ORDER BY n.id ASC
  `

	rows, err := s.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []NodeTrafficSummary{}
	for rows.Next() {
		var it NodeTrafficSummary
		var last sql.NullString
		if err := rows.Scan(&it.NodeID, &it.Upload, &it.Download, &last, &it.Samples, &it.Inbounds); err != nil {
			return nil, err
		}
		if last.Valid && last.String != "" {
			if t, err := parseSQLiteTime(last.String); err == nil {
				it.LastRecordedAt = t
			}
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (s *Store) GetTrafficTotalSummary(ctx context.Context, since time.Time) (TrafficTotalSummary, error) {
	if s == nil || s.DB == nil {
		return TrafficTotalSummary{}, errors.New("missing db")
	}

	q := `
    SELECT
      COALESCE(SUM(upload), 0) AS upload,
      COALESCE(SUM(download), 0) AS download,
      COALESCE(MAX(recorded_at), '') AS last_recorded_at,
      COUNT(id) AS samples,
      COUNT(DISTINCT node_id) AS nodes,
      COUNT(DISTINCT inbound_tag) AS inbounds
    FROM traffic_stats
    WHERE user_id IS NULL
  `
	args := []any{}
	if !since.IsZero() {
		q += " AND recorded_at >= ?"
		args = append(args, since.UTC().Format(time.RFC3339))
	}

	row := s.DB.QueryRowContext(ctx, q, args...)

	var it TrafficTotalSummary
	var last sql.NullString
	if err := row.Scan(&it.Upload, &it.Download, &last, &it.Samples, &it.Nodes, &it.Inbounds); err != nil {
		return TrafficTotalSummary{}, err
	}
	if last.Valid && last.String != "" {
		if t, err := parseSQLiteTime(last.String); err == nil {
			it.LastRecordedAt = t
		}
	}

	return it, nil
}

// bucket can be "minute" | "hour" | "day".
func (s *Store) ListTrafficTimeseries(ctx context.Context, since time.Time, nodeID int64, bucket string) ([]TrafficTimeseriesPoint, error) {
	if s == nil || s.DB == nil {
		return nil, errors.New("missing db")
	}

	var fmt string
	switch bucket {
	case "minute":
		fmt = "%Y-%m-%dT%H:%M:00Z"
	case "hour":
		fmt = "%Y-%m-%dT%H:00:00Z"
	case "day":
		fmt = "%Y-%m-%dT00:00:00Z"
	default:
		return nil, errors.New("invalid bucket")
	}

	q := `
    SELECT
      strftime(?, recorded_at) AS bucket_start,
      COALESCE(SUM(upload), 0) AS upload,
      COALESCE(SUM(download), 0) AS download
    FROM traffic_stats
    WHERE user_id IS NULL
  `
	args := []any{fmt}
	if !since.IsZero() {
		q += " AND recorded_at >= ?"
		args = append(args, since.UTC().Format(time.RFC3339))
	}
	if nodeID > 0 {
		q += " AND node_id = ?"
		args = append(args, nodeID)
	}
	q += `
    GROUP BY bucket_start
    ORDER BY bucket_start ASC
  `

	rows, err := s.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []TrafficTimeseriesPoint{}
	for rows.Next() {
		var bucketStart string
		var it TrafficTimeseriesPoint
		if err := rows.Scan(&bucketStart, &it.Upload, &it.Download); err != nil {
			return nil, err
		}
		if t, err := parseSQLiteTime(bucketStart); err == nil {
			it.BucketStart = t
		}
		out = append(out, it)
	}

	return out, rows.Err()
}
