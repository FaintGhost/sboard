package state

import (
	"errors"
	"os"
	"path/filepath"
)

func Persist(path string, raw []byte) error {
	if path == "" {
		return nil
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, "sboard-node-state-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	// Ensure we don't leave temp files around on errors.
	defer func() { _ = os.Remove(tmpName) }()

	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(raw); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	return os.Rename(tmpName, path)
}

// Restore reads a previously persisted payload and calls apply.
// Returns (applied=false, nil) when the state file does not exist.
func Restore(path string, apply func([]byte) error) (bool, error) {
	if path == "" {
		return false, nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	if apply == nil {
		return false, errors.New("apply func is nil")
	}
	if err := apply(raw); err != nil {
		return false, err
	}
	return true, nil
}
