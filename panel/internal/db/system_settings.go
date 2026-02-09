package db

import (
	"context"
	"database/sql"
	"errors"
)

func (s *Store) GetSystemSetting(ctx context.Context, key string) (string, error) {
	row := s.DB.QueryRowContext(ctx, "SELECT value FROM system_settings WHERE key = ? LIMIT 1", key)
	var value string
	if err := row.Scan(&value); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}
	return value, nil
}

func (s *Store) UpsertSystemSetting(ctx context.Context, key, value string) error {
	_, err := s.DB.ExecContext(
		ctx,
		`INSERT INTO system_settings (key, value, updated_at)
     VALUES (?, ?, CURRENT_TIMESTAMP)
     ON CONFLICT(key) DO UPDATE SET
       value = excluded.value,
       updated_at = CURRENT_TIMESTAMP`,
		key,
		value,
	)
	return err
}

func (s *Store) DeleteSystemSetting(ctx context.Context, key string) error {
	_, err := s.DB.ExecContext(ctx, "DELETE FROM system_settings WHERE key = ?", key)
	return err
}
