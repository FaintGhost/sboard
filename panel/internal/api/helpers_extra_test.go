package api

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestParseWindowOrDefault(t *testing.T) {
	def := 24 * time.Hour
	if got := parseWindowOrDefault("", def); got != def {
		t.Fatalf("empty window expect %v, got %v", def, got)
	}
	if got := parseWindowOrDefault("all", def); got != 0 {
		t.Fatalf("all window expect 0, got %v", got)
	}
	if got := parseWindowOrDefault("7d", def); got != 7*24*time.Hour {
		t.Fatalf("7d expect 7 days, got %v", got)
	}
	if got := parseWindowOrDefault("999d", def); got != 90*24*time.Hour {
		t.Fatalf("999d should be clamped, got %v", got)
	}
	if got := parseWindowOrDefault("2h", def); got != 2*time.Hour {
		t.Fatalf("2h expect 2h, got %v", got)
	}
	if got := parseWindowOrDefault("30s", def); got != -1 {
		t.Fatalf("30s should be invalid, got %v", got)
	}
	if got := parseWindowOrDefault("0d", def); got != -1 {
		t.Fatalf("0d should be invalid, got %v", got)
	}
	if got := parseWindowOrDefault("bad", def); got != -1 {
		t.Fatalf("bad should be invalid, got %v", got)
	}
}

func TestParseBoolQuery(t *testing.T) {
	truthy := []string{"1", "true", "TRUE", "yes", "y", "on", "  on  "}
	for _, v := range truthy {
		if !parseBoolQuery(v) {
			t.Fatalf("expect truthy for %q", v)
		}
	}

	falsy := []string{"", "0", "false", "off", "no", "random"}
	for _, v := range falsy {
		if parseBoolQuery(v) {
			t.Fatalf("expect falsy for %q", v)
		}
	}
}

func TestSyncErrorHelpers(t *testing.T) {
	if !isNodeUnreachableSyncError("node sync request failed: timeout") {
		t.Fatal("expect unreachable error detected")
	}
	if isNodeUnreachableSyncError("other error") {
		t.Fatal("unexpected unreachable detection")
	}

	if got := normalizeSyncError(""); got != "sync failed" {
		t.Fatalf("empty error expect sync failed, got %q", got)
	}
	if got := normalizeSyncError("  boom  "); got != "boom" {
		t.Fatalf("trimmed error expect boom, got %q", got)
	}

	veryLong := strings.Repeat("x", maxSyncErrorSummaryLn+200)
	if got := normalizeSyncError(veryLong); len(got) != maxSyncErrorSummaryLn {
		t.Fatalf("long error should be truncated to %d, got %d", maxSyncErrorSummaryLn, len(got))
	}

	if got := normalizeSyncClientError(nil); got != "" {
		t.Fatalf("nil error expect empty string, got %q", got)
	}
	if got := normalizeSyncClientError(errors.New("timeout")); got != "node sync request failed: timeout" {
		t.Fatalf("unexpected normalizeSyncClientError: %q", got)
	}
	if got := normalizeSyncClientError(errors.New("node sync status 400: bad payload")); got != "node sync status 400: bad payload" {
		t.Fatalf("unexpected normalizeSyncClientError passthrough: %q", got)
	}

	if code := parseSyncHTTPStatus(nil); code != 200 {
		t.Fatalf("nil error should map to 200, got %d", code)
	}
	if code := parseSyncHTTPStatus(errors.New("node sync status 502: bad gateway")); code != 502 {
		t.Fatalf("expect 502, got %d", code)
	}
	if code := parseSyncHTTPStatus(errors.New("node sync status 99: invalid")); code != 0 {
		t.Fatalf("invalid status should map to 0, got %d", code)
	}
	if code := parseSyncHTTPStatus(errors.New("random error")); code != 0 {
		t.Fatalf("non-status error should map to 0, got %d", code)
	}
}

func TestShouldDebugSyncPayload(t *testing.T) {
	tests := []struct {
		name   string
		value  string
		expect bool
	}{
		{name: "empty", value: "", expect: false},
		{name: "true", value: "true", expect: true},
		{name: "one", value: "1", expect: true},
		{name: "yes", value: "  YES  ", expect: true},
		{name: "on", value: "on", expect: true},
		{name: "false", value: "false", expect: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(panelSyncDebugPayloadEnv, tc.value)
			if got := shouldDebugSyncPayload(); got != tc.expect {
				t.Fatalf("expect %v, got %v", tc.expect, got)
			}
		})
	}
}

func TestGenerateSetupToken(t *testing.T) {
	tok, err := GenerateSetupToken()
	if err != nil {
		t.Fatalf("GenerateSetupToken failed: %v", err)
	}
	if len(tok) < 32 {
		t.Fatalf("token too short: %d", len(tok))
	}
	if strings.Contains(tok, "=") {
		t.Fatalf("token should be raw url encoding without padding: %q", tok)
	}

	tok2, err := GenerateSetupToken()
	if err != nil {
		t.Fatalf("GenerateSetupToken second call failed: %v", err)
	}
	if tok == tok2 {
		t.Fatal("two generated tokens should not be equal")
	}
}
