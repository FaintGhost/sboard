package core

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	sbbox "github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/include"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"

	"sboard/node/internal/stats"
)

type Box interface {
	Start() error
	Inbound() adapter.InboundManager
	Router() adapter.Router
	Close() error
}

var NewBox = func(opts sbbox.Options) (Box, error) { return sbbox.New(opts) }

type Core struct {
	ctx        context.Context
	box        Box
	logFactory log.Factory
	hash       string
	at         time.Time

	traffic *stats.InboundTrafficTracker
}

func New(ctx context.Context, logLevel string) (*Core, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx = include.Context(ctx)

	opts := option.Options{
		Log: &option.LogOptions{Level: logLevel},
	}

	lf, err := log.New(log.Options{Context: ctx, Options: *opts.Log})
	if err != nil {
		return nil, err
	}

	b, err := NewBox(sbbox.Options{Options: opts, Context: ctx})
	if err != nil {
		return nil, err
	}
	if err := b.Start(); err != nil {
		return nil, err
	}

	// Track inbound-level traffic via router tracker. This is independent from sync/persistence.
	tracker := stats.NewInboundTrafficTracker()
	if router := b.Router(); router != nil {
		router.AppendTracker(tracker)
	}

	return &Core{ctx: ctx, box: b, logFactory: lf, traffic: tracker}, nil
}

func (c *Core) Apply(inbounds []option.Inbound, raw []byte) error {
	return c.ApplyOptions(option.Options{Inbounds: inbounds}, raw)
}

func (c *Core) ApplyOptions(options option.Options, raw []byte) error {
	newBox, err := NewBox(sbbox.Options{Options: options, Context: c.ctx})
	if err != nil {
		return err
	}

	oldBox := c.box
	if oldBox != nil {
		_ = oldBox.Close()
	}

	if err := newBox.Start(); err != nil {
		_ = newBox.Close()
		return err
	}

	if c.traffic == nil {
		c.traffic = stats.NewInboundTrafficTracker()
	}
	if router := newBox.Router(); router != nil {
		router.AppendTracker(c.traffic)
	}
	c.box = newBox

	sum := sha256.Sum256(raw)
	c.hash = hex.EncodeToString(sum[:])
	c.at = time.Now()
	return nil
}

func (c *Core) InboundTraffic(reset bool) []stats.InboundTraffic {
	if c == nil || c.traffic == nil {
		return nil
	}
	return c.traffic.Snapshot(reset)
}

func (c *Core) InboundTrafficMeta() stats.InboundTrafficMeta {
	if c == nil || c.traffic == nil {
		return stats.InboundTrafficMeta{}
	}
	return c.traffic.Meta()
}

// Close stops the sing-box instance and releases resources.
func (c *Core) Close() error {
	if c == nil || c.box == nil {
		return nil
	}
	return c.box.Close()
}
