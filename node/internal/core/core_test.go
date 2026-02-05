package core_test

import (
  "context"
  "testing"

  "sboard/node/internal/core"
  "github.com/sagernet/sing-box/adapter"
  sbbox "github.com/sagernet/sing-box"
  "github.com/stretchr/testify/require"
)

type fakeBox struct{ started bool }

func (f *fakeBox) Start() error { f.started = true; return nil }
func (f *fakeBox) Inbound() adapter.InboundManager { return nil }
func (f *fakeBox) Router() adapter.Router { return nil }
func (f *fakeBox) Close() error { return nil }

func TestNewCoreUsesNewBox(t *testing.T) {
  old := core.NewBox
  t.Cleanup(func() { core.NewBox = old })

  core.NewBox = func(opts sbbox.Options) (core.Box, error) {
    return &fakeBox{}, nil
  }

  c, err := core.New(context.Background(), "info")
  require.NoError(t, err)
  require.NotNil(t, c)
}
