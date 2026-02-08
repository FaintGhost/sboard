package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"time"
)

func (s *Store) ListGroupUsers(ctx context.Context, groupID int64) ([]User, error) {
	if groupID <= 0 {
		return nil, errors.New("invalid group id")
	}
	if _, err := s.GetGroupByID(ctx, groupID); err != nil {
		return nil, err
	}

	rows, err := s.DB.QueryContext(ctx, `
    SELECT u.id, u.uuid, u.username, u.traffic_limit, u.traffic_used, u.traffic_reset_day, u.traffic_last_reset_at, u.expire_at, u.status
    FROM user_groups ug
    JOIN users u ON u.id = ug.user_id
    WHERE ug.group_id = ?
    ORDER BY u.id ASC
  `, groupID)
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

func (s *Store) ReplaceGroupUsers(ctx context.Context, groupID int64, userIDs []int64) error {
	if groupID <= 0 {
		return errors.New("invalid group id")
	}
	if _, err := s.GetGroupByID(ctx, groupID); err != nil {
		return err
	}

	uniq := make([]int64, 0, len(userIDs))
	seen := map[int64]struct{}{}
	for _, id := range userIDs {
		if id <= 0 {
			return errors.New("invalid user id")
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniq = append(uniq, id)
	}
	sort.Slice(uniq, func(i, j int) bool { return uniq[i] < uniq[j] })

	if len(uniq) > 0 {
		args := make([]any, 0, len(uniq))
		placeholders := make([]string, 0, len(uniq))
		for _, id := range uniq {
			args = append(args, id)
			placeholders = append(placeholders, "?")
		}
		q := fmt.Sprintf("SELECT COUNT(1) FROM users WHERE id IN (%s)", stringsJoin(placeholders, ","))
		row := s.DB.QueryRowContext(ctx, q, args...)
		var cnt int
		if err := row.Scan(&cnt); err != nil {
			return err
		}
		if cnt != len(uniq) {
			return ErrNotFound
		}
	}

	tx, err := s.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, "DELETE FROM user_groups WHERE group_id = ?", groupID); err != nil {
		return err
	}
	for _, userID := range uniq {
		if _, err := tx.ExecContext(ctx, "INSERT INTO user_groups (user_id, group_id) VALUES (?, ?)", userID, groupID); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

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
      AND (u.traffic_limit <= 0 OR u.traffic_used < u.traffic_limit)
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
