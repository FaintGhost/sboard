package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// LoadOrGenerateUUID reads a UUID from the file at path. If the file does not
// exist or contains only whitespace, a new UUIDv4 is generated, written
// atomically to path (with 0600 permissions), and returned.
func LoadOrGenerateUUID(path string) (string, error) {
	// Try to read an existing UUID.
	data, err := os.ReadFile(path)
	if err == nil {
		if id := strings.TrimSpace(string(data)); id != "" {
			return id, nil
		}
	}

	// Generate a new UUIDv4.
	id := uuid.NewString()

	// Write atomically: temp file in the same directory, then rename.
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("create uuid dir: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".node-uuid-*")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.WriteString(id); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return "", fmt.Errorf("write uuid: %w", err)
	}
	if err := tmp.Chmod(0600); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return "", fmt.Errorf("chmod uuid file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return "", fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return "", fmt.Errorf("rename uuid file: %w", err)
	}

	return id, nil
}
