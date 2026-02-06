package sskey

import (
  "encoding/base64"
  "testing"
)

func TestKeyLength(t *testing.T) {
  tests := []struct {
    method   string
    expected int
  }{
    {"2022-blake3-aes-128-gcm", 16},
    {"2022-blake3-aes-256-gcm", 32},
    {"2022-blake3-chacha20-poly1305", 32},
    {"aes-128-gcm", 0},
    {"chacha20-ietf-poly1305", 0},
    {"none", 0},
    {"", 0},
  }

  for _, tt := range tests {
    got := KeyLength(tt.method)
    if got != tt.expected {
      t.Errorf("KeyLength(%q) = %d, want %d", tt.method, got, tt.expected)
    }
  }
}

func TestIs2022Method(t *testing.T) {
  tests := []struct {
    method   string
    expected bool
  }{
    {"2022-blake3-aes-128-gcm", true},
    {"2022-blake3-aes-256-gcm", true},
    {"2022-blake3-chacha20-poly1305", true},
    {"aes-128-gcm", false},
    {"none", false},
    {"", false},
  }

  for _, tt := range tests {
    got := Is2022Method(tt.method)
    if got != tt.expected {
      t.Errorf("Is2022Method(%q) = %v, want %v", tt.method, got, tt.expected)
    }
  }
}

func TestDerivePassword(t *testing.T) {
  uuid := "6fd47678-1f45-48f1-8051-fdbcdc2a3ccb"

  // Classic method should return UUID as-is
  pw, err := DerivePassword(uuid, "aes-128-gcm")
  if err != nil {
    t.Fatalf("DerivePassword for classic method: %v", err)
  }
  if pw != uuid {
    t.Errorf("classic method password = %q, want %q", pw, uuid)
  }

  // 2022-aes-128-gcm should return 16-byte base64 (24 chars)
  pw128, err := DerivePassword(uuid, "2022-blake3-aes-128-gcm")
  if err != nil {
    t.Fatalf("DerivePassword for 2022-aes-128: %v", err)
  }
  decoded128, err := base64.StdEncoding.DecodeString(pw128)
  if err != nil {
    t.Fatalf("decode 2022-aes-128 password: %v", err)
  }
  if len(decoded128) != 16 {
    t.Errorf("2022-aes-128 password decoded length = %d, want 16", len(decoded128))
  }

  // 2022-aes-256-gcm should return 32-byte base64 (44 chars)
  pw256, err := DerivePassword(uuid, "2022-blake3-aes-256-gcm")
  if err != nil {
    t.Fatalf("DerivePassword for 2022-aes-256: %v", err)
  }
  decoded256, err := base64.StdEncoding.DecodeString(pw256)
  if err != nil {
    t.Fatalf("decode 2022-aes-256 password: %v", err)
  }
  if len(decoded256) != 32 {
    t.Errorf("2022-aes-256 password decoded length = %d, want 32", len(decoded256))
  }

  // Same UUID should produce same password (deterministic)
  pw256Again, _ := DerivePassword(uuid, "2022-blake3-aes-256-gcm")
  if pw256Again != pw256 {
    t.Error("DerivePassword is not deterministic")
  }
}

func TestDeriveBase64Key_InvalidUUID(t *testing.T) {
  _, err := DeriveBase64Key("not-a-uuid", 16)
  if err == nil {
    t.Error("expected error for invalid UUID")
  }
}
