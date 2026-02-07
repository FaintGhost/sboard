package db

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type TrafficStat struct {
	ID         int64
	UserID     *int64
	NodeID     *int64
	Upload     int64
	Download   int64
	RecordedAt time.Time
}

func (s *Store) InsertNodeTrafficSample(ctx context.Context, nodeID int64, uploadBytes, downloadBytes int64, recordedAt time.Time) (TrafficStat, error) {
	if nodeID <= 0 {
		return TrafficStat{}, errors.New("invalid node_id")
	}
	if recordedAt.IsZero() {
		recordedAt = s.nowUTC()
	}
	res, err := s.DB.ExecContext(
		ctx,
		`
    INSERT INTO traffic_stats (user_id, node_id, upload, download, recorded_at)
    VALUES (NULL, ?, ?, ?, ?)
  `,
		nodeID,
		uploadBytes,
		downloadBytes,
		recordedAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return TrafficStat{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return TrafficStat{}, err
	}
	return s.GetTrafficStatByID(ctx, id)
}

func (s *Store) GetTrafficStatByID(ctx context.Context, id int64) (TrafficStat, error) {
	row := s.DB.QueryRowContext(ctx, `
    SELECT id, user_id, node_id, upload, download, recorded_at
    FROM traffic_stats WHERE id = ?
  `, id)
	return scanTrafficStat(row)
}

func (s *Store) ListNodeTrafficSamples(ctx context.Context, nodeID int64, limit, offset int) ([]TrafficStat, error) {
	if nodeID <= 0 {
		return nil, errors.New("invalid node_id")
	}
	rows, err := s.DB.QueryContext(ctx, `
    SELECT id, user_id, node_id, upload, download, recorded_at
    FROM traffic_stats
    WHERE node_id = ? AND user_id IS NULL
    ORDER BY recorded_at DESC, id DESC
    LIMIT ? OFFSET ?
  `, nodeID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []TrafficStat{}
	for rows.Next() {
		st, err := scanTrafficStat(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, st)
	}
	return out, rows.Err()
}

type trafficStatRowScanner interface {
	Scan(dest ...any) error
}

func scanTrafficStat(row trafficStatRowScanner) (TrafficStat, error) {
	var st TrafficStat
	var userID sql.NullInt64
	var nodeID sql.NullInt64
	var recorded sql.NullString
	if err := row.Scan(&st.ID, &userID, &nodeID, &st.Upload, &st.Download, &recorded); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return TrafficStat{}, ErrNotFound
		}
		return TrafficStat{}, err
	}
	if userID.Valid {
		v := userID.Int64
		st.UserID = &v
	}
	if nodeID.Valid {
		v := nodeID.Int64
		st.NodeID = &v
	}
	if recorded.Valid {
		if t, err := parseSQLiteTime(recorded.String); err == nil {
			st.RecordedAt = t
		}
	}
	return st, nil
}
