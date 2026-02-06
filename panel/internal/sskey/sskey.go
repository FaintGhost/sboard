package sskey

import (
  "crypto/sha256"
  "encoding/base64"
  "errors"
  "strings"

  "github.com/google/uuid"
)

// KeyLength returns the required key length in bytes for a Shadowsocks method.
// Returns 0 for non-2022 methods.
func KeyLength(method string) int {
  switch strings.TrimSpace(method) {
  case "2022-blake3-aes-128-gcm":
    return 16
  case "2022-blake3-aes-256-gcm", "2022-blake3-chacha20-poly1305":
    return 32
  default:
    return 0
  }
}

// DerivePassword derives a password from a UUID for the given Shadowsocks method.
// For 2022 methods, returns a base64-encoded key of appropriate length.
// For classic methods, returns the UUID string as-is.
func DerivePassword(uuidStr, method string) (string, error) {
  keyLen := KeyLength(method)
  if keyLen == 0 {
    // Classic methods use plain string password
    return uuidStr, nil
  }
  return DeriveBase64Key(uuidStr, keyLen)
}

// DeriveBase64Key derives a base64-encoded key from a UUID.
// keyLen must be 16 or 32.
func DeriveBase64Key(uuidStr string, keyLen int) (string, error) {
  id, err := uuid.Parse(strings.TrimSpace(uuidStr))
  if err != nil {
    return "", err
  }
  b := id[:]
  if keyLen == 16 {
    return base64.StdEncoding.EncodeToString(b), nil
  }
  if keyLen == 32 {
    sum := sha256.Sum256(b)
    return base64.StdEncoding.EncodeToString(sum[:]), nil
  }
  return "", errors.New("unsupported key length")
}

// Is2022Method returns true if the method is a Shadowsocks 2022 method.
func Is2022Method(method string) bool {
  return strings.HasPrefix(strings.TrimSpace(method), "2022-")
}
