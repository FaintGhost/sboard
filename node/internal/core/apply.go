package core

import (
  "context"
  "fmt"

  "github.com/sagernet/sing-box/adapter"
  "github.com/sagernet/sing-box/log"
  "github.com/sagernet/sing-box/option"
)

type InboundManager interface {
  Inbounds() []adapter.Inbound
  Remove(tag string) error
  Create(ctx context.Context, router adapter.Router, logger log.ContextLogger, tag, inboundType string, options any) error
}

type LoggerFactory func(typ, tag string) log.ContextLogger

func ApplyInbounds(ctx context.Context, router adapter.Router, loggerFactory LoggerFactory, mgr InboundManager, inbounds []option.Inbound) error {
  // Build a set of tags we want to keep.
  wantTags := make(map[string]struct{}, len(inbounds))
  for _, inb := range inbounds {
    wantTags[inb.Tag] = struct{}{}
  }

  // Remove inbounds that are no longer in the config or will be recreated.
  // We must remove first to free ports before creating new inbounds.
  for _, existing := range mgr.Inbounds() {
    if err := mgr.Remove(existing.Tag()); err != nil {
      return fmt.Errorf("remove inbound (tag=%s): %w", existing.Tag(), err)
    }
  }

  // Create all inbounds from scratch.
  for i := range inbounds {
    inb := inbounds[i]
    var lg log.ContextLogger
    if loggerFactory != nil {
      lg = loggerFactory(inb.Type, inb.Tag)
    }
    if err := mgr.Create(ctx, router, lg, inb.Tag, inb.Type, inb.Options); err != nil {
      // sing-box errors are often context-free; add tag/type to speed up debugging.
      return fmt.Errorf("create inbound (tag=%s type=%s): %w", inb.Tag, inb.Type, err)
    }
  }
  return nil
}
