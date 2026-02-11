package api

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"sboard/panel/internal/db"
)

func TestNormalizeAndSetSystemTimezone(t *testing.T) {
	_, _ = setSystemTimezone(defaultSystemTimezone)

	name, loc, err := normalizeSystemTimezone("")
	if err != nil {
		t.Fatalf("normalize empty failed: %v", err)
	}
	if name != defaultSystemTimezone || loc == nil {
		t.Fatalf("unexpected normalize result name=%s loc=%v", name, loc)
	}

	if _, _, err := normalizeSystemTimezone("Mars/Base"); err == nil {
		t.Fatal("expect invalid timezone error")
	}

	prev := currentSystemTimezoneName()
	if _, err := setSystemTimezone("Mars/Base"); err == nil {
		t.Fatal("expect set timezone error")
	}
	if got := currentSystemTimezoneName(); got != prev {
		t.Fatalf("timezone should stay unchanged, want=%s got=%s", prev, got)
	}

	if _, err := setSystemTimezone("Asia/Shanghai"); err != nil {
		t.Fatalf("set timezone failed: %v", err)
	}
	if got := currentSystemTimezoneName(); got != "Asia/Shanghai" {
		t.Fatalf("unexpected timezone name: %s", got)
	}
	_, _ = setSystemTimezone(defaultSystemTimezone)
}

func TestInitSystemTimezone_WithNilAndStore(t *testing.T) {
	_, _ = setSystemTimezone("Asia/Shanghai")
	if err := InitSystemTimezone(context.Background(), nil); err != nil {
		t.Fatalf("init with nil store failed: %v", err)
	}
	if got := currentSystemTimezoneName(); got != defaultSystemTimezone {
		t.Fatalf("nil store should reset to default, got=%s", got)
	}

	dir := t.TempDir()
	database, err := db.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("open db failed: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	if err := db.MigrateUp(database, ""); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}

	store := db.NewStore(database)

	// no setting => default
	if err := InitSystemTimezone(context.Background(), store); err != nil {
		t.Fatalf("init with empty setting failed: %v", err)
	}
	if got := currentSystemTimezoneName(); got != defaultSystemTimezone {
		t.Fatalf("expected default timezone, got=%s", got)
	}

	// valid setting
	if _, err := store.DB.Exec("INSERT INTO system_settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value=excluded.value", systemTimezoneKey, "Asia/Shanghai"); err != nil {
		t.Fatalf("set timezone in db failed: %v", err)
	}
	if err := InitSystemTimezone(context.Background(), store); err != nil {
		t.Fatalf("init with valid setting failed: %v", err)
	}
	if got := currentSystemTimezoneName(); got != "Asia/Shanghai" {
		t.Fatalf("expected Asia/Shanghai, got=%s", got)
	}

	// invalid setting => fallback UTC without error
	if _, err := store.DB.Exec("UPDATE system_settings SET value = ? WHERE key = ?", "Mars/Base", systemTimezoneKey); err != nil {
		t.Fatalf("update invalid timezone failed: %v", err)
	}
	if err := InitSystemTimezone(context.Background(), store); err != nil {
		t.Fatalf("init with invalid setting should not error, got=%v", err)
	}
	if got := currentSystemTimezoneName(); got != defaultSystemTimezone {
		t.Fatalf("invalid setting should fallback UTC, got=%s", got)
	}

	_, _ = setSystemTimezone(defaultSystemTimezone)
}

func TestFormatAndTimeInSystemTimezone(t *testing.T) {
	_, _ = setSystemTimezone("Asia/Shanghai")
	defer func() { _, _ = setSystemTimezone(defaultSystemTimezone) }()

	if got := formatTimeRFC3339OrEmpty(time.Time{}); got != "" {
		t.Fatalf("zero time should format empty, got=%q", got)
	}

	ts := time.Date(2026, 2, 11, 2, 0, 0, 0, time.UTC)
	if got := formatTimeRFC3339OrEmpty(ts); got != "2026-02-11T10:00:00+08:00" {
		t.Fatalf("unexpected formatted time: %s", got)
	}

	if got := timeInSystemTimezonePtr(nil); got != nil {
		t.Fatalf("nil time ptr should stay nil")
	}
	converted := timeInSystemTimezonePtr(&ts)
	if converted == nil || converted.Format(time.RFC3339) != "2026-02-11T10:00:00+08:00" {
		t.Fatalf("unexpected converted time: %v", converted)
	}
}
