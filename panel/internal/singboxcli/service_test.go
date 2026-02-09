package singboxcli

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServiceFormat_KeepEmptyUsers(t *testing.T) {
	service := New()

	input := `{
  "type": "shadowsocks",
  "tag": "ss-in",
  "listen_port": 8388,
  "method": "2022-blake3-aes-128-gcm",
  "password": "8JCsPssfgS8tiRwiMlhARg==",
  "users": []
}`

	formatted, err := service.Format(context.Background(), input)
	require.NoError(t, err)
	require.Contains(t, formatted, `"users": []`)
	require.True(t, strings.Contains(formatted, `"type": "shadowsocks"`))
}

func TestServiceGenerate_RandBase64Length(t *testing.T) {
	service := New()

	out16, err := service.Generate(context.Background(), "rand-base64-16")
	require.NoError(t, err)
	raw16, err := base64.StdEncoding.DecodeString(out16)
	require.NoError(t, err)
	require.Len(t, raw16, 16)

	out32, err := service.Generate(context.Background(), "rand-base64-32")
	require.NoError(t, err)
	raw32, err := base64.StdEncoding.DecodeString(out32)
	require.NoError(t, err)
	require.Len(t, raw32, 32)
}

func TestServiceGenerate_RandBase64Invalid(t *testing.T) {
	service := New()

	_, err := service.Generate(context.Background(), "rand-base64-0")
	require.ErrorIs(t, err, ErrInvalidGenerateKind)

	_, err = service.Generate(context.Background(), "rand-base64-abc")
	require.ErrorIs(t, err, ErrInvalidGenerateKind)
}
