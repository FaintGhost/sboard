package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"sboard/panel/internal/buildinfo"
	"sboard/panel/internal/db"
	"sboard/panel/internal/userstate"
)

const defaultUserListLimit = 50

var allowedUserStatus = map[string]struct{}{
	"active":           {},
	"disabled":         {},
	"expired":          {},
	"traffic_exceeded": {},
}

type userDTO struct {
	ID              int64      `json:"id"`
	UUID            string     `json:"uuid"`
	Username        string     `json:"username"`
	GroupIDs        []int64    `json:"group_ids"`
	TrafficLimit    int64      `json:"traffic_limit"`
	TrafficUsed     int64      `json:"traffic_used"`
	TrafficResetDay int        `json:"traffic_reset_day"`
	ExpireAt        *time.Time `json:"expire_at"`
	Status          string     `json:"status"`
}

type createUserReq struct {
	Username string `json:"username"`
}

type updateUserReq struct {
	Username        *string `json:"username"`
	Status          *string `json:"status"`
	ExpireAt        *string `json:"expire_at"`
	TrafficLimit    *int64  `json:"traffic_limit"`
	TrafficResetDay *int    `json:"traffic_reset_day"`
}

type systemInfoDTO struct {
	PanelVersion   string `json:"panel_version"`
	PanelCommitID  string `json:"panel_commit_id"`
	SingBoxVersion string `json:"sing_box_version"`
}

func UsersCreate(store *db.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !ensureStore(c, store) {
			return
		}
		var req createUserReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		username := strings.TrimSpace(req.Username)
		if username == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid username"})
			return
		}
		user, err := store.CreateUser(c.Request.Context(), username)
		if err != nil {
			if errors.Is(err, db.ErrConflict) {
				c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "create user failed"})
			return
		}
		dto, err := buildUserDTO(c.Request.Context(), store, user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "load user groups failed"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": dto})
	}
}

func UsersList(store *db.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !ensureStore(c, store) {
			return
		}
		limit, offset, err := parseLimitOffset(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pagination"})
			return
		}
		status := strings.TrimSpace(c.Query("status"))
		if status != "" && !isValidStatus(status) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
			return
		}
		users, err := listUsersForStatus(c.Request.Context(), store, status, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "list users failed"})
			return
		}

		// Batch fetch group IDs to avoid N+1 queries
		userIDs := make([]int64, len(users))
		for i, u := range users {
			userIDs[i] = u.ID
		}
		groupIDsMap, err := store.ListUserGroupIDsBatch(c.Request.Context(), userIDs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "load user groups failed"})
			return
		}

		out := make([]userDTO, 0, len(users))
		for _, u := range users {
			dto := toUserDTO(u)
			dto.GroupIDs = groupIDsMap[u.ID]
			out = append(out, dto)
		}
		c.JSON(http.StatusOK, gin.H{"data": out})
	}
}

func UsersGet(store *db.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !ensureStore(c, store) {
			return
		}
		id, err := parseID(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		user, err := store.GetUserByID(c.Request.Context(), id)
		if err != nil {
			if errors.Is(err, db.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "get user failed"})
			return
		}
		dto, err := buildUserDTO(c.Request.Context(), store, user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "load user groups failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": dto})
	}
}

func UsersUpdate(store *db.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !ensureStore(c, store) {
			return
		}
		id, err := parseID(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		var req updateUserReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		update, err := parseUserUpdate(req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		user, err := store.UpdateUser(c.Request.Context(), id, update)
		if err != nil {
			if errors.Is(err, db.ErrConflict) {
				c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
				return
			}
			if errors.Is(err, db.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "update user failed"})
			return
		}
		syncNodesForUser(c.Request.Context(), store, user.ID)
		dto, err := buildUserDTO(c.Request.Context(), store, user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "load user groups failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": dto})
	}
}

func UsersDelete(store *db.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !ensureStore(c, store) {
			return
		}
		id, err := parseID(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		// Check if hard delete is requested
		hard := c.Query("hard") == "true"

		if hard {
			groupIDs, err := store.ListUserGroupIDs(c.Request.Context(), id)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "list user groups failed"})
				return
			}
			if err := store.DeleteUser(c.Request.Context(), id); err != nil {
				if errors.Is(err, db.ErrNotFound) {
					c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": "delete user failed"})
				return
			}
			syncNodesByGroupIDsWithSource(c.Request.Context(), store, groupIDs, triggerSourceUser)
			c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
			return
		}

		// Soft delete: disable user
		if err := store.DisableUser(c.Request.Context(), id); err != nil {
			if errors.Is(err, db.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "disable user failed"})
			return
		}
		user, err := store.GetUserByID(c.Request.Context(), id)
		if err != nil {
			if errors.Is(err, db.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "get user failed"})
			return
		}
		syncNodesForUser(c.Request.Context(), store, user.ID)
		dto, err := buildUserDTO(c.Request.Context(), store, user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "load user groups failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": dto})
	}
}

func SystemInfoGet() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"data": systemInfoDTO{
				PanelVersion:   nonEmptyOrNA(buildinfo.PanelVersion),
				PanelCommitID:  nonEmptyOrNA(buildinfo.PanelCommitID),
				SingBoxVersion: nonEmptyOrNA(buildinfo.SingBoxVersion),
			},
		})
	}
}

func ensureStore(c *gin.Context, store *db.Store) bool {
	if store == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "store not ready"})
		return false
	}
	return true
}

func listUsersForStatus(ctx context.Context, store *db.Store, status string, limit int, offset int) ([]db.User, error) {
	if status == "" {
		return store.ListUsers(ctx, limit, offset, "")
	}
	if status == userstate.StatusActive ||
		status == userstate.StatusDisabled ||
		status == userstate.StatusExpired ||
		status == userstate.StatusTrafficExceeded {
		return store.ListUsersByEffectiveStatus(ctx, limit, offset, status, time.Now().UTC())
	}
	return nil, errors.New("unsupported status")
}

func parseLimitOffset(c *gin.Context) (int, int, error) {
	limitStr := c.DefaultQuery("limit", strconv.Itoa(defaultUserListLimit))
	offsetStr := c.DefaultQuery("offset", "0")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 0 {
		return 0, 0, errors.New("invalid limit")
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		return 0, 0, errors.New("invalid offset")
	}
	return limit, offset, nil
}

func parseID(value string) (int64, error) {
	return strconv.ParseInt(value, 10, 64)
}

func parseUserUpdate(req updateUserReq) (db.UserUpdate, error) {
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

func isValidStatus(status string) bool {
	_, ok := allowedUserStatus[status]
	return ok
}

func toUserDTO(u db.User) userDTO {
	return userDTO{
		ID:              u.ID,
		UUID:            u.UUID,
		Username:        u.Username,
		GroupIDs:        []int64{},
		TrafficLimit:    u.TrafficLimit,
		TrafficUsed:     u.TrafficUsed,
		TrafficResetDay: u.TrafficResetDay,
		ExpireAt:        timeInSystemTimezonePtr(u.ExpireAt),
		Status:          effectiveUserStatus(u),
	}
}

func buildUserDTO(ctx context.Context, store *db.Store, u db.User) (userDTO, error) {
	dto := toUserDTO(u)
	groupIDs, err := store.ListUserGroupIDs(ctx, u.ID)
	if err != nil {
		return userDTO{}, err
	}
	dto.GroupIDs = groupIDs
	return dto, nil
}

func effectiveUserStatus(u db.User) string {
	return userstate.EffectiveStatus(u, time.Now().UTC())
}

func nonEmptyOrNA(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "N/A"
	}
	return value
}
