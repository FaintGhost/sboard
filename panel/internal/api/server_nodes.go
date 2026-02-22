package api

import (
  "context"
  "encoding/json"
  "errors"
  "strings"

  "sboard/panel/internal/db"
  "sboard/panel/internal/node"
)

func (s *Server) ListNodes(ctx context.Context, request ListNodesRequestObject) (ListNodesResponseObject, error) {
  limit, offset, err := paginationDefaults(request.Params.Limit, request.Params.Offset)
  if err != nil {
    return ListNodes400JSONResponse{errBadRequest("invalid pagination")}, nil
  }

  items, err := s.store.ListNodes(ctx, limit, offset)
  if err != nil {
    return ListNodes500JSONResponse{errInternal("list nodes failed")}, nil
  }

  out := make([]Node, 0, len(items))
  for _, n := range items {
    out = append(out, dbNodeToAPI(n))
  }
  return ListNodes200JSONResponse{Data: out}, nil
}

func (s *Server) CreateNode(ctx context.Context, request CreateNodeRequestObject) (CreateNodeResponseObject, error) {
  name := strings.TrimSpace(request.Body.Name)
  apiAddr := strings.TrimSpace(request.Body.ApiAddress)
  pubAddr := strings.TrimSpace(request.Body.PublicAddress)
  if name == "" || apiAddr == "" || request.Body.ApiPort <= 0 || strings.TrimSpace(request.Body.SecretKey) == "" || pubAddr == "" {
    return CreateNode400JSONResponse{errBadRequest("invalid node")}, nil
  }

  n, err := s.store.CreateNode(ctx, db.NodeCreate{
    Name:          name,
    APIAddress:    apiAddr,
    APIPort:       request.Body.ApiPort,
    SecretKey:     request.Body.SecretKey,
    PublicAddress: pubAddr,
    GroupID:       request.Body.GroupId,
  })
  if err != nil {
    if errors.Is(err, db.ErrConflict) {
      return CreateNode409JSONResponse{errConflict("conflict")}, nil
    }
    return CreateNode500JSONResponse{errInternal("create node failed")}, nil
  }
  return CreateNode201JSONResponse{Data: dbNodeToAPI(n)}, nil
}

func (s *Server) GetNode(ctx context.Context, request GetNodeRequestObject) (GetNodeResponseObject, error) {
  n, err := s.store.GetNodeByID(ctx, request.Id)
  if err != nil {
    if errors.Is(err, db.ErrNotFound) {
      return GetNode404JSONResponse{errNotFound("node not found")}, nil
    }
    return GetNode500JSONResponse{errInternal("get node failed")}, nil
  }
  return GetNode200JSONResponse{Data: dbNodeToAPI(n)}, nil
}

func (s *Server) UpdateNode(ctx context.Context, request UpdateNodeRequestObject) (UpdateNodeResponseObject, error) {
  upd := db.NodeUpdate{}
  if request.Body.Name != nil {
    name := strings.TrimSpace(*request.Body.Name)
    if name == "" {
      return UpdateNode400JSONResponse{errBadRequest("invalid name")}, nil
    }
    upd.Name = &name
  }
  if request.Body.ApiAddress != nil {
    addr := strings.TrimSpace(*request.Body.ApiAddress)
    if addr == "" {
      return UpdateNode400JSONResponse{errBadRequest("invalid api_address")}, nil
    }
    upd.APIAddress = &addr
  }
  if request.Body.ApiPort != nil {
    if *request.Body.ApiPort <= 0 {
      return UpdateNode400JSONResponse{errBadRequest("invalid api_port")}, nil
    }
    upd.APIPort = request.Body.ApiPort
  }
  if request.Body.SecretKey != nil {
    if strings.TrimSpace(*request.Body.SecretKey) == "" {
      return UpdateNode400JSONResponse{errBadRequest("invalid secret_key")}, nil
    }
    upd.SecretKey = request.Body.SecretKey
  }
  if request.Body.PublicAddress != nil {
    addr := strings.TrimSpace(*request.Body.PublicAddress)
    if addr == "" {
      return UpdateNode400JSONResponse{errBadRequest("invalid public_address")}, nil
    }
    upd.PublicAddress = &addr
  }
  // GroupId is always present in the request body (nullable, not omitempty in spec).
  // If it's nil, we clear the group; if it's set, we assign it.
  upd.GroupIDSet = true
  upd.GroupID = request.Body.GroupId
  if upd.GroupID != nil && *upd.GroupID <= 0 {
    return UpdateNode400JSONResponse{errBadRequest("invalid group_id")}, nil
  }

  n, err := s.store.UpdateNode(ctx, request.Id, upd)
  if err != nil {
    if errors.Is(err, db.ErrNotFound) {
      return UpdateNode404JSONResponse{errNotFound("node not found")}, nil
    }
    if errors.Is(err, db.ErrConflict) {
      return UpdateNode409JSONResponse{errConflict("conflict")}, nil
    }
    return UpdateNode500JSONResponse{errInternal("update node failed")}, nil
  }
  return UpdateNode200JSONResponse{Data: dbNodeToAPI(n)}, nil
}

func (s *Server) DeleteNode(ctx context.Context, request DeleteNodeRequestObject) (DeleteNodeResponseObject, error) {
  force := request.Params.Force != nil && *request.Params.Force == DeleteNodeParamsForceTrue

  if !force {
    if err := s.store.DeleteNode(ctx, request.Id); err != nil {
      if errors.Is(err, db.ErrNotFound) {
        return DeleteNode404JSONResponse{errNotFound("node not found")}, nil
      }
      if errors.Is(err, db.ErrConflict) {
        return DeleteNode409JSONResponse{errConflict("node is in use")}, nil
      }
      return DeleteNode500JSONResponse{errInternal("delete node failed")}, nil
    }
    return DeleteNode200JSONResponse{Status: "ok"}, nil
  }

  n, err := s.store.GetNodeByID(ctx, request.Id)
  if err != nil {
    if errors.Is(err, db.ErrNotFound) {
      return DeleteNode404JSONResponse{errNotFound("node not found")}, nil
    }
    return DeleteNode500JSONResponse{errInternal("get node failed")}, nil
  }

  inbounds, err := s.store.ListInbounds(ctx, 10000, 0, n.ID)
  if err != nil {
    return DeleteNode500JSONResponse{errInternal("list inbounds failed")}, nil
  }

  if len(inbounds) > 0 {
    lock := nodeLock(n.ID)
    lock.Lock()
    defer lock.Unlock()

    client := nodeClientFactory()
    emptyPayload := node.SyncPayload{Inbounds: []map[string]any{}}
    if err := client.SyncConfig(ctx, n, emptyPayload); err != nil {
      return DeleteNode502JSONResponse{errBadGateway("force drain failed: " + err.Error())}, nil
    }
    _ = s.store.MarkNodeOnline(ctx, n.ID, s.store.NowUTC())
  }

  deletedInbounds, err := s.store.DeleteInboundsByNode(ctx, request.Id)
  if err != nil {
    return DeleteNode500JSONResponse{errInternal("delete node inbounds failed")}, nil
  }
  if err := s.store.DeleteNode(ctx, request.Id); err != nil {
    if errors.Is(err, db.ErrNotFound) {
      return DeleteNode404JSONResponse{errNotFound("node not found")}, nil
    }
    if errors.Is(err, db.ErrConflict) {
      return DeleteNode409JSONResponse{errConflict("node is in use")}, nil
    }
    return DeleteNode500JSONResponse{errInternal("delete node failed")}, nil
  }

  forceTrue := true
  deletedCount := int(deletedInbounds)
  return DeleteNode200JSONResponse{Status: "ok", Force: &forceTrue, DeletedInbounds: &deletedCount}, nil
}

func (s *Server) GetNodeHealth(ctx context.Context, request GetNodeHealthRequestObject) (GetNodeHealthResponseObject, error) {
  n, err := s.store.GetNodeByID(ctx, request.Id)
  if err != nil {
    if errors.Is(err, db.ErrNotFound) {
      return GetNodeHealth404JSONResponse{errNotFound("node not found")}, nil
    }
    return GetNodeHealth500JSONResponse{errInternal("get node failed")}, nil
  }

  client := nodeClientFactory()
  if err := client.Health(ctx, n); err != nil {
    _ = s.store.MarkNodeOffline(ctx, n.ID)
    return GetNodeHealth502JSONResponse{errBadGateway(err.Error())}, nil
  }
  _ = s.store.MarkNodeOnline(ctx, n.ID, s.store.Now().UTC())
  return GetNodeHealth200JSONResponse{Status: "ok"}, nil
}

func (s *Server) SyncNode(ctx context.Context, request SyncNodeRequestObject) (SyncNodeResponseObject, error) {
  n, err := s.store.GetNodeByID(ctx, request.Id)
  if err != nil {
    if errors.Is(err, db.ErrNotFound) {
      return SyncNode404JSONResponse{errNotFound("node not found")}, nil
    }
    return SyncNode500JSONResponse{errInternal("get node failed")}, nil
  }
  if n.GroupID == nil {
    return SyncNode400JSONResponse{errBadRequest("node group_id not set")}, nil
  }
  res := trySyncNode(ctx, s.store, n)
  if res.Status != "ok" {
    return SyncNode502JSONResponse{errBadGateway(res.Error)}, nil
  }
  _ = s.store.MarkNodeOnline(ctx, n.ID, s.store.Now().UTC())
  return SyncNode200JSONResponse{Status: "ok"}, nil
}

func (s *Server) ListNodeTraffic(ctx context.Context, request ListNodeTrafficRequestObject) (ListNodeTrafficResponseObject, error) {
  if !s.storeReady() {
    return ListNodeTraffic500JSONResponse{errInternal("store not ready")}, nil
  }
  limit, offset, err := paginationDefaults(request.Params.Limit, request.Params.Offset)
  if err != nil {
    return ListNodeTraffic400JSONResponse{errBadRequest("invalid pagination")}, nil
  }

  items, err := s.store.ListNodeTrafficSamples(ctx, request.Id, limit, offset)
  if err != nil {
    if errors.Is(err, db.ErrNotFound) {
      return ListNodeTraffic404JSONResponse{errNotFound("not found")}, nil
    }
    return ListNodeTraffic500JSONResponse{errInternal(err.Error())}, nil
  }

  out := make([]NodeTrafficSample, 0, len(items))
  for _, it := range items {
    out = append(out, NodeTrafficSample{
      Id:         it.ID,
      InboundTag: it.InboundTag,
      Upload:     it.Upload,
      Download:   it.Download,
      RecordedAt: timeRFC3339OrEmpty(it.RecordedAt),
    })
  }
  return ListNodeTraffic200JSONResponse{Data: out}, nil
}

func dbNodeToAPI(n db.Node) Node {
  return Node{
    Id:            n.ID,
    Uuid:          n.UUID,
    Name:          n.Name,
    ApiAddress:    n.APIAddress,
    ApiPort:       n.APIPort,
    SecretKey:     n.SecretKey,
    PublicAddress: n.PublicAddress,
    GroupId:       n.GroupID,
    Status:        n.Status,
    LastSeenAt:    timeInSystemTimezonePtr(n.LastSeenAt),
  }
}

// rawJSONToMap converts a json.RawMessage to map[string]interface{}.
func rawJSONToMap(raw json.RawMessage) map[string]interface{} {
  if len(raw) == 0 {
    return map[string]interface{}{}
  }
  var m map[string]interface{}
  if err := json.Unmarshal(raw, &m); err != nil {
    return map[string]interface{}{}
  }
  return m
}
