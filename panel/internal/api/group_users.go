package api

import (
	"errors"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"sboard/panel/internal/db"
)

type groupUsersListItemDTO struct {
	ID           int64  `json:"id"`
	UUID         string `json:"uuid"`
	Username     string `json:"username"`
	TrafficLimit int64  `json:"traffic_limit"`
	TrafficUsed  int64  `json:"traffic_used"`
	Status       string `json:"status"`
}

type groupUsersReplaceReq struct {
	UserIDs []int64 `json:"user_ids"`
}

type groupUsersReplaceDTO struct {
	UserIDs []int64 `json:"user_ids"`
}

func GroupUsersList(store *db.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !ensureStore(c, store) {
			return
		}

		groupID, err := parseID(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		users, err := store.ListGroupUsers(c.Request.Context(), groupID)
		if err != nil {
			if errors.Is(err, db.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "list group users failed"})
			return
		}

		out := make([]groupUsersListItemDTO, 0, len(users))
		for _, u := range users {
			out = append(out, groupUsersListItemDTO{
				ID:           u.ID,
				UUID:         u.UUID,
				Username:     u.Username,
				TrafficLimit: u.TrafficLimit,
				TrafficUsed:  u.TrafficUsed,
				Status:       effectiveUserStatus(u),
			})
		}

		c.JSON(http.StatusOK, gin.H{"data": out})
	}
}

func GroupUsersReplace(store *db.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !ensureStore(c, store) {
			return
		}

		groupID, err := parseID(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		var req groupUsersReplaceReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}

		if err := store.ReplaceGroupUsers(c.Request.Context(), groupID, req.UserIDs); err != nil {
			if errors.Is(err, db.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			if errors.Is(err, db.ErrConflict) {
				c.JSON(http.StatusConflict, gin.H{"error": "conflict"})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_ids"})
			return
		}

		syncNodesByGroupIDs(c.Request.Context(), store, []int64{groupID})

		// 返回去重排序后的 user_ids，方便前端直接使用。
		unique := uniquePositiveInt64(req.UserIDs)
		sort.Slice(unique, func(i, j int) bool { return unique[i] < unique[j] })

		c.JSON(http.StatusOK, gin.H{"data": groupUsersReplaceDTO{UserIDs: unique}})
	}
}
