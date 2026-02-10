package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"sboard/panel/internal/db"
	"sboard/panel/internal/password"
)

type adminProfileDTO struct {
	Username string `json:"username"`
}

type updateAdminProfileReq struct {
	NewUsername     string `json:"new_username"`
	OldPassword     string `json:"old_password"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}

func AdminProfileGet(store *db.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !ensureStore(c, store) {
			return
		}

		admin, ok, err := db.AdminGetFirst(store)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		if !ok {
			c.JSON(http.StatusPreconditionRequired, gin.H{"error": "needs setup"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": adminProfileDTO{Username: admin.Username}})
	}
}

func AdminProfilePut(store *db.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !ensureStore(c, store) {
			return
		}

		var req updateAdminProfileReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}

		admin, ok, err := db.AdminGetFirst(store)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		if !ok {
			c.JSON(http.StatusPreconditionRequired, gin.H{"error": "needs setup"})
			return
		}

		if strings.TrimSpace(req.OldPassword) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing old_password"})
			return
		}

		if !password.Verify(admin.PasswordHash, req.OldPassword) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "old password mismatch"})
			return
		}

		newUsername := strings.TrimSpace(req.NewUsername)
		if newUsername == "" {
			newUsername = admin.Username
		}

		newPasswordHash := admin.PasswordHash
		hasPasswordChange := strings.TrimSpace(req.NewPassword) != "" || strings.TrimSpace(req.ConfirmPassword) != ""
		if hasPasswordChange {
			if req.NewPassword != req.ConfirmPassword {
				c.JSON(http.StatusBadRequest, gin.H{"error": "passwords do not match"})
				return
			}
			if len(req.NewPassword) < 8 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "password too short"})
				return
			}

			hashed, err := password.Hash(req.NewPassword)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid password"})
				return
			}
			newPasswordHash = hashed
		}

		if newUsername == admin.Username && newPasswordHash == admin.PasswordHash {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no changes"})
			return
		}

		if err := db.AdminUpdateCredentials(store, admin.ID, newUsername, newPasswordHash); err != nil {
			if errors.Is(err, db.ErrConflict) {
				c.JSON(http.StatusConflict, gin.H{"error": "username exists"})
				return
			}
			if errors.Is(err, db.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "admin not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": adminProfileDTO{Username: newUsername}})
	}
}
