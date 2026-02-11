package db

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrateUp_EmbeddedAndIdempotent(t *testing.T) {
	dir := t.TempDir()
	database, err := Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("open db failed: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	// dir empty => use embedded migrations
	if err := MigrateUp(database, ""); err != nil {
		t.Fatalf("embedded migrate failed: %v", err)
	}

	// run again should be no-op
	if err := MigrateUp(database, ""); err != nil {
		t.Fatalf("second migrate should be no-op, got: %v", err)
	}
}

func TestMigrateUp_InvalidDir(t *testing.T) {
	dir := t.TempDir()
	database, err := Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("open db failed: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	err = MigrateUp(database, filepath.Join(dir, "not-exist"))
	if err == nil {
		t.Fatal("expect migrate error for invalid dir")
	}
	if strings.TrimSpace(err.Error()) == "" {
		t.Fatalf("unexpected empty migrate error")
	}
}
