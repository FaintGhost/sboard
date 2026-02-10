package db

import (
	"database/sql"
	"errors"
	"strings"
)

type Admin struct {
	ID           int
	Username     string
	PasswordHash string
}

func AdminCount(store *Store) (int, error) {
	if store == nil || store.DB == nil {
		return 0, errors.New("nil store")
	}
	row := store.DB.QueryRow(`SELECT COUNT(1) FROM admins`)
	var n int
	if err := row.Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

func AdminGetByID(store *Store, id int) (Admin, bool, error) {
	if store == nil || store.DB == nil {
		return Admin{}, false, errors.New("nil store")
	}
	if id <= 0 {
		return Admin{}, false, nil
	}
	row := store.DB.QueryRow(`SELECT id, username, password_hash FROM admins WHERE id = ? LIMIT 1`, id)
	var a Admin
	if err := row.Scan(&a.ID, &a.Username, &a.PasswordHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Admin{}, false, nil
		}
		return Admin{}, false, err
	}
	return a, true, nil
}

func AdminGetByUsername(store *Store, username string) (Admin, bool, error) {
	if store == nil || store.DB == nil {
		return Admin{}, false, errors.New("nil store")
	}
	u := strings.TrimSpace(username)
	if u == "" {
		return Admin{}, false, nil
	}
	row := store.DB.QueryRow(`SELECT id, username, password_hash FROM admins WHERE username = ? LIMIT 1`, u)
	var a Admin
	if err := row.Scan(&a.ID, &a.Username, &a.PasswordHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Admin{}, false, nil
		}
		return Admin{}, false, err
	}
	return a, true, nil
}

func AdminGetFirst(store *Store) (Admin, bool, error) {
	if store == nil || store.DB == nil {
		return Admin{}, false, errors.New("nil store")
	}
	row := store.DB.QueryRow(`SELECT id, username, password_hash FROM admins ORDER BY id ASC LIMIT 1`)
	var a Admin
	if err := row.Scan(&a.ID, &a.Username, &a.PasswordHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Admin{}, false, nil
		}
		return Admin{}, false, err
	}
	return a, true, nil
}

func AdminUpdateCredentials(store *Store, id int, username, passwordHash string) error {
	if store == nil || store.DB == nil {
		return errors.New("nil store")
	}
	if id <= 0 {
		return ErrNotFound
	}

	u := strings.TrimSpace(username)
	if u == "" {
		return errors.New("missing username")
	}
	if strings.TrimSpace(passwordHash) == "" {
		return errors.New("missing password_hash")
	}

	res, err := store.DB.Exec(
		`UPDATE admins
     SET username = ?,
         password_hash = ?,
         updated_at = CURRENT_TIMESTAMP
     WHERE id = ?`,
		u,
		passwordHash,
		id,
	)
	if err != nil {
		if isConflict(err) {
			return ErrConflict
		}
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// AdminCreateIfNone creates the first admin only if the admins table is empty.
// Returns (created=false, nil) when an admin already exists.
func AdminCreateIfNone(store *Store, username, passwordHash string) (created bool, err error) {
	if store == nil || store.DB == nil {
		return false, errors.New("nil store")
	}
	u := strings.TrimSpace(username)
	if u == "" {
		return false, errors.New("missing username")
	}
	if strings.TrimSpace(passwordHash) == "" {
		return false, errors.New("missing password_hash")
	}

	tx, err := store.DB.Begin()
	if err != nil {
		return false, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var n int
	if err := tx.QueryRow(`SELECT COUNT(1) FROM admins`).Scan(&n); err != nil {
		return false, err
	}
	if n > 0 {
		err = tx.Rollback()
		return false, err
	}

	if _, err := tx.Exec(`INSERT INTO admins (username, password_hash) VALUES (?, ?)`, u, passwordHash); err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}
