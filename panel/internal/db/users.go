package db

import (
  "context"
  "database/sql"
  "errors"
  "fmt"
  "strings"
  "time"

  "github.com/google/uuid"
  sqlite3 "github.com/mattn/go-sqlite3"
)

var (
  ErrConflict = errors.New("conflict")
  ErrNotFound = errors.New("not found")
)

type User struct {
  ID              int64
  UUID            string
  Username        string
  TrafficLimit    int64
  TrafficUsed     int64
  TrafficResetDay int
  ExpireAt        *time.Time
  Status          string
}

type UserUpdate struct {
  Username        *string
  Status          *string
  ExpireAt        *time.Time
  ExpireAtSet     bool
  TrafficLimit    *int64
  TrafficResetDay *int
}

func (s *Store) CreateUser(ctx context.Context, username string) (User, error) {
  id := uuid.NewString()
  res, err := s.DB.ExecContext(ctx, "INSERT INTO users (uuid, username) VALUES (?, ?)", id, username)
  if err != nil {
    if isConflict(err) {
      return User{}, ErrConflict
    }
    return User{}, err
  }
  rowID, err := res.LastInsertId()
  if err != nil {
    return User{}, err
  }
  return s.GetUserByID(ctx, rowID)
}

func (s *Store) GetUserByID(ctx context.Context, id int64) (User, error) {
  row := s.DB.QueryRowContext(ctx, `SELECT id, uuid, username, traffic_limit, traffic_used, traffic_reset_day, expire_at, status
    FROM users WHERE id = ?`, id)
  return scanUser(row)
}

func (s *Store) GetUserByUUID(ctx context.Context, uuid string) (User, error) {
  row := s.DB.QueryRowContext(ctx, `SELECT id, uuid, username, traffic_limit, traffic_used, traffic_reset_day, expire_at, status
    FROM users WHERE uuid = ?`, uuid)
  return scanUser(row)
}

func (s *Store) ListUsers(ctx context.Context, limit, offset int, status string) ([]User, error) {
  args := []any{}
  where := ""
  if status == "" {
    where = "WHERE status != 'disabled'"
  } else {
    where = "WHERE status = ?"
    args = append(args, status)
  }
  args = append(args, limit, offset)
  rows, err := s.DB.QueryContext(ctx, fmt.Sprintf(`SELECT id, uuid, username, traffic_limit, traffic_used, traffic_reset_day, expire_at, status
    FROM users %s ORDER BY id DESC LIMIT ? OFFSET ?`, where), args...)
  if err != nil {
    return nil, err
  }
  defer rows.Close()

  out := []User{}
  for rows.Next() {
    u, err := scanUser(rows)
    if err != nil {
      return nil, err
    }
    out = append(out, u)
  }
  return out, rows.Err()
}

func (s *Store) UpdateUser(ctx context.Context, id int64, update UserUpdate) (User, error) {
  sets := []string{}
  args := []any{}

  if update.Username != nil {
    sets = append(sets, "username = ?")
    args = append(args, *update.Username)
  }
  if update.Status != nil {
    sets = append(sets, "status = ?")
    args = append(args, *update.Status)
  }
  if update.ExpireAtSet {
    sets = append(sets, "expire_at = ?")
    if update.ExpireAt != nil {
      args = append(args, sql.NullTime{Time: *update.ExpireAt, Valid: true})
    } else {
      args = append(args, sql.NullTime{})
    }
  }
  if update.TrafficLimit != nil {
    sets = append(sets, "traffic_limit = ?")
    args = append(args, *update.TrafficLimit)
  }
  if update.TrafficResetDay != nil {
    sets = append(sets, "traffic_reset_day = ?")
    args = append(args, *update.TrafficResetDay)
  }

  if len(sets) == 0 {
    return s.GetUserByID(ctx, id)
  }

  args = append(args, id)
  _, err := s.DB.ExecContext(ctx, "UPDATE users SET "+strings.Join(sets, ", ")+" WHERE id = ?", args...)
  if err != nil {
    if isConflict(err) {
      return User{}, ErrConflict
    }
    return User{}, err
  }
  return s.GetUserByID(ctx, id)
}

func (s *Store) DisableUser(ctx context.Context, id int64) error {
  res, err := s.DB.ExecContext(ctx, "UPDATE users SET status = 'disabled' WHERE id = ?", id)
  if err != nil {
    return err
  }
  rows, err := res.RowsAffected()
  if err == nil && rows == 0 {
    return ErrNotFound
  }
  return err
}

type rowScanner interface {
  Scan(dest ...any) error
}

func scanUser(row rowScanner) (User, error) {
  var u User
  var expire sql.NullString
  err := row.Scan(
    &u.ID,
    &u.UUID,
    &u.Username,
    &u.TrafficLimit,
    &u.TrafficUsed,
    &u.TrafficResetDay,
    &expire,
    &u.Status,
  )
  if err != nil {
    if errors.Is(err, sql.ErrNoRows) {
      return User{}, ErrNotFound
    }
    return User{}, err
  }
  if expire.Valid {
    if t, err := time.Parse(time.RFC3339, expire.String); err == nil {
      u.ExpireAt = &t
    } else if t, err := time.Parse("2006-01-02 15:04:05", expire.String); err == nil {
      u.ExpireAt = &t
    }
  }
  return u, nil
}

func isConflict(err error) bool {
  if errors.Is(err, sqlite3.ErrConstraint) {
    return true
  }
  var se sqlite3.Error
  if errors.As(err, &se) {
    if se.ExtendedCode == sqlite3.ErrConstraintUnique {
      return true
    }
  }
  return strings.Contains(err.Error(), "UNIQUE constraint failed")
}
