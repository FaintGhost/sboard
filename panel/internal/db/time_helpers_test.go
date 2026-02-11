package db

import (
	"testing"
	"time"
)

func TestParseSQLiteTime(t *testing.T) {
	t1, err := parseSQLiteTime("2026-02-11T10:00:00Z")
	if err != nil {
		t.Fatalf("parse RFC3339 failed: %v", err)
	}
	if t1.Year() != 2026 || t1.Month() != 2 || t1.Day() != 11 {
		t.Fatalf("unexpected t1: %v", t1)
	}

	t2, err := parseSQLiteTime("2026-02-11 10:00:00")
	if err != nil {
		t.Fatalf("parse sqlite datetime failed: %v", err)
	}
	if t2.Year() != 2026 || t2.Month() != 2 || t2.Day() != 11 {
		t.Fatalf("unexpected t2: %v", t2)
	}

	if _, err := parseSQLiteTime("bad-time"); err == nil {
		t.Fatal("expect invalid time error")
	}
}

func TestStoreNowUTCAndNowUTCFallback(t *testing.T) {
	s := &Store{Now: func() time.Time { return time.Date(2026, 2, 11, 18, 0, 0, 0, time.FixedZone("UTC+8", 8*3600)) }}
	got := s.NowUTC()
	if got.Location() != time.UTC {
		t.Fatalf("expect utc location, got %v", got.Location())
	}
	if got.Hour() != 10 {
		t.Fatalf("expect 10:00Z, got %v", got)
	}

	s2 := &Store{}
	_ = s2.NowUTC() // should not panic when Now is nil
}
