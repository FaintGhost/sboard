package api

import (
  "errors"
  "net/http"
  "strings"

  "sboard/panel/internal/db"
  "github.com/gin-gonic/gin"
)

type groupDTO struct {
  ID          int64  `json:"id"`
  Name        string `json:"name"`
  Description string `json:"description"`
}

type createGroupReq struct {
  Name        string `json:"name"`
  Description string `json:"description"`
}

type updateGroupReq struct {
  Name        *string `json:"name"`
  Description *string `json:"description"`
}

func GroupsCreate(store *db.Store) gin.HandlerFunc {
  return func(c *gin.Context) {
    if !ensureStore(c, store) {
      return
    }
    var req createGroupReq
    if err := c.ShouldBindJSON(&req); err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
      return
    }
    name := strings.TrimSpace(req.Name)
    if name == "" {
      c.JSON(http.StatusBadRequest, gin.H{"error": "invalid name"})
      return
    }
    g, err := store.CreateGroup(c.Request.Context(), name, strings.TrimSpace(req.Description))
    if err != nil {
      if errors.Is(err, db.ErrConflict) {
        c.JSON(http.StatusConflict, gin.H{"error": "group name already exists"})
        return
      }
      c.JSON(http.StatusInternalServerError, gin.H{"error": "create group failed"})
      return
    }
    c.JSON(http.StatusCreated, gin.H{"data": toGroupDTO(g)})
  }
}

func GroupsList(store *db.Store) gin.HandlerFunc {
  return func(c *gin.Context) {
    if !ensureStore(c, store) {
      return
    }
    limit, offset, err := parseLimitOffset(c)
    if err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pagination"})
      return
    }
    groups, err := store.ListGroups(c.Request.Context(), limit, offset)
    if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "list groups failed"})
      return
    }
    out := make([]groupDTO, 0, len(groups))
    for _, g := range groups {
      out = append(out, toGroupDTO(g))
    }
    c.JSON(http.StatusOK, gin.H{"data": out})
  }
}

func GroupsGet(store *db.Store) gin.HandlerFunc {
  return func(c *gin.Context) {
    if !ensureStore(c, store) {
      return
    }
    id, err := parseID(c.Param("id"))
    if err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
      return
    }
    g, err := store.GetGroupByID(c.Request.Context(), id)
    if err != nil {
      if errors.Is(err, db.ErrNotFound) {
        c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
        return
      }
      c.JSON(http.StatusInternalServerError, gin.H{"error": "get group failed"})
      return
    }
    c.JSON(http.StatusOK, gin.H{"data": toGroupDTO(g)})
  }
}

func GroupsUpdate(store *db.Store) gin.HandlerFunc {
  return func(c *gin.Context) {
    if !ensureStore(c, store) {
      return
    }
    id, err := parseID(c.Param("id"))
    if err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
      return
    }
    var req updateGroupReq
    if err := c.ShouldBindJSON(&req); err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
      return
    }
    upd := db.GroupUpdate{}
    if req.Name != nil {
      name := strings.TrimSpace(*req.Name)
      if name == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid name"})
        return
      }
      upd.Name = &name
    }
    if req.Description != nil {
      desc := strings.TrimSpace(*req.Description)
      upd.Description = &desc
    }
    g, err := store.UpdateGroup(c.Request.Context(), id, upd)
    if err != nil {
      if errors.Is(err, db.ErrConflict) {
        c.JSON(http.StatusConflict, gin.H{"error": "group name already exists"})
        return
      }
      if errors.Is(err, db.ErrNotFound) {
        c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
        return
      }
      c.JSON(http.StatusInternalServerError, gin.H{"error": "update group failed"})
      return
    }
    c.JSON(http.StatusOK, gin.H{"data": toGroupDTO(g)})
  }
}

func GroupsDelete(store *db.Store) gin.HandlerFunc {
  return func(c *gin.Context) {
    if !ensureStore(c, store) {
      return
    }
    id, err := parseID(c.Param("id"))
    if err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
      return
    }
    if err := store.DeleteGroup(c.Request.Context(), id); err != nil {
      if errors.Is(err, db.ErrNotFound) {
        c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
        return
      }
      if errors.Is(err, db.ErrConflict) {
        c.JSON(http.StatusConflict, gin.H{"error": "group is in use"})
        return
      }
      c.JSON(http.StatusInternalServerError, gin.H{"error": "delete group failed"})
      return
    }
    c.JSON(http.StatusOK, gin.H{"status": "ok"})
  }
}

func toGroupDTO(g db.Group) groupDTO {
  return groupDTO{
    ID:          g.ID,
    Name:        g.Name,
    Description: g.Description,
  }
}

