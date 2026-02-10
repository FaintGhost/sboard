package api

import "testing"

func TestShouldDebugNodeSyncPayload(t *testing.T) {
	tests := []struct {
		name   string
		value  string
		expect bool
	}{
		{name: "empty", value: "", expect: false},
		{name: "true", value: "true", expect: true},
		{name: "one", value: "1", expect: true},
		{name: "yes with spaces", value: "  YES  ", expect: true},
		{name: "on", value: "on", expect: true},
		{name: "false", value: "false", expect: false},
		{name: "zero", value: "0", expect: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(nodeSyncDebugPayloadEnv, tc.value)
			if got := shouldDebugNodeSyncPayload(); got != tc.expect {
				t.Fatalf("expect %v, got %v", tc.expect, got)
			}
		})
	}
}
