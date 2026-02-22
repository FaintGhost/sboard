package api

import (
  "context"
  "encoding/json"
  "errors"
  "strings"

  "sboard/panel/internal/db"
  inbval "sboard/panel/internal/inbounds"
)

func (s *Server) ListInbounds(ctx context.Context, request ListInboundsRequestObject) (ListInboundsResponseObject, error) {
  limit, offset, err := paginationDefaults(request.Params.Limit, request.Params.Offset)
  if err != nil {
    return ListInbounds400JSONResponse{errBadRequest("invalid pagination")}, nil
  }

  var nodeID int64
  if request.Params.NodeId != nil {
    nodeID = *request.Params.NodeId
  }

  items, err := s.store.ListInbounds(ctx, limit, offset, nodeID)
  if err != nil {
    return ListInbounds500JSONResponse{errInternal("list inbounds failed")}, nil
  }

  out := make([]Inbound, 0, len(items))
  for _, inb := range items {
    out = append(out, dbInboundToAPI(inb))
  }
  return ListInbounds200JSONResponse{Data: out}, nil
}

func (s *Server) CreateInbound(ctx context.Context, request CreateInboundRequestObject) (CreateInboundResponseObject, error) {
  tag := strings.TrimSpace(request.Body.Tag)
  proto := strings.TrimSpace(request.Body.Protocol)
  if tag == "" || proto == "" || request.Body.NodeId <= 0 || request.Body.ListenPort <= 0 {
    return CreateInbound400JSONResponse{errBadRequest("invalid inbound")}, nil
  }

  settingsBytes, err := json.Marshal(request.Body.Settings)
  if err != nil || len(settingsBytes) == 0 {
    return CreateInbound400JSONResponse{errBadRequest("invalid settings")}, nil
  }

  if err := inbval.ValidateSettings(proto, request.Body.Settings); err != nil {
    return CreateInbound400JSONResponse{errBadRequest(err.Error())}, nil
  }

  tlsBytes := marshalOptionalMap(request.Body.TlsSettings)
  transportBytes := marshalOptionalMap(request.Body.TransportSettings)

  publicPort := request.Body.ListenPort
  if request.Body.PublicPort != nil {
    publicPort = *request.Body.PublicPort
  }

  inb, err := s.store.CreateInbound(ctx, db.InboundCreate{
    Tag:               tag,
    NodeID:            request.Body.NodeId,
    Protocol:          proto,
    ListenPort:        request.Body.ListenPort,
    PublicPort:        publicPort,
    Settings:          settingsBytes,
    TLSSettings:       tlsBytes,
    TransportSettings: transportBytes,
  })
  if err != nil {
    if errors.Is(err, db.ErrConflict) {
      return CreateInbound409JSONResponse{errConflict("conflict")}, nil
    }
    if errors.Is(err, db.ErrNotFound) {
      return CreateInbound404JSONResponse{errNotFound("node not found")}, nil
    }
    return CreateInbound500JSONResponse{errInternal("create inbound failed")}, nil
  }

  n, err := s.store.GetNodeByID(ctx, inb.NodeID)
  if err != nil {
    return CreateInbound201JSONResponse{
      Data: dbInboundToAPI(inb),
      Sync: SyncResult{Status: "error", Error: strPtr("get node failed")},
    }, nil
  }
  sync := trySyncNodeWithSource(ctx, s.store, n, triggerSourceInbound)
  return CreateInbound201JSONResponse{
    Data: dbInboundToAPI(inb),
    Sync: syncResultToAPI(sync),
  }, nil
}

func (s *Server) GetInbound(ctx context.Context, request GetInboundRequestObject) (GetInboundResponseObject, error) {
  inb, err := s.store.GetInboundByID(ctx, request.Id)
  if err != nil {
    if errors.Is(err, db.ErrNotFound) {
      return GetInbound404JSONResponse{errNotFound("inbound not found")}, nil
    }
    return GetInbound500JSONResponse{errInternal("get inbound failed")}, nil
  }
  return GetInbound200JSONResponse{Data: dbInboundToAPI(inb)}, nil
}

func (s *Server) UpdateInbound(ctx context.Context, request UpdateInboundRequestObject) (UpdateInboundResponseObject, error) {
  cur, err := s.store.GetInboundByID(ctx, request.Id)
  if err != nil {
    if errors.Is(err, db.ErrNotFound) {
      return UpdateInbound404JSONResponse{errNotFound("inbound not found")}, nil
    }
    return UpdateInbound500JSONResponse{errInternal("get inbound failed")}, nil
  }

  upd := db.InboundUpdate{}
  if request.Body.Tag != nil {
    tag := strings.TrimSpace(*request.Body.Tag)
    if tag == "" {
      return UpdateInbound400JSONResponse{errBadRequest("invalid tag")}, nil
    }
    upd.Tag = &tag
  }
  if request.Body.Protocol != nil {
    p := strings.TrimSpace(*request.Body.Protocol)
    if p == "" {
      return UpdateInbound400JSONResponse{errBadRequest("invalid protocol")}, nil
    }
    upd.Protocol = &p
  }
  if request.Body.ListenPort != nil {
    if *request.Body.ListenPort <= 0 {
      return UpdateInbound400JSONResponse{errBadRequest("invalid listen_port")}, nil
    }
    upd.ListenPort = request.Body.ListenPort
  }
  if request.Body.PublicPort != nil {
    if *request.Body.PublicPort < 0 {
      return UpdateInbound400JSONResponse{errBadRequest("invalid public_port")}, nil
    }
    upd.PublicPort = request.Body.PublicPort
  }
  if request.Body.Settings != nil {
    b, _ := json.Marshal(*request.Body.Settings)
    if len(b) == 0 || !json.Valid(b) {
      return UpdateInbound400JSONResponse{errBadRequest("invalid settings")}, nil
    }
    raw := json.RawMessage(b)
    upd.Settings = &raw
  }
  if request.Body.TlsSettings != nil {
    b, _ := json.Marshal(*request.Body.TlsSettings)
    raw := json.RawMessage(b)
    upd.TLSSettings = &raw
  }
  if request.Body.TransportSettings != nil {
    b, _ := json.Marshal(*request.Body.TransportSettings)
    raw := json.RawMessage(b)
    upd.TransportSettings = &raw
  }

  // Validate final protocol+settings.
  finalProto := cur.Protocol
  if upd.Protocol != nil {
    finalProto = *upd.Protocol
  }
  finalSettings := cur.Settings
  if upd.Settings != nil {
    finalSettings = *upd.Settings
  }
  if len(finalSettings) == 0 || !json.Valid(finalSettings) {
    return UpdateInbound400JSONResponse{errBadRequest("invalid settings")}, nil
  }
  settingsMap := map[string]any{}
  if err := json.Unmarshal(finalSettings, &settingsMap); err != nil {
    return UpdateInbound400JSONResponse{errBadRequest("invalid settings")}, nil
  }
  if err := inbval.ValidateSettings(finalProto, settingsMap); err != nil {
    return UpdateInbound400JSONResponse{errBadRequest(err.Error())}, nil
  }

  inb, err := s.store.UpdateInbound(ctx, request.Id, upd)
  if err != nil {
    if errors.Is(err, db.ErrNotFound) {
      return UpdateInbound404JSONResponse{errNotFound("inbound not found")}, nil
    }
    if errors.Is(err, db.ErrConflict) {
      return UpdateInbound409JSONResponse{errConflict("conflict")}, nil
    }
    return UpdateInbound500JSONResponse{errInternal("update inbound failed")}, nil
  }

  n, err := s.store.GetNodeByID(ctx, inb.NodeID)
  if err != nil {
    return UpdateInbound200JSONResponse{
      Data: dbInboundToAPI(inb),
      Sync: SyncResult{Status: "error", Error: strPtr("get node failed")},
    }, nil
  }
  sync := trySyncNodeWithSource(ctx, s.store, n, triggerSourceInbound)
  return UpdateInbound200JSONResponse{
    Data: dbInboundToAPI(inb),
    Sync: syncResultToAPI(sync),
  }, nil
}

func (s *Server) DeleteInbound(ctx context.Context, request DeleteInboundRequestObject) (DeleteInboundResponseObject, error) {
  cur, err := s.store.GetInboundByID(ctx, request.Id)
  if err != nil {
    if errors.Is(err, db.ErrNotFound) {
      return DeleteInbound404JSONResponse{errNotFound("inbound not found")}, nil
    }
    return DeleteInbound500JSONResponse{errInternal("get inbound failed")}, nil
  }

  if err := s.store.DeleteInbound(ctx, request.Id); err != nil {
    if errors.Is(err, db.ErrNotFound) {
      return DeleteInbound404JSONResponse{errNotFound("inbound not found")}, nil
    }
    return DeleteInbound500JSONResponse{errInternal("delete inbound failed")}, nil
  }

  n, err := s.store.GetNodeByID(ctx, cur.NodeID)
  if err != nil {
    return DeleteInbound200JSONResponse{
      Status: "ok",
      Sync:   SyncResult{Status: "error", Error: strPtr("get node failed")},
    }, nil
  }
  sync := trySyncNodeWithSource(ctx, s.store, n, triggerSourceInbound)
  return DeleteInbound200JSONResponse{
    Status: "ok",
    Sync:   syncResultToAPI(sync),
  }, nil
}

func dbInboundToAPI(inb db.Inbound) Inbound {
  return Inbound{
    Id:                inb.ID,
    Uuid:              inb.UUID,
    Tag:               inb.Tag,
    NodeId:            inb.NodeID,
    Protocol:          inb.Protocol,
    ListenPort:        inb.ListenPort,
    PublicPort:        inb.PublicPort,
    Settings:          rawJSONToMap(inb.Settings),
    TlsSettings:       rawJSONToMap(inb.TLSSettings),
    TransportSettings: rawJSONToMap(inb.TransportSettings),
  }
}

func syncResultToAPI(r syncResult) SyncResult {
  sr := SyncResult{Status: r.Status}
  if r.Error != "" {
    sr.Error = strPtr(r.Error)
  }
  return sr
}

func strPtr(s string) *string { return &s }

func marshalOptionalMap(m *map[string]interface{}) json.RawMessage {
  if m == nil {
    return nil
  }
  b, _ := json.Marshal(*m)
  return b
}
