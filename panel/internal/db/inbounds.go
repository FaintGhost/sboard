package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
)

type Inbound struct {
	ID                int64
	UUID              string
	Tag               string
	NodeID            int64
	Protocol          string
	ListenPort        int
	PublicPort        int
	Settings          json.RawMessage
	TLSSettings       json.RawMessage
	TransportSettings json.RawMessage
}

type InboundCreate struct {
	Tag               string
	NodeID            int64
	Protocol          string
	ListenPort        int
	PublicPort        int
	Settings          json.RawMessage
	TLSSettings       json.RawMessage
	TransportSettings json.RawMessage
}

type InboundUpdate struct {
	Tag               *string
	Protocol          *string
	ListenPort        *int
	PublicPort        *int
	Settings          *json.RawMessage
	TLSSettings       *json.RawMessage
	TransportSettings *json.RawMessage
}

func (s *Store) CreateInbound(ctx context.Context, req InboundCreate) (Inbound, error) {
	if _, err := s.GetNodeByID(ctx, req.NodeID); err != nil {
		return Inbound{}, err
	}
	id := uuid.NewString()
	res, err := s.DB.ExecContext(ctx, `
    INSERT INTO inbounds (uuid, tag, node_id, protocol, listen_port, public_port, settings, tls_settings, transport_settings)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
  `,
		id,
		req.Tag,
		req.NodeID,
		req.Protocol,
		req.ListenPort,
		nullInt(req.PublicPort),
		string(req.Settings),
		nullJSON(req.TLSSettings),
		nullJSON(req.TransportSettings),
	)
	if err != nil {
		if isConflict(err) {
			return Inbound{}, ErrConflict
		}
		return Inbound{}, err
	}
	rowID, err := res.LastInsertId()
	if err != nil {
		return Inbound{}, err
	}
	return s.GetInboundByID(ctx, rowID)
}

func (s *Store) GetInboundByID(ctx context.Context, id int64) (Inbound, error) {
	row := s.DB.QueryRowContext(ctx, `
    SELECT id, uuid, tag, node_id, protocol, listen_port, COALESCE(public_port, 0), settings, tls_settings, transport_settings
    FROM inbounds WHERE id = ?
  `, id)
	return scanInbound(row)
}

func (s *Store) ListInbounds(ctx context.Context, limit, offset int, nodeID int64) ([]Inbound, error) {
	q := `
    SELECT id, uuid, tag, node_id, protocol, listen_port, COALESCE(public_port, 0), settings, tls_settings, transport_settings
    FROM inbounds `
	args := []any{}
	if nodeID > 0 {
		q += "WHERE node_id = ? "
		args = append(args, nodeID)
	}
	q += "ORDER BY id DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := s.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []Inbound{}
	for rows.Next() {
		inb, err := scanInbound(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, inb)
	}
	return out, rows.Err()
}

func (s *Store) UpdateInbound(ctx context.Context, id int64, update InboundUpdate) (Inbound, error) {
	sets := []string{}
	args := []any{}

	if update.Tag != nil {
		sets = append(sets, "tag = ?")
		args = append(args, *update.Tag)
	}
	if update.Protocol != nil {
		sets = append(sets, "protocol = ?")
		args = append(args, *update.Protocol)
	}
	if update.ListenPort != nil {
		sets = append(sets, "listen_port = ?")
		args = append(args, *update.ListenPort)
	}
	if update.PublicPort != nil {
		sets = append(sets, "public_port = ?")
		args = append(args, nullInt(*update.PublicPort))
	}
	if update.Settings != nil {
		sets = append(sets, "settings = ?")
		args = append(args, string(*update.Settings))
	}
	if update.TLSSettings != nil {
		sets = append(sets, "tls_settings = ?")
		args = append(args, nullJSON(*update.TLSSettings))
	}
	if update.TransportSettings != nil {
		sets = append(sets, "transport_settings = ?")
		args = append(args, nullJSON(*update.TransportSettings))
	}

	if len(sets) == 0 {
		return s.GetInboundByID(ctx, id)
	}

	args = append(args, id)
	res, err := s.DB.ExecContext(ctx, "UPDATE inbounds SET "+stringsJoin(sets, ", ")+" WHERE id = ?", args...)
	if err != nil {
		if isConflict(err) {
			return Inbound{}, ErrConflict
		}
		return Inbound{}, err
	}
	n, err := res.RowsAffected()
	if err == nil && n == 0 {
		return Inbound{}, ErrNotFound
	}
	if err != nil {
		return Inbound{}, err
	}
	return s.GetInboundByID(ctx, id)
}

func (s *Store) DeleteInbound(ctx context.Context, id int64) error {
	res, err := s.DB.ExecContext(ctx, "DELETE FROM inbounds WHERE id = ?", id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err == nil && n == 0 {
		return ErrNotFound
	}
	return err
}

func (s *Store) DeleteInboundsByNode(ctx context.Context, nodeID int64) (int64, error) {
	res, err := s.DB.ExecContext(ctx, "DELETE FROM inbounds WHERE node_id = ?", nodeID)
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return n, nil
}

type inboundRowScanner interface {
	Scan(dest ...any) error
}

func scanInbound(row inboundRowScanner) (Inbound, error) {
	var inb Inbound
	var settings string
	var tls sql.NullString
	var transport sql.NullString
	if err := row.Scan(
		&inb.ID,
		&inb.UUID,
		&inb.Tag,
		&inb.NodeID,
		&inb.Protocol,
		&inb.ListenPort,
		&inb.PublicPort,
		&settings,
		&tls,
		&transport,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Inbound{}, ErrNotFound
		}
		return Inbound{}, err
	}
	inb.Settings = json.RawMessage(settings)
	if tls.Valid {
		inb.TLSSettings = json.RawMessage(tls.String)
	}
	if transport.Valid {
		inb.TransportSettings = json.RawMessage(transport.String)
	}
	return inb, nil
}

func nullInt(v int) sql.NullInt64 {
	if v <= 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(v), Valid: true}
}

func nullJSON(v json.RawMessage) sql.NullString {
	if len(v) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{String: string(v), Valid: true}
}
