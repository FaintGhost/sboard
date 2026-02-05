package db

import (
  "database/sql"
  "fmt"

  "github.com/golang-migrate/migrate/v4"
  "github.com/golang-migrate/migrate/v4/database/sqlite3"
  _ "github.com/golang-migrate/migrate/v4/source/file"
)

func MigrateUp(db *sql.DB, dir string) error {
  driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
  if err != nil {
    return err
  }
  m, err := migrate.NewWithDatabaseInstance(
    "file://"+dir,
    "sqlite3",
    driver,
  )
  if err != nil {
    return err
  }
  if err := m.Up(); err != nil && err != migrate.ErrNoChange {
    return fmt.Errorf("migrate up: %w", err)
  }
  return nil
}
