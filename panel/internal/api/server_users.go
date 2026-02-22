package api

import (
  "context"
  "errors"
  "strings"
  "time"

  "sboard/panel/internal/db"
  "sboard/panel/internal/userstate"
)

func (s *Server) ListUsers(ctx context.Context, request ListUsersRequestObject) (ListUsersResponseObject, error) {
  limit, offset, err := paginationDefaults(request.Params.Limit, request.Params.Offset)
  if err != nil {
    return ListUsers400JSONResponse{errBadRequest("invalid pagination")}, nil
  }

  var status string
  if request.Params.Status != nil {
    status = string(*request.Params.Status)
  }

  users, err := listUsersForStatus(ctx, s.store, status, limit, offset)
  if err != nil {
    return ListUsers500JSONResponse{errInternal("list users failed")}, nil
  }

  userIDs := make([]int64, len(users))
  for i, u := range users {
    userIDs[i] = u.ID
  }
  groupIDsMap, err := s.store.ListUserGroupIDsBatch(ctx, userIDs)
  if err != nil {
    return ListUsers500JSONResponse{errInternal("load user groups failed")}, nil
  }

  out := make([]User, 0, len(users))
  for _, u := range users {
    dto := dbUserToAPI(u)
    dto.GroupIds = groupIDsMap[u.ID]
    if dto.GroupIds == nil {
      dto.GroupIds = []int64{}
    }
    out = append(out, dto)
  }
  return ListUsers200JSONResponse{Data: out}, nil
}

func (s *Server) CreateUser(ctx context.Context, request CreateUserRequestObject) (CreateUserResponseObject, error) {
  username := strings.TrimSpace(request.Body.Username)
  if username == "" {
    return CreateUser400JSONResponse{errBadRequest("invalid username")}, nil
  }

  user, err := s.store.CreateUser(ctx, username)
  if err != nil {
    if errors.Is(err, db.ErrConflict) {
      return CreateUser409JSONResponse{errConflict("username already exists")}, nil
    }
    return CreateUser500JSONResponse{errInternal("create user failed")}, nil
  }

  dto, err := buildUserAPIDTO(ctx, s.store, user)
  if err != nil {
    return CreateUser500JSONResponse{errInternal("load user groups failed")}, nil
  }
  return CreateUser201JSONResponse{Data: dto}, nil
}

func (s *Server) GetUser(ctx context.Context, request GetUserRequestObject) (GetUserResponseObject, error) {
  user, err := s.store.GetUserByID(ctx, request.Id)
  if err != nil {
    if errors.Is(err, db.ErrNotFound) {
      return GetUser404JSONResponse{errNotFound("user not found")}, nil
    }
    return GetUser500JSONResponse{errInternal("get user failed")}, nil
  }

  dto, err := buildUserAPIDTO(ctx, s.store, user)
  if err != nil {
    return GetUser500JSONResponse{errInternal("load user groups failed")}, nil
  }
  return GetUser200JSONResponse{Data: dto}, nil
}

func (s *Server) UpdateUser(ctx context.Context, request UpdateUserRequestObject) (UpdateUserResponseObject, error) {
  update, err := parseUserUpdateFromAPI(request.Body)
  if err != nil {
    return UpdateUser400JSONResponse{errBadRequest(err.Error())}, nil
  }

  user, err := s.store.UpdateUser(ctx, request.Id, update)
  if err != nil {
    if errors.Is(err, db.ErrConflict) {
      return UpdateUser409JSONResponse{errConflict("username already exists")}, nil
    }
    if errors.Is(err, db.ErrNotFound) {
      return UpdateUser404JSONResponse{errNotFound("user not found")}, nil
    }
    return UpdateUser500JSONResponse{errInternal("update user failed")}, nil
  }

  syncNodesForUser(ctx, s.store, user.ID)

  dto, err := buildUserAPIDTO(ctx, s.store, user)
  if err != nil {
    return UpdateUser500JSONResponse{errInternal("load user groups failed")}, nil
  }
  return UpdateUser200JSONResponse{Data: dto}, nil
}

func (s *Server) DeleteUser(ctx context.Context, request DeleteUserRequestObject) (DeleteUserResponseObject, error) {
  hard := request.Params.Hard != nil && *request.Params.Hard == DeleteUserParamsHardTrue

  if hard {
    groupIDs, err := s.store.ListUserGroupIDs(ctx, request.Id)
    if err != nil {
      return DeleteUser500JSONResponse{errInternal("list user groups failed")}, nil
    }
    if err := s.store.DeleteUser(ctx, request.Id); err != nil {
      if errors.Is(err, db.ErrNotFound) {
        return DeleteUser404JSONResponse{errNotFound("user not found")}, nil
      }
      return DeleteUser500JSONResponse{errInternal("delete user failed")}, nil
    }
    syncNodesByGroupIDsWithSource(ctx, s.store, groupIDs, triggerSourceUser)
    msg := "user deleted"
    return DeleteUser200JSONResponse{Message: &msg}, nil
  }

  // Soft delete: disable user
  if err := s.store.DisableUser(ctx, request.Id); err != nil {
    if errors.Is(err, db.ErrNotFound) {
      return DeleteUser404JSONResponse{errNotFound("user not found")}, nil
    }
    return DeleteUser500JSONResponse{errInternal("disable user failed")}, nil
  }

  user, err := s.store.GetUserByID(ctx, request.Id)
  if err != nil {
    if errors.Is(err, db.ErrNotFound) {
      return DeleteUser404JSONResponse{errNotFound("user not found")}, nil
    }
    return DeleteUser500JSONResponse{errInternal("get user failed")}, nil
  }

  syncNodesForUser(ctx, s.store, user.ID)

  dto, err := buildUserAPIDTO(ctx, s.store, user)
  if err != nil {
    return DeleteUser500JSONResponse{errInternal("load user groups failed")}, nil
  }
  return DeleteUser200JSONResponse{Data: &dto}, nil
}

func (s *Server) GetUserGroups(ctx context.Context, request GetUserGroupsRequestObject) (GetUserGroupsResponseObject, error) {
  if _, err := s.store.GetUserByID(ctx, request.Id); err != nil {
    if errors.Is(err, db.ErrNotFound) {
      return GetUserGroups404JSONResponse{errNotFound("user not found")}, nil
    }
    return GetUserGroups500JSONResponse{errInternal("get user failed")}, nil
  }

  ids, err := s.store.ListUserGroupIDs(ctx, request.Id)
  if err != nil {
    return GetUserGroups500JSONResponse{errInternal("list user groups failed")}, nil
  }
  if ids == nil {
    ids = []int64{}
  }
  return GetUserGroups200JSONResponse{Data: struct {
    GroupIds []int64 `json:"group_ids"`
  }{GroupIds: ids}}, nil
}

func (s *Server) ReplaceUserGroups(ctx context.Context, request ReplaceUserGroupsRequestObject) (ReplaceUserGroupsResponseObject, error) {
  previousGroupIDs, err := s.store.ListUserGroupIDs(ctx, request.Id)
  if err != nil {
    return ReplaceUserGroups500JSONResponse{errInternal("list user groups failed")}, nil
  }

  if err := s.store.ReplaceUserGroups(ctx, request.Id, request.Body.GroupIds); err != nil {
    if errors.Is(err, db.ErrNotFound) {
      return ReplaceUserGroups404JSONResponse{errNotFound("not found")}, nil
    }
    if errors.Is(err, db.ErrConflict) {
      return ReplaceUserGroups409JSONResponse{errConflict("conflict")}, nil
    }
    return ReplaceUserGroups400JSONResponse{errBadRequest("invalid group_ids")}, nil
  }

  ids, err := s.store.ListUserGroupIDs(ctx, request.Id)
  if err != nil {
    return ReplaceUserGroups500JSONResponse{errInternal("list user groups failed")}, nil
  }
  if ids == nil {
    ids = []int64{}
  }

  syncGroupIDs := append([]int64{}, previousGroupIDs...)
  syncGroupIDs = append(syncGroupIDs, ids...)
  syncNodesByGroupIDsWithSource(ctx, s.store, syncGroupIDs, triggerSourceUser)

  return ReplaceUserGroups200JSONResponse{Data: struct {
    GroupIds []int64 `json:"group_ids"`
  }{GroupIds: ids}}, nil
}

// dbUserToAPI converts a db.User to the generated API User type.
func dbUserToAPI(u db.User) User {
  return User{
    Id:              u.ID,
    Uuid:            u.UUID,
    Username:        u.Username,
    GroupIds:        []int64{},
    TrafficLimit:    u.TrafficLimit,
    TrafficUsed:     u.TrafficUsed,
    TrafficResetDay: u.TrafficResetDay,
    ExpireAt:        timeInSystemTimezonePtr(u.ExpireAt),
    Status:          userstate.EffectiveStatus(u, time.Now().UTC()),
  }
}

// buildUserAPIDTO converts a db.User to an API User with group IDs populated.
func buildUserAPIDTO(ctx context.Context, store *db.Store, u db.User) (User, error) {
  dto := dbUserToAPI(u)
  groupIDs, err := store.ListUserGroupIDs(ctx, u.ID)
  if err != nil {
    return User{}, err
  }
  dto.GroupIds = groupIDs
  if dto.GroupIds == nil {
    dto.GroupIds = []int64{}
  }
  return dto, nil
}

// parseUserUpdateFromAPI converts an API UpdateUserRequest to a db.UserUpdate.
func parseUserUpdateFromAPI(req *UpdateUserJSONRequestBody) (db.UserUpdate, error) {
  update := db.UserUpdate{}
  if req.Username != nil {
    name := strings.TrimSpace(*req.Username)
    if name == "" {
      return update, errors.New("invalid username")
    }
    update.Username = &name
  }
  if req.Status != nil {
    if !isValidStatus(*req.Status) {
      return update, errors.New("invalid status")
    }
    update.Status = req.Status
  }
  if req.ExpireAt != nil {
    update.ExpireAtSet = true
    if strings.TrimSpace(*req.ExpireAt) == "" {
      update.ExpireAt = nil
    } else {
      t, err := time.Parse(time.RFC3339, *req.ExpireAt)
      if err != nil {
        return update, errors.New("invalid expire_at")
      }
      update.ExpireAt = &t
    }
  }
  if req.TrafficLimit != nil {
    if *req.TrafficLimit < 0 {
      return update, errors.New("invalid traffic_limit")
    }
    update.TrafficLimit = req.TrafficLimit
  }
  if req.TrafficResetDay != nil {
    if *req.TrafficResetDay < 0 || *req.TrafficResetDay > 31 {
      return update, errors.New("invalid traffic_reset_day")
    }
    update.TrafficResetDay = req.TrafficResetDay
  }
  return update, nil
}
