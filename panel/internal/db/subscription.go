package db

import (
  "context"
  "database/sql"
)

type SubscriptionInbound struct {
  NodePublicAddress string
  InboundUUID       string
  InboundType       string
  InboundTag        string
  InboundListenPort int
  InboundPublicPort int
  Settings          []byte
  TLSSettings       []byte
  TransportSettings []byte
}

func (s *Store) ListUserInbounds(ctx context.Context, userID int64) ([]SubscriptionInbound, error) {
  rows, err := s.DB.QueryContext(ctx, `
    SELECT n.public_address, i.uuid, i.protocol, i.tag, i.listen_port, i.public_port, i.settings, i.tls_settings, i.transport_settings
    FROM user_groups ug
    JOIN nodes n ON n.group_id = ug.group_id
    JOIN inbounds i ON i.node_id = n.id
    WHERE ug.user_id = ?
    ORDER BY i.id ASC
  `, userID)
  if err != nil {
    return nil, err
  }
  defer rows.Close()

  out := []SubscriptionInbound{}
  for rows.Next() {
    var item SubscriptionInbound
    var publicPort sql.NullInt64
    var tls sql.NullString
    var transport sql.NullString
    var settings string
    if err := rows.Scan(
      &item.NodePublicAddress,
      &item.InboundUUID,
      &item.InboundType,
      &item.InboundTag,
      &item.InboundListenPort,
      &publicPort,
      &settings,
      &tls,
      &transport,
    ); err != nil {
      return nil, err
    }
    if publicPort.Valid {
      item.InboundPublicPort = int(publicPort.Int64)
    }
    item.Settings = []byte(settings)
    if tls.Valid {
      item.TLSSettings = []byte(tls.String)
    }
    if transport.Valid {
      item.TransportSettings = []byte(transport.String)
    }
    out = append(out, item)
  }
  if err := rows.Err(); err != nil {
    return nil, err
  }
  return out, nil
}
