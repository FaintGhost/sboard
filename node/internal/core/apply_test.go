package core_test

import (
  "context"
  "testing"

  "github.com/FaintGhost/sboard/node/internal/core"
  "github.com/sagernet/sing-box/adapter"
  "github.com/sagernet/sing-box/log"
  "github.com/sagernet/sing-box/option"
  "github.com/stretchr/testify/require"
)

type fakeInboundManager struct {
  calls int
}

func (f *fakeInboundManager) Create(ctx context.Context, router adapter.Router, logger log.ContextLogger, tag, inboundType string, options any) error {
  f.calls++
  return nil
}

func TestApplyInbounds(t *testing.T) {
  mgr := &fakeInboundManager{}
  inbounds := []option.Inbound{{Type: "mixed", Tag: "m1", Options: struct{}{}}}
  err := core.ApplyInbounds(context.Background(), nil, nil, mgr, inbounds)
  require.NoError(t, err)
  require.Equal(t, 1, mgr.calls)
}
