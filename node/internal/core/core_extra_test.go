package core

import (
	"context"
	"errors"
	"testing"
	"time"

	sbbox "github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/option"
	"sboard/node/internal/stats"
)

type testBox struct {
	startErr error
	closeErr error
	started  bool
	closed   bool
}

func (b *testBox) Start() error {
	b.started = true
	return b.startErr
}

func (b *testBox) Inbound() adapter.InboundManager { return nil }

func (b *testBox) Router() adapter.Router { return nil }

func (b *testBox) Close() error {
	b.closed = true
	return b.closeErr
}

func TestCoreApplyAndApplyOptions(t *testing.T) {
	oldFactory := NewBox
	t.Cleanup(func() { NewBox = oldFactory })

	newCreated := &testBox{}
	NewBox = func(opts sbbox.Options) (Box, error) {
		if len(opts.Options.Inbounds) != 1 || opts.Options.Inbounds[0].Tag != "in-1" {
			t.Fatalf("unexpected inbounds passed to NewBox: %+v", opts.Options.Inbounds)
		}
		return newCreated, nil
	}

	old := &testBox{}
	c := &Core{ctx: context.Background(), box: old}
	raw := []byte(`{"inbounds":[{"tag":"in-1"}]}`)

	err := c.Apply([]option.Inbound{{Type: "mixed", Tag: "in-1"}}, raw)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}
	if !old.closed {
		t.Fatal("old box should be closed")
	}
	if !newCreated.started {
		t.Fatal("new box should be started")
	}
	if c.hash == "" {
		t.Fatal("config hash should be set")
	}
	if c.at.IsZero() || time.Since(c.at) > time.Minute {
		t.Fatalf("apply time should be updated, got %v", c.at)
	}
}

func TestCoreApplyOptionsFactoryAndStartErrors(t *testing.T) {
	oldFactory := NewBox
	t.Cleanup(func() { NewBox = oldFactory })

	c := &Core{ctx: context.Background(), box: &testBox{}}

	NewBox = func(opts sbbox.Options) (Box, error) {
		return nil, errors.New("new box failed")
	}
	if err := c.ApplyOptions(option.Options{}, []byte(`{}`)); err == nil {
		t.Fatal("expected NewBox error")
	}

	created := &testBox{startErr: errors.New("start failed")}
	NewBox = func(opts sbbox.Options) (Box, error) {
		return created, nil
	}
	if err := c.ApplyOptions(option.Options{}, []byte(`{}`)); err == nil {
		t.Fatal("expected start error")
	}
	if !created.closed {
		t.Fatal("failed-start box should be closed")
	}
}

func TestCoreInboundTrafficMetaAndClose(t *testing.T) {
	var nilCore *Core
	if got := nilCore.InboundTraffic(false); got != nil {
		t.Fatalf("nil core InboundTraffic should return nil, got %+v", got)
	}
	if meta := nilCore.InboundTrafficMeta(); meta != (stats.InboundTrafficMeta{}) {
		t.Fatalf("nil core meta should be zero, got %+v", meta)
	}
	if err := nilCore.Close(); err != nil {
		t.Fatalf("nil core close should be nil, got %v", err)
	}

	c := &Core{traffic: stats.NewInboundTrafficTracker()}
	if got := c.InboundTraffic(false); len(got) != 0 {
		t.Fatalf("empty tracker should return empty slice, got %+v", got)
	}
	meta := c.InboundTrafficMeta()
	if meta.TrackedTags != 0 || meta.TCPConns != 0 || meta.UDPConns != 0 {
		t.Fatalf("unexpected meta: %+v", meta)
	}

	c.box = &testBox{closeErr: errors.New("close failed")}
	if err := c.Close(); err == nil {
		t.Fatal("expected close error")
	}

	c2 := &Core{}
	if err := c2.Close(); err != nil {
		t.Fatalf("core with nil box close should be nil, got %v", err)
	}
}
