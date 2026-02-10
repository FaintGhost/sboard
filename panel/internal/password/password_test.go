package password

import (
	"encoding/base64"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHashAndVerify(t *testing.T) {
	hash, err := Hash("pass123456")
	require.NoError(t, err)
	require.NotEmpty(t, hash)

	parts := strings.Split(hash, "$")
	require.Len(t, parts, 4)
	require.Equal(t, schemePBKDF2SHA256, parts[0])

	iter, err := strconv.Atoi(parts[1])
	require.NoError(t, err)
	require.Equal(t, defaultIter, iter)

	salt, err := base64.StdEncoding.DecodeString(parts[2])
	require.NoError(t, err)
	require.Len(t, salt, saltBytes)

	dk, err := base64.StdEncoding.DecodeString(parts[3])
	require.NoError(t, err)
	require.Len(t, dk, keyBytes)

	require.True(t, Verify(hash, "pass123456"))
	require.False(t, Verify(hash, "wrong-pass"))
}

func TestHash_EmptyPassword(t *testing.T) {
	_, err := Hash("")
	require.ErrorContains(t, err, "empty password")

	_, err = Hash("   ")
	require.ErrorContains(t, err, "empty password")
}

func TestVerify_MalformedHash(t *testing.T) {
	require.False(t, Verify("", "x"))
	require.False(t, Verify("abc", "x"))
	require.False(t, Verify("bcrypt$10$salt$hash", "x"))
	require.False(t, Verify("pbkdf2_sha256$bad$salt$hash", "x"))
	require.False(t, Verify("pbkdf2_sha256$1000$bad-base64$hash", "x"))
	require.False(t, Verify("pbkdf2_sha256$1000$c2FsdA==$bad-base64", "x"))
}

func TestPBKDF2SHA256_DeterministicAndLength(t *testing.T) {
	password := []byte("password")
	salt := []byte("salt")

	out1 := pbkdf2SHA256(password, salt, 2, 32)
	out2 := pbkdf2SHA256(password, salt, 2, 32)

	require.Equal(t, out1, out2)
	require.Len(t, out1, 32)

	short := pbkdf2SHA256(password, salt, 2, 20)
	require.Len(t, short, 20)

	longer := pbkdf2SHA256(password, salt, 2, 48)
	require.Len(t, longer, 48)
}
