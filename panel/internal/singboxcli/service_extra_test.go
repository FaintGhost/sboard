package singboxcli

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServiceFormat_ErrorBranches(t *testing.T) {
	service := New()

	_, err := service.Format(context.Background(), "")
	require.ErrorContains(t, err, "config is empty")

	_, err = service.Format(context.Background(), "{bad-json")
	require.ErrorContains(t, err, "invalid json")
}

func TestServiceCheck_ErrorBranches(t *testing.T) {
	service := New()

	_, err := service.Check(context.Background(), "")
	require.ErrorContains(t, err, "config is empty")

	_, err = service.Check(context.Background(), "{bad-json")
	require.Error(t, err)

	// Minimal empty config is accepted by current sing-box runtime path.
	ok, err := service.Check(context.Background(), `{}`)
	require.NoError(t, err)
	require.Equal(t, "ok", ok)
}

func TestServiceGenerate_MainBranches(t *testing.T) {
	service := New()

	t.Run("uuid", func(t *testing.T) {
		out, err := service.Generate(context.Background(), "uuid")
		require.NoError(t, err)
		require.Len(t, out, 36)
		require.Equal(t, 4, strings.Count(out, "-"))
	})

	t.Run("reality-keypair", func(t *testing.T) {
		out, err := service.Generate(context.Background(), "reality-keypair")
		require.NoError(t, err)
		require.Contains(t, out, "PrivateKey: ")
		require.Contains(t, out, "PublicKey: ")

		lines := strings.Split(out, "\n")
		require.Len(t, lines, 2)
		priv := strings.TrimPrefix(lines[0], "PrivateKey: ")
		pub := strings.TrimPrefix(lines[1], "PublicKey: ")
		_, err = base64.RawURLEncoding.DecodeString(priv)
		require.NoError(t, err)
		_, err = base64.RawURLEncoding.DecodeString(pub)
		require.NoError(t, err)
	})

	t.Run("wg-keypair", func(t *testing.T) {
		out, err := service.Generate(context.Background(), "wg-keypair")
		require.NoError(t, err)
		require.Contains(t, out, "PrivateKey: ")
		require.Contains(t, out, "PublicKey: ")

		lines := strings.Split(out, "\n")
		require.Len(t, lines, 2)
		priv := strings.TrimPrefix(lines[0], "PrivateKey: ")
		pub := strings.TrimPrefix(lines[1], "PublicKey: ")
		_, err = base64.StdEncoding.DecodeString(priv)
		require.NoError(t, err)
		_, err = base64.StdEncoding.DecodeString(pub)
		require.NoError(t, err)
	})

	t.Run("vapid-keypair", func(t *testing.T) {
		out, err := service.Generate(context.Background(), "vapid-keypair")
		require.NoError(t, err)
		require.Contains(t, out, "PrivateKey: ")
		require.Contains(t, out, "PublicKey: ")

		lines := strings.Split(out, "\n")
		require.Len(t, lines, 2)
		priv := strings.TrimPrefix(lines[0], "PrivateKey: ")
		pub := strings.TrimPrefix(lines[1], "PublicKey: ")
		_, err = base64.RawURLEncoding.DecodeString(priv)
		require.NoError(t, err)
		_, err = base64.RawURLEncoding.DecodeString(pub)
		require.NoError(t, err)
	})

	t.Run("invalid command", func(t *testing.T) {
		_, err := service.Generate(context.Background(), "not-supported")
		require.ErrorIs(t, err, ErrInvalidGenerateKind)
	})
}
