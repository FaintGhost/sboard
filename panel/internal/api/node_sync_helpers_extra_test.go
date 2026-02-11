package api

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"sboard/panel/internal/node"
)

func TestUsersCountFromAny(t *testing.T) {
	if got := usersCountFromAny([]map[string]any{{"a": 1}, {"b": 2}}); got != 2 {
		t.Fatalf("expect 2, got %d", got)
	}
	if got := usersCountFromAny([]any{"a", 1, true}); got != 3 {
		t.Fatalf("expect 3, got %d", got)
	}
	if got := usersCountFromAny("x"); got != 0 {
		t.Fatalf("expect 0 for invalid type, got %d", got)
	}
}

func TestIntFromAny(t *testing.T) {
	cases := []struct {
		in   any
		want int
	}{
		{in: int(7), want: 7},
		{in: int32(8), want: 8},
		{in: int64(9), want: 9},
		{in: float64(10.9), want: 10},
		{in: float32(11.9), want: 11},
		{in: "12", want: 0},
	}
	for _, tc := range cases {
		if got := intFromAny(tc.in); got != tc.want {
			t.Fatalf("intFromAny(%T=%v) want=%d got=%d", tc.in, tc.in, tc.want, got)
		}
	}
}

func TestPayloadHash_DeterministicAndInsensitiveToSecrets(t *testing.T) {
	p1 := node.SyncPayload{Inbounds: []map[string]any{
		{
			"tag":         "in-1",
			"type":        "vless",
			"listen_port": 443,
			"users": []any{
				map[string]any{"uuid": "u1", "password": "p1"},
			},
			"password": "server-secret-1",
		},
	}}

	p2 := node.SyncPayload{Inbounds: []map[string]any{
		{
			"tag":         "in-1",
			"type":        "vless",
			"listen_port": 443,
			"users": []any{
				map[string]any{"uuid": "u2", "password": "p2"},
			},
			"password": "server-secret-2",
		},
	}}

	h1 := payloadHash(p1)
	h2 := payloadHash(p1)
	if h1 == "" {
		t.Fatal("hash should not be empty")
	}
	if h1 != h2 {
		t.Fatalf("same payload should produce same hash, h1=%s h2=%s", h1, h2)
	}

	// secrets changed but summary fields unchanged => hash unchanged
	if h1 != payloadHash(p2) {
		t.Fatal("hash should ignore secret value changes in payload")
	}

	p3 := node.SyncPayload{Inbounds: []map[string]any{{"tag": "in-1", "type": "vless", "listen_port": 444, "users": []any{}}}}
	if h1 == payloadHash(p3) {
		t.Fatal("hash should change when summary fields change")
	}
}

func TestSleepWithContextAndDurationMSSince(t *testing.T) {
	if err := sleepWithContext(context.Background(), 0); err != nil {
		t.Fatalf("zero delay should return nil: %v", err)
	}

	cctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := sleepWithContext(cctx, 50*time.Millisecond)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expect context canceled, got %v", err)
	}

	started := time.Date(2026, 2, 11, 2, 0, 0, 0, time.UTC)
	finished := started.Add(1500 * time.Millisecond)
	if got := durationMSSince(started, finished); got != 1500 {
		t.Fatalf("expect 1500ms, got %d", got)
	}
	if got := durationMSSince(time.Time{}, finished); got != 0 {
		t.Fatalf("zero start should return 0, got %d", got)
	}
	if got := durationMSSince(started, started.Add(-time.Second)); got != 0 {
		t.Fatalf("finished before start should return 0, got %d", got)
	}
}

func TestSanitizeAndMaskSyncPayloadForLog(t *testing.T) {
	payload := node.SyncPayload{Inbounds: []map[string]any{
		{
			"tag":      "in-1",
			"password": "abcdef1234567890",
			"users": []any{
				map[string]any{"uuid": "12345678-aaaa-bbbb-cccc-1234567890ab", "password": "user-secret-xyz"},
			},
		},
	}}

	out := syncPayloadDebugJSON(payload)
	if strings.Contains(out, "abcdef1234567890") {
		t.Fatalf("top-level password should be masked: %s", out)
	}
	if strings.Contains(out, "12345678-aaaa-bbbb-cccc-1234567890ab") {
		t.Fatalf("uuid should be masked: %s", out)
	}
	if strings.Contains(out, "user-secret-xyz") {
		t.Fatalf("user password should be masked: %s", out)
	}

	if got := maskSyncCredential(""); got != "" {
		t.Fatalf("empty mask should be empty, got %q", got)
	}
	if got := maskSyncCredential("1234567"); got != "***" {
		t.Fatalf("short secret should be ***, got %q", got)
	}
	if got := maskSyncCredential("1234567890"); got != "1234...7890" {
		t.Fatalf("unexpected mask result: %q", got)
	}
}
