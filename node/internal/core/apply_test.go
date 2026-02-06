package core_test

import (
  "context"
  "testing"

  "sboard/node/internal/core"
  "github.com/sagernet/sing-box/adapter"
  "github.com/sagernet/sing-box/log"
  "github.com/sagernet/sing-box/option"
  "github.com/stretchr/testify/require"
)

type fakeInboundManager struct {
  calls int
  err   error
}

func (f *fakeInboundManager) Create(ctx context.Context, router adapter.Router, logger log.ContextLogger, tag, inboundType string, options any) error {
  f.calls++
  return f.err
}

func TestApplyInbounds(t *testing.T) {
  mgr := &fakeInboundManager{}
  inbounds := []option.Inbound{{Type: "mixed", Tag: "m1", Options: struct{}{}}}
  err := core.ApplyInbounds(context.Background(), nil, nil, mgr, inbounds)
  require.NoError(t, err)
  require.Equal(t, 1, mgr.calls)
}

func TestApplyInbounds_WrapsCreateErrorWithMeta(t *testing.T) {
  mgr := &fakeInboundManager{err: context.Canceled}
  inbounds := []option.Inbound{{Type: "shadowsocks", Tag: "ss-in", Options: struct{}{}}}
  err := core.ApplyInbounds(context.Background(), nil, nil, mgr, inbounds)
  require.Error(t, err)
  require.Contains(t, err.Error(), "tag=ss-in")
  require.Contains(t, err.Error(), "type=shadowsocks")
}
