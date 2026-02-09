package db

import (
  "context"
  "database/sql"
  "errors"
  "fmt"
  "sort"
)

func (s *Store) ListUserGroupIDs(ctx context.Context, userID int64) ([]int64, error) {
  rows, err := s.DB.QueryContext(ctx, "SELECT group_id FROM user_groups WHERE user_id = ? ORDER BY group_id ASC", userID)
  if err != nil {
    return nil, err
  }
  defer rows.Close()
  out := []int64{}
  for rows.Next() {
    var id int64
    if err := rows.Scan(&id); err != nil {
      return nil, err
    }
    out = append(out, id)
  }
  return out, rows.Err()
}

// ListUserGroupIDsBatch returns a map from user IDs to their group IDs.
// This avoids N+1 queries when building DTOs for multiple users.
func (s *Store) ListUserGroupIDsBatch(ctx context.Context, userIDs []int64) (map[int64][]int64, error) {
  result := make(map[int64][]int64, len(userIDs))
  if len(userIDs) == 0 {
    return result, nil
  }

  // Initialize all user IDs with empty slices
  for _, id := range userIDs {
    result[id] = []int64{}
  }

  // Build query with IN clause
  placeholders := make([]string, len(userIDs))
  args := make([]any, len(userIDs))
  for i, id := range userIDs {
    placeholders[i] = "?"
    args[i] = id
  }

  query := fmt.Sprintf(
    "SELECT user_id, group_id FROM user_groups WHERE user_id IN (%s) ORDER BY user_id, group_id ASC",
    stringsJoin(placeholders, ","),
  )

  rows, err := s.DB.QueryContext(ctx, query, args...)
  if err != nil {
    return nil, err
  }
  defer rows.Close()

  for rows.Next() {
    var userID, groupID int64
    if err := rows.Scan(&userID, &groupID); err != nil {
      return nil, err
    }
    result[userID] = append(result[userID], groupID)
  }

  return result, rows.Err()
}

// ReplaceUserGroups replaces the user's group memberships (set semantics).
func (s *Store) ReplaceUserGroups(ctx context.Context, userID int64, groupIDs []int64) error {
  // Verify user exists early to give consistent errors.
  if _, err := s.GetUserByID(ctx, userID); err != nil {
    return err
  }

  uniq := make([]int64, 0, len(groupIDs))
  seen := map[int64]struct{}{}
  for _, id := range groupIDs {
    if id <= 0 {
      return errors.New("invalid group id")
    }
    if _, ok := seen[id]; ok {
      continue
    }
    seen[id] = struct{}{}
    uniq = append(uniq, id)
  }
  sort.Slice(uniq, func(i, j int) bool { return uniq[i] < uniq[j] })

  if len(uniq) > 0 {
    // Validate referenced groups exist.
    args := make([]any, 0, len(uniq))
    placeholders := make([]string, 0, len(uniq))
    for _, id := range uniq {
      args = append(args, id)
      placeholders = append(placeholders, "?")
    }
    q := fmt.Sprintf("SELECT COUNT(1) FROM groups WHERE id IN (%s)", stringsJoin(placeholders, ","))
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

  if _, err := tx.ExecContext(ctx, "DELETE FROM user_groups WHERE user_id = ?", userID); err != nil {
    return err
  }
  for _, gid := range uniq {
    if _, err := tx.ExecContext(ctx, "INSERT INTO user_groups (user_id, group_id) VALUES (?, ?)", userID, gid); err != nil {
      return err
    }
  }
  if err := tx.Commit(); err != nil {
    return err
  }
  return nil
}

func stringsJoin(parts []string, sep string) string {
  if len(parts) == 0 {
    return ""
  }
  out := parts[0]
  for i := 1; i < len(parts); i++ {
    out += sep + parts[i]
  }
  return out
}

