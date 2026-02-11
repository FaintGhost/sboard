package api

import "testing"

func TestIsSingboxUA(t *testing.T) {
	truthy := []string{
		"sing-box 1.11",
		"SFA/1.0",
		"SFI/2.0",
	}
	for _, ua := range truthy {
		if !isSingboxUA(ua) {
			t.Fatalf("expected true for ua=%q", ua)
		}
	}

	falsy := []string{
		"clash-meta",
		"v2rayN",
		"",
	}
	for _, ua := range falsy {
		if isSingboxUA(ua) {
			t.Fatalf("expected false for ua=%q", ua)
		}
	}
}
