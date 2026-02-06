package password

import (
  "crypto/hmac"
  "crypto/rand"
  "crypto/sha256"
  "crypto/subtle"
  "encoding/base64"
  "errors"
  "fmt"
  "strconv"
  "strings"
)

const (
  schemePBKDF2SHA256 = "pbkdf2_sha256"
  defaultIter        = 50000
  saltBytes          = 16
  keyBytes           = 32
)

// Hash returns a self-describing password hash string.
//
// Format:
//   pbkdf2_sha256$<iter>$<salt_b64>$<dk_b64>
func Hash(raw string) (string, error) {
  pw := []byte(raw)
  if len(strings.TrimSpace(raw)) == 0 {
    return "", errors.New("empty password")
  }
  salt := make([]byte, saltBytes)
  if _, err := rand.Read(salt); err != nil {
    return "", err
  }
  dk := pbkdf2SHA256(pw, salt, defaultIter, keyBytes)
  return fmt.Sprintf(
    "%s$%d$%s$%s",
    schemePBKDF2SHA256,
    defaultIter,
    base64.StdEncoding.EncodeToString(salt),
    base64.StdEncoding.EncodeToString(dk),
  ), nil
}

func Verify(hash, raw string) bool {
  parts := strings.Split(hash, "$")
  if len(parts) != 4 {
    return false
  }
  if parts[0] != schemePBKDF2SHA256 {
    return false
  }
  iter, err := strconv.Atoi(parts[1])
  if err != nil || iter <= 0 {
    return false
  }
  salt, err := base64.StdEncoding.DecodeString(parts[2])
  if err != nil || len(salt) == 0 {
    return false
  }
  want, err := base64.StdEncoding.DecodeString(parts[3])
  if err != nil || len(want) == 0 {
    return false
  }
  got := pbkdf2SHA256([]byte(raw), salt, iter, len(want))
  return subtle.ConstantTimeCompare(got, want) == 1
}

// pbkdf2SHA256 implements PBKDF2(HMAC-SHA256) in a small, dependency-free way.
// RFC 8018, section 5.2.
func pbkdf2SHA256(password, salt []byte, iter, keyLen int) []byte {
  hLen := sha256.Size
  nBlocks := (keyLen + hLen - 1) / hLen
  out := make([]byte, 0, nBlocks*hLen)

  for block := 1; block <= nBlocks; block++ {
    // U1 = PRF(password, salt || INT(block))
    mac := hmac.New(sha256.New, password)
    mac.Write(salt)
    mac.Write([]byte{byte(block >> 24), byte(block >> 16), byte(block >> 8), byte(block)})
    u := mac.Sum(nil)

    t := make([]byte, hLen)
    copy(t, u)

    for i := 2; i <= iter; i++ {
      mac = hmac.New(sha256.New, password)
      mac.Write(u)
      u = mac.Sum(nil)
      for j := 0; j < hLen; j++ {
        t[j] ^= u[j]
      }
    }

    out = append(out, t...)
  }

  return out[:keyLen]
}

