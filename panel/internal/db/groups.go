package db

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

type Group struct {
	ID          int64
	Name        string
	Description string
	MemberCount int64
}

type GroupUpdate struct {
	Name        *string
	Description *string
}

func (s *Store) CreateGroup(ctx context.Context, name, description string) (Group, error) {
	res, err := s.DB.ExecContext(ctx, "INSERT INTO groups (name, description) VALUES (?, ?)", name, description)
	if err != nil {
		if isConflict(err) {
			return Group{}, ErrConflict
		}
		return Group{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Group{}, err
	}
	return s.GetGroupByID(ctx, id)
}

func (s *Store) GetGroupByID(ctx context.Context, id int64) (Group, error) {
	row := s.DB.QueryRowContext(ctx, `
    SELECT
      g.id,
      g.name,
      COALESCE(g.description, ''),
      (SELECT COUNT(1) FROM user_groups ug WHERE ug.group_id = g.id) AS member_count
    FROM groups g
    WHERE g.id = ?
  `, id)
	return scanGroup(row)
}

func (s *Store) ListGroups(ctx context.Context, limit, offset int) ([]Group, error) {
	rows, err := s.DB.QueryContext(ctx, `
    SELECT
      g.id,
      g.name,
      COALESCE(g.description, ''),
      (SELECT COUNT(1) FROM user_groups ug WHERE ug.group_id = g.id) AS member_count
    FROM groups g
    ORDER BY g.id DESC
    LIMIT ? OFFSET ?
  `, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []Group{}
	for rows.Next() {
		g, err := scanGroup(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

func (s *Store) UpdateGroup(ctx context.Context, id int64, update GroupUpdate) (Group, error) {
	sets := []string{}
	args := []any{}
	if update.Name != nil {
		sets = append(sets, "name = ?")
		args = append(args, *update.Name)
	}
	if update.Description != nil {
		sets = append(sets, "description = ?")
		args = append(args, *update.Description)
	}
	if len(sets) == 0 {
		return s.GetGroupByID(ctx, id)
	}
	args = append(args, id)
	res, err := s.DB.ExecContext(ctx, "UPDATE groups SET "+strings.Join(sets, ", ")+" WHERE id = ?", args...)
	if err != nil {
		if isConflict(err) {
			return Group{}, ErrConflict
		}
		return Group{}, err
	}
	n, err := res.RowsAffected()
	if err == nil && n == 0 {
		return Group{}, ErrNotFound
	}
	if err != nil {
		return Group{}, err
	}
	return s.GetGroupByID(ctx, id)
}

func (s *Store) DeleteGroup(ctx context.Context, id int64) error {
	// Prevent deleting a group still referenced by nodes.
	row := s.DB.QueryRowContext(ctx, "SELECT COUNT(1) FROM nodes WHERE group_id = ?", id)
	var cnt int
	if err := row.Scan(&cnt); err != nil {
		return err
	}
	if cnt > 0 {
		return ErrConflict
	}
	res, err := s.DB.ExecContext(ctx, "DELETE FROM groups WHERE id = ?", id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err == nil && n == 0 {
		return ErrNotFound
	}
	return err
}

type groupRowScanner interface {
	Scan(dest ...any) error
}

func scanGroup(row groupRowScanner) (Group, error) {
	var g Group
	if err := row.Scan(&g.ID, &g.Name, &g.Description, &g.MemberCount); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Group{}, ErrNotFound
		}
		return Group{}, err
	}
	return g, nil
}
