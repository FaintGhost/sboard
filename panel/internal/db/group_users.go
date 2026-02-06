package db

import (
  "context"
  "fmt"
  "time"
)

func (s *Store) ListActiveUsersForGroup(ctx context.Context, groupID int64) ([]User, error) {
  now := s.nowUTC()
  nowStr := now.Format(time.RFC3339)
  rows, err := s.DB.QueryContext(ctx, `
    SELECT u.id, u.uuid, u.username, u.traffic_limit, u.traffic_used, u.traffic_reset_day, u.traffic_last_reset_at, u.expire_at, u.status
    FROM user_groups ug
    JOIN users u ON u.id = ug.user_id
    WHERE ug.group_id = ?
      AND u.status = 'active'
      AND (u.expire_at IS NULL OR u.expire_at > ?)
    ORDER BY u.id ASC
  `, groupID, nowStr)
  if err != nil {
    return nil, err
  }
  defer rows.Close()

  out := []User{}
  for rows.Next() {
    u, err := scanUser(rows)
    if err != nil {
      return nil, fmt.Errorf("scan user: %w", err)
    }
    if err := s.applyTrafficResetIfNeeded(ctx, &u); err != nil {
      return nil, err
    }
    out = append(out, u)
  }
  if err := rows.Err(); err != nil {
    return nil, err
  }
  return out, nil
}

