package sync

import (
  "context"
  "encoding/json"
  "errors"
  "fmt"
  "strings"

  sbjson "github.com/sagernet/sing/common/json"
  "github.com/sagernet/sing-box/include"
  "github.com/sagernet/sing-box/option"
)

// BadRequestError indicates the client sent an invalid sync payload.
// The API layer maps this error to HTTP 400.
type BadRequestError struct {
  Message string
}

func (e BadRequestError) Error() string { return e.Message }

type inboundMeta struct {
  Tag        string `json:"tag"`
  Type       string `json:"type"`
  Listen     string `json:"listen"`
  ListenPort int    `json:"listen_port"`
  Method     string `json:"method"`
  Password   string `json:"password"`
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
    return nil, BadRequestError{Message: "invalid json"}
  }

  inbounds := make([]option.Inbound, 0, len(req.Inbounds))
  seen := map[string]struct{}{}
  for i, raw := range req.Inbounds {
    var meta inboundMeta
    if err := json.Unmarshal(raw, &meta); err != nil {
      return nil, BadRequestError{Message: fmt.Sprintf("inbounds[%d] invalid json", i)}
    }
    if meta.Tag == "" {
      return nil, BadRequestError{Message: fmt.Sprintf("inbounds[%d].tag required", i)}
    }
    if meta.Type == "" {
      return nil, BadRequestError{Message: fmt.Sprintf("inbounds[%d].type required", i)}
    }
    if meta.ListenPort <= 0 || meta.ListenPort > 65535 {
      return nil, BadRequestError{Message: fmt.Sprintf("inbounds[%d].listen_port invalid", i)}
    }
    if _, ok := seen[meta.Tag]; ok {
      return nil, BadRequestError{Message: fmt.Sprintf("inbounds[%d].tag duplicated", i)}
    }
    seen[meta.Tag] = struct{}{}

    if strings.TrimSpace(strings.ToLower(meta.Type)) == "shadowsocks" &&
      strings.HasPrefix(strings.TrimSpace(meta.Method), "2022-") &&
      strings.TrimSpace(meta.Password) == "" {
      return nil, BadRequestError{Message: fmt.Sprintf("inbounds[%d] (tag=%s type=%s): password required for method %s", i, meta.Tag, meta.Type, meta.Method)}
    }

    var inb option.Inbound
    if err := sbjson.UnmarshalContext(ctx, raw, &inb); err != nil {
      // sing-box errors can be too context-free; include index/tag/type for debugging.
      // Treat unmarshal/type errors as a client-side config issue (HTTP 400).
      if errors.Is(err, context.Canceled) {
        return nil, err
      }
      return nil, BadRequestError{Message: fmt.Sprintf("inbounds[%d] (tag=%s type=%s): %v", i, meta.Tag, meta.Type, err)}
    }
    inbounds = append(inbounds, inb)
  }
  return inbounds, nil
}
