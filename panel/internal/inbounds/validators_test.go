package inbounds

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateSettings_Shadowsocks(t *testing.T) {
	t.Run("missing method", func(t *testing.T) {
		err := ValidateSettings("shadowsocks", map[string]any{})
		require.ErrorContains(t, err, "settings.method required")
	})

	t.Run("unsupported 2022 method", func(t *testing.T) {
		err := ValidateSettings("shadowsocks", map[string]any{
			"method": "2022-blake3-chacha20-poly1305",
		})
		require.ErrorContains(t, err, "does not support multi-user mode")
	})

	t.Run("supported 2022 methods", func(t *testing.T) {
		for _, method := range []string{"2022-blake3-aes-128-gcm", "2022-blake3-aes-256-gcm"} {
			err := ValidateSettings("shadowsocks", map[string]any{"method": method})
			require.NoError(t, err)
		}
	})

	t.Run("non-2022 method", func(t *testing.T) {
		err := ValidateSettings("shadowsocks", map[string]any{"method": "aes-128-gcm"})
		require.NoError(t, err)
	})
}

func TestValidateSettings_UnknownProtocolAndBlank(t *testing.T) {
	require.NoError(t, ValidateSettings("", map[string]any{"method": "x"}))
	require.NoError(t, ValidateSettings("not-exist", map[string]any{"method": "x"}))
}

func TestRegisterSettingsValidator_CustomProtocol(t *testing.T) {
	protocol := "unit-custom"
	wantErr := errors.New("custom validation failed")

	RegisterSettingsValidator(protocol, func(settings map[string]any) error {
		if settings["k"] != "v" {
			return wantErr
		}
		return nil
	})

	err := ValidateSettings("  UNIT-CUSTOM  ", map[string]any{"k": "x"})
	require.ErrorIs(t, err, wantErr)

	require.NoError(t, ValidateSettings("unit-custom", map[string]any{"k": "v"}))
}

func TestRegisterSettingsValidator_IgnoresInvalidInput(t *testing.T) {
	RegisterSettingsValidator("", func(map[string]any) error { return errors.New("x") })
	RegisterSettingsValidator("x", nil)

	// Should remain no-op because registration above is ignored.
	require.NoError(t, ValidateSettings("x", map[string]any{}))
}
