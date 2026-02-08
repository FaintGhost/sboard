package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"sboard/panel/internal/db"
)

type userGroupsDTO struct {
	GroupIDs []int64 `json:"group_ids"`
}

type putUserGroupsReq struct {
	GroupIDs []int64 `json:"group_ids"`
}

func UserGroupsGet(store *db.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !ensureStore(c, store) {
			return
		}
		id, err := parseID(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		if _, err := store.GetUserByID(c.Request.Context(), id); err != nil {
			if errors.Is(err, db.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "get user failed"})
			return
		}
		ids, err := store.ListUserGroupIDs(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "list user groups failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": userGroupsDTO{GroupIDs: ids}})
	}
}

func UserGroupsPut(store *db.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !ensureStore(c, store) {
			return
		}
		id, err := parseID(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		var req putUserGroupsReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		previousGroupIDs, err := store.ListUserGroupIDs(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "list user groups failed"})
			return
		}

		if err := store.ReplaceUserGroups(c.Request.Context(), id, req.GroupIDs); err != nil {
			if errors.Is(err, db.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			if errors.Is(err, db.ErrConflict) {
				c.JSON(http.StatusConflict, gin.H{"error": "conflict"})
				return
			}
			// ReplaceUserGroups uses plain errors for input validation.
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_ids"})
			return
		}
		ids, err := store.ListUserGroupIDs(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "list user groups failed"})
			return
		}

		syncGroupIDs := append([]int64{}, previousGroupIDs...)
		syncGroupIDs = append(syncGroupIDs, ids...)
		syncNodesByGroupIDsWithSource(c.Request.Context(), store, syncGroupIDs, triggerSourceUser)

		c.JSON(http.StatusOK, gin.H{"data": userGroupsDTO{GroupIDs: ids}})
	}
}
