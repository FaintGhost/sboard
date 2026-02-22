package api

import (
  "context"
  "errors"
  "sort"
  "strings"

  "sboard/panel/internal/db"
)

func (s *Server) ListGroups(ctx context.Context, request ListGroupsRequestObject) (ListGroupsResponseObject, error) {
  limit, offset, err := paginationDefaults(request.Params.Limit, request.Params.Offset)
  if err != nil {
    return ListGroups400JSONResponse{errBadRequest("invalid pagination")}, nil
  }

  groups, err := s.store.ListGroups(ctx, limit, offset)
  if err != nil {
    return ListGroups500JSONResponse{errInternal("list groups failed")}, nil
  }

  out := make([]Group, 0, len(groups))
  for _, g := range groups {
    out = append(out, dbGroupToAPI(g))
  }
  return ListGroups200JSONResponse{Data: out}, nil
}

func (s *Server) CreateGroup(ctx context.Context, request CreateGroupRequestObject) (CreateGroupResponseObject, error) {
  name := strings.TrimSpace(request.Body.Name)
  if name == "" {
    return CreateGroup400JSONResponse{errBadRequest("invalid name")}, nil
  }

  desc := ""
  if request.Body.Description != nil {
    desc = strings.TrimSpace(*request.Body.Description)
  }

  g, err := s.store.CreateGroup(ctx, name, desc)
  if err != nil {
    if errors.Is(err, db.ErrConflict) {
      return CreateGroup409JSONResponse{errConflict("group name already exists")}, nil
    }
    return CreateGroup500JSONResponse{errInternal("create group failed")}, nil
  }
  return CreateGroup201JSONResponse{Data: dbGroupToAPI(g)}, nil
}

func (s *Server) GetGroup(ctx context.Context, request GetGroupRequestObject) (GetGroupResponseObject, error) {
  g, err := s.store.GetGroupByID(ctx, request.Id)
  if err != nil {
    if errors.Is(err, db.ErrNotFound) {
      return GetGroup404JSONResponse{errNotFound("group not found")}, nil
    }
    return GetGroup500JSONResponse{errInternal("get group failed")}, nil
  }
  return GetGroup200JSONResponse{Data: dbGroupToAPI(g)}, nil
}

func (s *Server) UpdateGroup(ctx context.Context, request UpdateGroupRequestObject) (UpdateGroupResponseObject, error) {
  upd := db.GroupUpdate{}
  if request.Body.Name != nil {
    name := strings.TrimSpace(*request.Body.Name)
    if name == "" {
      return UpdateGroup400JSONResponse{errBadRequest("invalid name")}, nil
    }
    upd.Name = &name
  }
  if request.Body.Description != nil {
    desc := strings.TrimSpace(*request.Body.Description)
    upd.Description = &desc
  }

  g, err := s.store.UpdateGroup(ctx, request.Id, upd)
  if err != nil {
    if errors.Is(err, db.ErrConflict) {
      return UpdateGroup409JSONResponse{errConflict("group name already exists")}, nil
    }
    if errors.Is(err, db.ErrNotFound) {
      return UpdateGroup404JSONResponse{errNotFound("group not found")}, nil
    }
    return UpdateGroup500JSONResponse{errInternal("update group failed")}, nil
  }
  return UpdateGroup200JSONResponse{Data: dbGroupToAPI(g)}, nil
}

func (s *Server) DeleteGroup(ctx context.Context, request DeleteGroupRequestObject) (DeleteGroupResponseObject, error) {
  if err := s.store.DeleteGroup(ctx, request.Id); err != nil {
    if errors.Is(err, db.ErrNotFound) {
      return DeleteGroup404JSONResponse{errNotFound("group not found")}, nil
    }
    if errors.Is(err, db.ErrConflict) {
      return DeleteGroup409JSONResponse{errConflict("group is in use")}, nil
    }
    return DeleteGroup500JSONResponse{errInternal("delete group failed")}, nil
  }
  return DeleteGroup200JSONResponse{Status: "ok"}, nil
}

func (s *Server) ListGroupUsers(ctx context.Context, request ListGroupUsersRequestObject) (ListGroupUsersResponseObject, error) {
  users, err := s.store.ListGroupUsers(ctx, request.Id)
  if err != nil {
    if errors.Is(err, db.ErrNotFound) {
      return ListGroupUsers404JSONResponse{errNotFound("group not found")}, nil
    }
    return ListGroupUsers500JSONResponse{errInternal("list group users failed")}, nil
  }

  out := make([]GroupUsersListItem, 0, len(users))
  for _, u := range users {
    out = append(out, GroupUsersListItem{
      Id:           u.ID,
      Uuid:         u.UUID,
      Username:     u.Username,
      TrafficLimit: u.TrafficLimit,
      TrafficUsed:  u.TrafficUsed,
      Status:       effectiveUserStatus(u),
    })
  }
  return ListGroupUsers200JSONResponse{Data: out}, nil
}

func (s *Server) ReplaceGroupUsers(ctx context.Context, request ReplaceGroupUsersRequestObject) (ReplaceGroupUsersResponseObject, error) {
  if err := s.store.ReplaceGroupUsers(ctx, request.Id, request.Body.UserIds); err != nil {
    if errors.Is(err, db.ErrNotFound) {
      return ReplaceGroupUsers404JSONResponse{errNotFound("not found")}, nil
    }
    if errors.Is(err, db.ErrConflict) {
      return ReplaceGroupUsers409JSONResponse{errConflict("conflict")}, nil
    }
    return ReplaceGroupUsers400JSONResponse{errBadRequest("invalid user_ids")}, nil
  }

  syncNodesByGroupIDs(ctx, s.store, []int64{request.Id})

  unique := uniquePositiveInt64(request.Body.UserIds)
  sort.Slice(unique, func(i, j int) bool { return unique[i] < unique[j] })

  return ReplaceGroupUsers200JSONResponse{Data: struct {
    UserIds []int64 `json:"user_ids"`
  }{UserIds: unique}}, nil
}

func dbGroupToAPI(g db.Group) Group {
  return Group{
    Id:          g.ID,
    Name:        g.Name,
    Description: g.Description,
    MemberCount: g.MemberCount,
  }
}
