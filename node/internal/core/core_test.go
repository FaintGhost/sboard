package core_test

import (
	"context"
	"errors"
	"testing"

	sbbox "github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/option"
	"github.com/stretchr/testify/require"
	"sboard/node/internal/core"
)

type fakeBox struct{ started bool }

func (f *fakeBox) Start() error                    { f.started = true; return nil }
func (f *fakeBox) Inbound() adapter.InboundManager { return nil }
func (f *fakeBox) Router() adapter.Router          { return nil }
func (f *fakeBox) Close() error                    { return nil }

func TestNewCoreUsesNewBox(t *testing.T) {
	old := core.NewBox
	t.Cleanup(func() { core.NewBox = old })

	var got sbbox.Options
	core.NewBox = func(opts sbbox.Options) (core.Box, error) {
		got = opts
		return &fakeBox{}, nil
	}

	c, err := core.New(context.Background(), "info")
	require.NoError(t, err)
	require.NotNil(t, c)
	require.True(t, got.Options.Route == nil || got.Options.Route.Final == "")
	if got.Options.DNS != nil {
		require.Empty(t, got.Options.DNS.RawDNSOptions.Final)
	}
}

type fakeClosableBox struct {
	startErr error
	startFn  func() error
	closed   bool
}

func (f *fakeClosableBox) Start() error {
	if f.startFn != nil {
		return f.startFn()
	}
	return f.startErr
}
func (f *fakeClosableBox) Inbound() adapter.InboundManager { return nil }
func (f *fakeClosableBox) Router() adapter.Router          { return nil }
func (f *fakeClosableBox) Close() error {
	f.closed = true
	return nil
}

func TestApplyOptions_ClosesOldBoxBeforeStartingNewOne(t *testing.T) {
	oldFactory := core.NewBox
	t.Cleanup(func() { core.NewBox = oldFactory })

	first := &fakeClosableBox{}
	second := &fakeClosableBox{}
	call := 0

	core.NewBox = func(opts sbbox.Options) (core.Box, error) {
		call++
		if call == 1 {
			return first, nil
		}
		require.Equal(t, option.Options{Inbounds: []option.Inbound{{Tag: "ss-in", Type: "shadowsocks"}}}, opts.Options)
		second.startFn = func() error {
			if !first.closed {
				return errors.New("old box still open")
			}
			return nil
		}
		return second, nil
	}

	c, err := core.New(context.Background(), "info")
	require.NoError(t, err)

	err = c.ApplyOptions(option.Options{Inbounds: []option.Inbound{{Tag: "ss-in", Type: "shadowsocks"}}}, []byte(`{"inbounds":[]}`))
	require.NoError(t, err)
	require.True(t, first.closed)
}
