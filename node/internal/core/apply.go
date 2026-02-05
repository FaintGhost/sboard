package core

import (
  "context"

  "github.com/sagernet/sing-box/adapter"
  "github.com/sagernet/sing-box/log"
  "github.com/sagernet/sing-box/option"
)

type InboundCreator interface {
  Create(ctx context.Context, router adapter.Router, logger log.ContextLogger, tag, inboundType string, options any) error
}

type LoggerFactory func(typ, tag string) log.ContextLogger

func ApplyInbounds(ctx context.Context, router adapter.Router, loggerFactory LoggerFactory, mgr InboundCreator, inbounds []option.Inbound) error {
  for i := range inbounds {
    inb := inbounds[i]
    var lg log.ContextLogger
    if loggerFactory != nil {
      lg = loggerFactory(inb.Type, inb.Tag)
    }
    if err := mgr.Create(ctx, router, lg, inb.Tag, inb.Type, inb.Options); err != nil {
      return err
    }
  }
  return nil
}
