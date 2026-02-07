package db

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Node struct {
	ID            int64
	UUID          string
	Name          string
	APIAddress    string
	APIPort       int
	SecretKey     string
	PublicAddress string
	GroupID       *int64
	Status        string
	LastSeenAt    *time.Time
}

type NodeCreate struct {
	Name          string
	APIAddress    string
	APIPort       int
	SecretKey     string
	PublicAddress string
	GroupID       *int64
}

type NodeUpdate struct {
	Name          *string
	APIAddress    *string
	APIPort       *int
	SecretKey     *string
	PublicAddress *string
	GroupID       *int64
	GroupIDSet    bool
}

func (s *Store) CreateNode(ctx context.Context, req NodeCreate) (Node, error) {
	id := uuid.NewString()
	res, err := s.DB.ExecContext(ctx, `
    INSERT INTO nodes (uuid, name, address, port, secret_key, api_address, api_port, public_address, group_id)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
  `,
		id,
		req.Name,
		req.APIAddress,
		req.APIPort,
		req.SecretKey,
		req.APIAddress,
		req.APIPort,
		req.PublicAddress,
		nullInt64(req.GroupID),
	)
	if err != nil {
		if isConflict(err) {
			return Node{}, ErrConflict
		}
		return Node{}, err
	}
	rowID, err := res.LastInsertId()
	if err != nil {
		return Node{}, err
	}
	return s.GetNodeByID(ctx, rowID)
}

func (s *Store) GetNodeByID(ctx context.Context, id int64) (Node, error) {
	row := s.DB.QueryRowContext(ctx, `
    SELECT id, uuid, name, COALESCE(api_address, address), COALESCE(api_port, port), secret_key, COALESCE(public_address, ''), group_id, status, last_seen_at
    FROM nodes WHERE id = ?
  `, id)
	n, err := scanNode(row)
	if err != nil {
		return Node{}, err
	}
	return n, nil
}

func (s *Store) ListNodes(ctx context.Context, limit, offset int) ([]Node, error) {
	rows, err := s.DB.QueryContext(ctx, `
    SELECT id, uuid, name, COALESCE(api_address, address), COALESCE(api_port, port), secret_key, COALESCE(public_address, ''), group_id, status, last_seen_at
    FROM nodes ORDER BY id DESC LIMIT ? OFFSET ?
  `, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []Node{}
	for rows.Next() {
		n, err := scanNode(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func (s *Store) UpdateNode(ctx context.Context, id int64, update NodeUpdate) (Node, error) {
	sets := []string{}
	args := []any{}

	if update.Name != nil {
		sets = append(sets, "name = ?")
		args = append(args, *update.Name)
	}
	if update.APIAddress != nil {
		// Keep legacy address column in sync.
		sets = append(sets, "api_address = ?", "address = ?")
		args = append(args, *update.APIAddress, *update.APIAddress)
	}
	if update.APIPort != nil {
		sets = append(sets, "api_port = ?", "port = ?")
		args = append(args, *update.APIPort, *update.APIPort)
	}
	if update.SecretKey != nil {
		sets = append(sets, "secret_key = ?")
		args = append(args, *update.SecretKey)
	}
	if update.PublicAddress != nil {
		sets = append(sets, "public_address = ?")
		args = append(args, *update.PublicAddress)
	}
	if update.GroupIDSet {
		sets = append(sets, "group_id = ?")
		args = append(args, nullInt64(update.GroupID))
	}

	if len(sets) == 0 {
		return s.GetNodeByID(ctx, id)
	}

	args = append(args, id)
	res, err := s.DB.ExecContext(ctx, "UPDATE nodes SET "+stringsJoin(sets, ", ")+" WHERE id = ?", args...)
	if err != nil {
		if isConflict(err) {
			return Node{}, ErrConflict
		}
		return Node{}, err
	}
	n, err := res.RowsAffected()
	if err == nil && n == 0 {
		return Node{}, ErrNotFound
	}
	if err != nil {
		return Node{}, err
	}
	return s.GetNodeByID(ctx, id)
}

func (s *Store) DeleteNode(ctx context.Context, id int64) error {
	row := s.DB.QueryRowContext(ctx, "SELECT COUNT(1) FROM inbounds WHERE node_id = ?", id)
	var cnt int
	if err := row.Scan(&cnt); err != nil {
		return err
	}
	if cnt > 0 {
		return ErrConflict
	}
	res, err := s.DB.ExecContext(ctx, "DELETE FROM nodes WHERE id = ?", id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err == nil && n == 0 {
		return ErrNotFound
	}
	return err
}

func (s *Store) MarkNodeOnline(ctx context.Context, id int64, seenAt time.Time) error {
	if id <= 0 {
		return errors.New("invalid id")
	}
	if seenAt.IsZero() {
		seenAt = s.nowUTC()
	}
	res, err := s.DB.ExecContext(
		ctx,
		"UPDATE nodes SET status = 'online', last_seen_at = ? WHERE id = ?",
		seenAt.UTC().Format(time.RFC3339),
		id,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err == nil && n == 0 {
		return ErrNotFound
	}
	return err
}

func (s *Store) MarkNodeOffline(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("invalid id")
	}
	res, err := s.DB.ExecContext(ctx, "UPDATE nodes SET status = 'offline' WHERE id = ?", id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err == nil && n == 0 {
		return ErrNotFound
	}
	return err
}

type nodeRowScanner interface {
	Scan(dest ...any) error
}

func scanNode(row nodeRowScanner) (Node, error) {
	var n Node
	var groupID sql.NullInt64
	var lastSeen sql.NullString
	if err := row.Scan(
		&n.ID,
		&n.UUID,
		&n.Name,
		&n.APIAddress,
		&n.APIPort,
		&n.SecretKey,
		&n.PublicAddress,
		&groupID,
		&n.Status,
		&lastSeen,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Node{}, ErrNotFound
		}
		return Node{}, err
	}
	if groupID.Valid {
		v := groupID.Int64
		n.GroupID = &v
	}
	if lastSeen.Valid {
		if t, err := parseSQLiteTime(lastSeen.String); err == nil {
			n.LastSeenAt = &t
		}
	}
	return n, nil
}

func nullInt64(v *int64) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *v, Valid: true}
}
