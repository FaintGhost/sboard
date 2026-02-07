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
	b.Router().AppendTracker(tracker)

	return &Core{ctx: ctx, box: b, logFactory: lf, traffic: tracker}, nil
}

func (c *Core) Apply(inbounds []option.Inbound, raw []byte) error {
	loggerFactory := func(typ, tag string) log.ContextLogger {
		if c.logFactory == nil {
			return nil
		}
		return c.logFactory.NewLogger("inbound/" + typ + "[" + tag + "]")
	}
	if err := ApplyInbounds(c.ctx, c.box.Router(), loggerFactory, c.box.Inbound(), inbounds); err != nil {
		return err
	}
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
