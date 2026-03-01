package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadOrGenerateUUID_NewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "node-uuid")

	id, err := LoadOrGenerateUUID(path)
	require.NoError(t, err)

	// Must be a valid UUIDv4.
	parsed, err := uuid.Parse(id)
	require.NoError(t, err)
	assert.Equal(t, uuid.Version(4), parsed.Version())

	// File must exist with 0600 permissions.
	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())

	// File content must match returned UUID.
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, id, string(data))
}

func TestLoadOrGenerateUUID_ExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "node-uuid")

	existing := "existing-uuid"
	require.NoError(t, os.WriteFile(path, []byte(existing), 0600))

	id, err := LoadOrGenerateUUID(path)
	require.NoError(t, err)
	assert.Equal(t, existing, id)

	// File content must be unchanged.
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, existing, string(data))
}

func TestLoadOrGenerateUUID_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "node-uuid")

	require.NoError(t, os.WriteFile(path, []byte(""), 0600))

	id, err := LoadOrGenerateUUID(path)
	require.NoError(t, err)

	// Must regenerate a valid UUIDv4.
	parsed, err := uuid.Parse(id)
	require.NoError(t, err)
	assert.Equal(t, uuid.Version(4), parsed.Version())

	// File must be updated.
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, id, string(data))
}

func TestLoadOrGenerateUUID_WhitespaceOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "node-uuid")

	require.NoError(t, os.WriteFile(path, []byte("  \n\t "), 0600))

	id, err := LoadOrGenerateUUID(path)
	require.NoError(t, err)

	// Must regenerate a valid UUIDv4.
	parsed, err := uuid.Parse(id)
	require.NoError(t, err)
	assert.Equal(t, uuid.Version(4), parsed.Version())

	// File must be updated.
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, id, string(data))
}
