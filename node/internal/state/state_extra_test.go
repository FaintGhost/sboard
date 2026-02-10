package state

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestPersist_EdgeCases(t *testing.T) {
	t.Run("empty path", func(t *testing.T) {
		err := Persist("", []byte(`{"ok":true}`))
		if err != nil {
			t.Fatalf("Persist empty path should be nil: %v", err)
		}
	})

	t.Run("mkdirall fails when parent is file", func(t *testing.T) {
		dir := t.TempDir()
		block := filepath.Join(dir, "block")
		if err := os.WriteFile(block, []byte("x"), 0o600); err != nil {
			t.Fatalf("prepare block file: %v", err)
		}

		err := Persist(filepath.Join(block, "state.json"), []byte(`{}`))
		if err == nil {
			t.Fatal("expected mkdirall error")
		}
	})

	t.Run("rename fails when target is directory", func(t *testing.T) {
		dir := t.TempDir()
		err := Persist(dir, []byte(`{}`))
		if err == nil {
			t.Fatal("expected rename error when target is directory")
		}
	})
}

func TestRestore_EdgeCases(t *testing.T) {
	t.Run("empty path", func(t *testing.T) {
		applied, err := Restore("", nil)
		if err != nil {
			t.Fatalf("Restore empty path should be nil: %v", err)
		}
		if applied {
			t.Fatal("Restore empty path should not apply")
		}
	})

	t.Run("apply nil", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "last_sync.json")
		if err := os.WriteFile(path, []byte(`{"inbounds":[]}`), 0o600); err != nil {
			t.Fatalf("write state file failed: %v", err)
		}

		applied, err := Restore(path, nil)
		if err == nil {
			t.Fatal("expected nil apply func error")
		}
		if applied {
			t.Fatal("apply=nil should not mark applied")
		}
	})

	t.Run("apply returns error", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "last_sync.json")
		if err := os.WriteFile(path, []byte(`{"inbounds":[]}`), 0o600); err != nil {
			t.Fatalf("write state file failed: %v", err)
		}

		wantErr := errors.New("apply failed")
		applied, err := Restore(path, func([]byte) error {
			return wantErr
		})
		if !errors.Is(err, wantErr) {
			t.Fatalf("expected apply error, got %v", err)
		}
		if applied {
			t.Fatal("failed apply should not mark applied")
		}
	})
}
