package state

import (
	"path/filepath"
	"testing"
)

func TestPersistAndRestore(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "last_sync.json")
	raw := []byte(`{"inbounds":[{"type":"mixed","tag":"m1","listen":"0.0.0.0","listen_port":1080}]}`)

	if err := Persist(p, raw); err != nil {
		t.Fatalf("Persist: %v", err)
	}

	var got []byte
	applied, err := Restore(p, func(b []byte) error {
		got = append([]byte(nil), b...)
		return nil
	})
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if !applied {
		t.Fatalf("expected applied=true")
	}
	if string(got) != string(raw) {
		t.Fatalf("unexpected payload\n got: %s\nwant: %s", string(got), string(raw))
	}
}

func TestRestoreMissing(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "missing.json")
	applied, err := Restore(p, func([]byte) error { return nil })
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if applied {
		t.Fatalf("expected applied=false")
	}
}
