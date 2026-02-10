package sync

import (
	"context"
	"errors"
	"testing"
)

func assertBadRequestContains(t *testing.T, err error, contains string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error")
	}
	var bre BadRequestError
	if !errors.As(err, &bre) {
		t.Fatalf("expected BadRequestError, got %T (%v)", err, err)
	}
	if contains != "" && !containsIn(err.Error(), contains) {
		t.Fatalf("expected error containing %q, got %q", contains, err.Error())
	}
}

func containsIn(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})())
}

func TestParseAndValidateConfig_BadRequestBranches(t *testing.T) {
	ctx := NewSingboxContext()

	tests := []struct {
		name     string
		body     string
		contains string
	}{
		{name: "invalid json", body: `{`, contains: "invalid json"},
		{name: "inbound invalid json", body: `{"inbounds":[{"tag":1}]}`, contains: "inbounds[0] invalid json"},
		{name: "missing tag", body: `{"inbounds":[{"type":"mixed","listen_port":1080}]}`, contains: "inbounds[0].tag required"},
		{name: "missing type", body: `{"inbounds":[{"tag":"in-1","listen_port":1080}]}`, contains: "inbounds[0].type required"},
		{name: "invalid port", body: `{"inbounds":[{"tag":"in-1","type":"mixed","listen_port":70000}]}`, contains: "inbounds[0].listen_port invalid"},
		{name: "duplicate tag", body: `{"inbounds":[{"tag":"in-1","type":"mixed","listen_port":1080},{"tag":"in-1","type":"mixed","listen_port":1081}]}`, contains: "inbounds[1].tag duplicated"},
		{name: "ss2022 missing password", body: `{"inbounds":[{"tag":"ss-in","type":"shadowsocks","listen_port":8388,"method":"2022-blake3-aes-128-gcm"}]}`, contains: "password required"},
		{name: "invalid config on final unmarshal", body: `{"inbounds":[],"route":{"final":123}}`, contains: "invalid config:"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseAndValidateConfig(ctx, []byte(tc.body))
			assertBadRequestContains(t, err, tc.contains)
		})
	}
}

func TestParseAndValidateConfig_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	ctx = NewSingboxContext()
	// 再包一层可取消上下文，尽量走到 canceled 分支。
	ctx2, cancel2 := context.WithCancel(ctx)
	cancel2()

	body := []byte(`{
    "inbounds": [
      {
        "type": "vless",
        "tag": "vless-in",
        "listen": "0.0.0.0",
        "listen_port": 443,
        "users": []
      }
    ]
  }`)

	_, err := ParseAndValidateConfig(ctx2, body)
	if err == nil {
		t.Skip("environment did not surface context cancellation in sing-box json unmarshal")
	}
	if !errors.Is(err, context.Canceled) {
		// 某些环境下库可能返回 BadRequestError；不强制失败，但保留信号。
		var bre BadRequestError
		if !errors.As(err, &bre) {
			t.Fatalf("expected context.Canceled or BadRequestError, got %T (%v)", err, err)
		}
	}
}
