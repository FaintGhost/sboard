package db

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func MigrateUp(db *sql.DB, dir string) error {
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return err
	}

	// Prefer embedded migrations when dir is empty. This avoids runtime path issues in Docker images.
	var m *migrate.Migrate
	if dir == "" {
		src, err := iofs.New(migrationsFS, "migrations")
		if err != nil {
			return err
		}
		m, err = migrate.NewWithInstance("iofs", src, "sqlite3", driver)
		if err != nil {
			return err
		}
	} else {
		m, err = migrate.NewWithDatabaseInstance(
			"file://"+dir,
			"sqlite3",
			driver,
		)
		if err != nil {
			return err
		}
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}
