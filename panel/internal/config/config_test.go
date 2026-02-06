package config_test

import (
  "testing"

  "sboard/panel/internal/config"
  "github.com/stretchr/testify/require"
)

func TestValidateConfig(t *testing.T) {
  cfg := config.Config{}
  err := config.Validate(cfg)
  require.Error(t, err)

  cfg = config.Config{JWTSecret: "secret"}
  require.NoError(t, config.Validate(cfg))
}
