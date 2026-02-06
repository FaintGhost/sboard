package sync

import (
  "context"
  "encoding/json"
  "fmt"

  sbjson "github.com/sagernet/sing/common/json"
  "github.com/sagernet/sing-box/include"
  "github.com/sagernet/sing-box/option"
)

type inboundMeta struct {
  Tag        string `json:"tag"`
  Type       string `json:"type"`
  Listen     string `json:"listen"`
  ListenPort int    `json:"listen_port"`
}

type syncRequest struct {
  Inbounds []json.RawMessage `json:"inbounds"`
}

func NewSingboxContext() context.Context {
  return include.Context(context.Background())
}

func ParseAndValidateInbounds(ctx context.Context, body []byte) ([]option.Inbound, error) {
  var req syncRequest
  if err := json.Unmarshal(body, &req); err != nil {
    return nil, err
  }

  inbounds := make([]option.Inbound, 0, len(req.Inbounds))
  seen := map[string]struct{}{}
  for i, raw := range req.Inbounds {
    var meta inboundMeta
    if err := json.Unmarshal(raw, &meta); err != nil {
      return nil, err
    }
    if meta.Tag == "" {
      return nil, fmt.Errorf("inbounds[%d].tag required", i)
    }
    if meta.Type == "" {
      return nil, fmt.Errorf("inbounds[%d].type required", i)
    }
    if meta.ListenPort <= 0 || meta.ListenPort > 65535 {
      return nil, fmt.Errorf("inbounds[%d].listen_port invalid", i)
    }
    if _, ok := seen[meta.Tag]; ok {
      return nil, fmt.Errorf("inbounds[%d].tag duplicated", i)
    }
    seen[meta.Tag] = struct{}{}

    var inb option.Inbound
    if err := sbjson.UnmarshalContext(ctx, raw, &inb); err != nil {
      // sing-box errors can be too context-free; include index/tag/type for debugging.
      return nil, fmt.Errorf("inbounds[%d] (tag=%s type=%s): %w", i, meta.Tag, meta.Type, err)
    }
    inbounds = append(inbounds, inb)
  }
  return inbounds, nil
}
