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

func TestValidateSettings_VLESS(t *testing.T) {
	t.Run("valid users", func(t *testing.T) {
		err := ValidateSettings("vless", map[string]any{
			"users": []any{
				map[string]any{"uuid": "abcd-uuid", "flow": "xtls-rprx-vision"},
			},
		})
		require.NoError(t, err)
	})

	t.Run("valid users without flow", func(t *testing.T) {
		err := ValidateSettings("vless", map[string]any{
			"users": []any{
				map[string]any{"uuid": "abcd-uuid"},
			},
		})
		require.NoError(t, err)
	})

	t.Run("missing users", func(t *testing.T) {
		err := ValidateSettings("vless", map[string]any{})
		require.ErrorContains(t, err, "users is required")
	})

	t.Run("empty users", func(t *testing.T) {
		err := ValidateSettings("vless", map[string]any{"users": []any{}})
		require.ErrorContains(t, err, "must not be empty")
	})

	t.Run("invalid flow", func(t *testing.T) {
		err := ValidateSettings("vless", map[string]any{
			"users": []any{
				map[string]any{"uuid": "abcd-uuid", "flow": "xtls-rprx-vless"},
			},
		})
		require.ErrorContains(t, err, "flow must be empty or \"xtls-rprx-vision\"")
	})

	t.Run("missing uuid", func(t *testing.T) {
		err := ValidateSettings("vless", map[string]any{
			"users": []any{
				map[string]any{"name": "user"},
			},
		})
		require.ErrorContains(t, err, ".uuid is required")
	})
}

func TestValidateSettings_VMESS(t *testing.T) {
	t.Run("valid users", func(t *testing.T) {
		err := ValidateSettings("vmess", map[string]any{
			"users": []any{
				map[string]any{"uuid": "abcd-uuid", "alterId": 0.0},
			},
		})
		require.NoError(t, err)
	})

	t.Run("missing users", func(t *testing.T) {
		err := ValidateSettings("vmess", map[string]any{})
		require.ErrorContains(t, err, "users is required")
	})

	t.Run("empty users", func(t *testing.T) {
		err := ValidateSettings("vmess", map[string]any{"users": []any{}})
		require.ErrorContains(t, err, "must not be empty")
	})

	t.Run("negative alterId", func(t *testing.T) {
		err := ValidateSettings("vmess", map[string]any{
			"users": []any{
				map[string]any{"uuid": "abcd-uuid", "alterId": -1.0},
			},
		})
		require.ErrorContains(t, err, "alterId must be >= 0")
	})

	t.Run("missing uuid", func(t *testing.T) {
		err := ValidateSettings("vmess", map[string]any{
			"users": []any{
				map[string]any{"name": "user"},
			},
		})
		require.ErrorContains(t, err, ".uuid is required")
	})
}

func TestValidateSettings_Trojan(t *testing.T) {
	t.Run("valid users", func(t *testing.T) {
		err := ValidateSettings("trojan", map[string]any{
			"users": []any{
				map[string]any{"password": "secret"},
			},
		})
		require.NoError(t, err)
	})

	t.Run("missing users", func(t *testing.T) {
		err := ValidateSettings("trojan", map[string]any{})
		require.ErrorContains(t, err, "users is required")
	})

	t.Run("empty users", func(t *testing.T) {
		err := ValidateSettings("trojan", map[string]any{"users": []any{}})
		require.ErrorContains(t, err, "must not be empty")
	})

	t.Run("missing password", func(t *testing.T) {
		err := ValidateSettings("trojan", map[string]any{
			"users": []any{
				map[string]any{"name": "user"},
			},
		})
		require.ErrorContains(t, err, ".password is required")
	})
}

func TestValidateSettings_SOCKS(t *testing.T) {
	t.Run("valid users with auth", func(t *testing.T) {
		err := ValidateSettings("socks", map[string]any{
			"users": []any{
				map[string]any{"username": "user", "password": "pass"},
			},
		})
		require.NoError(t, err)
	})

	t.Run("no users allowed", func(t *testing.T) {
		err := ValidateSettings("socks", map[string]any{})
		require.NoError(t, err)
	})

	t.Run("missing password in user", func(t *testing.T) {
		err := ValidateSettings("socks", map[string]any{
			"users": []any{
				map[string]any{"username": "user"},
			},
		})
		require.ErrorContains(t, err, ".password is required")
	})
}

func TestValidateSettings_Hysteria2(t *testing.T) {
	t.Run("valid users", func(t *testing.T) {
		err := ValidateSettings("hysteria2", map[string]any{
			"users": []any{
				map[string]any{"name": "user", "password": "secret"},
			},
		})
		require.NoError(t, err)
	})

	t.Run("missing users", func(t *testing.T) {
		err := ValidateSettings("hysteria2", map[string]any{})
		require.ErrorContains(t, err, "users is required")
	})

	t.Run("missing password", func(t *testing.T) {
		err := ValidateSettings("hysteria2", map[string]any{
			"users": []any{
				map[string]any{"name": "user"},
			},
		})
		require.ErrorContains(t, err, ".password is required")
	})
}

func TestValidateSettings_TUIC(t *testing.T) {
	t.Run("valid users", func(t *testing.T) {
		err := ValidateSettings("tuic", map[string]any{
			"users": []any{
				map[string]any{"uuid": "abcd-uuid", "password": "secret"},
			},
		})
		require.NoError(t, err)
	})

	t.Run("missing users", func(t *testing.T) {
		err := ValidateSettings("tuic", map[string]any{})
		require.ErrorContains(t, err, "users is required")
	})

	t.Run("missing uuid", func(t *testing.T) {
		err := ValidateSettings("tuic", map[string]any{
			"users": []any{
				map[string]any{"password": "secret"},
			},
		})
		require.ErrorContains(t, err, ".uuid is required")
	})

	t.Run("missing password", func(t *testing.T) {
		err := ValidateSettings("tuic", map[string]any{
			"users": []any{
				map[string]any{"uuid": "abcd-uuid"},
			},
		})
		require.ErrorContains(t, err, ".password is required")
	})
}

func TestValidateSettings_Naive(t *testing.T) {
	t.Run("valid users", func(t *testing.T) {
		err := ValidateSettings("naive", map[string]any{
			"users": []any{
				map[string]any{"username": "user", "password": "secret"},
			},
		})
		require.NoError(t, err)
	})

	t.Run("missing users", func(t *testing.T) {
		err := ValidateSettings("naive", map[string]any{})
		require.ErrorContains(t, err, "users is required")
	})

	t.Run("missing username", func(t *testing.T) {
		err := ValidateSettings("naive", map[string]any{
			"users": []any{
				map[string]any{"password": "secret"},
			},
		})
		require.ErrorContains(t, err, ".username is required")
	})

	t.Run("missing password", func(t *testing.T) {
		err := ValidateSettings("naive", map[string]any{
			"users": []any{
				map[string]any{"username": "user"},
			},
		})
		require.ErrorContains(t, err, ".password is required")
	})
}

func TestValidateSettings_ShadowTLS(t *testing.T) {
	t.Run("valid users and handshake", func(t *testing.T) {
		err := ValidateSettings("shadowtls", map[string]any{
			"users": []any{
				map[string]any{"name": "user", "password": "user-password"},
			},
			"handshake": map[string]any{
				"server": "google.com",
				"server_port": 443,
			},
		})
		require.NoError(t, err)
	})

	t.Run("missing users", func(t *testing.T) {
		err := ValidateSettings("shadowtls", map[string]any{
			"handshake": map[string]any{"server": "google.com", "server_port": 443},
		})
		require.ErrorContains(t, err, "users is required")
	})

	t.Run("missing handshake", func(t *testing.T) {
		err := ValidateSettings("shadowtls", map[string]any{
			"users": []any{
				map[string]any{"name": "user", "password": "user-password"},
			},
		})
		require.ErrorContains(t, err, "handshake is required")
	})

	t.Run("missing handshake server", func(t *testing.T) {
		err := ValidateSettings("shadowtls", map[string]any{
			"users": []any{
				map[string]any{"name": "user", "password": "user-password"},
			},
			"handshake": map[string]any{
				"server_port": 443,
			},
		})
		require.ErrorContains(t, err, "handshake.server is required")
	})
}

func TestValidateSettings_AnyTLS(t *testing.T) {
	t.Run("valid users", func(t *testing.T) {
		err := ValidateSettings("anytls", map[string]any{
			"users": []any{
				map[string]any{"name": "user", "password": "secret"},
			},
		})
		require.NoError(t, err)
	})

	t.Run("missing users", func(t *testing.T) {
		err := ValidateSettings("anytls", map[string]any{})
		require.ErrorContains(t, err, "users is required")
	})

	t.Run("missing password", func(t *testing.T) {
		err := ValidateSettings("anytls", map[string]any{
			"users": []any{
				map[string]any{"name": "user"},
			},
		})
		require.ErrorContains(t, err, ".password is required")
	})
}

func TestValidateSettings_HTTPAndMixed(t *testing.T) {
	// http and mixed share the socks validator
	for _, proto := range []string{"http", "mixed"} {
		t.Run(proto+"_valid users", func(t *testing.T) {
			err := ValidateSettings(proto, map[string]any{
				"users": []any{
					map[string]any{"username": "user", "password": "pass"},
				},
			})
			require.NoError(t, err)
		})
		t.Run(proto+"_no users allowed", func(t *testing.T) {
			err := ValidateSettings(proto, map[string]any{})
			require.NoError(t, err)
		})
	}
}
