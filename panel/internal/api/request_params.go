package api

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	defaultListLimit = 50
	maxListLimit     = 500
)

func parseLimitOffset(c *gin.Context) (int, int, error) {
	limitStr := c.DefaultQuery("limit", strconv.Itoa(defaultListLimit))
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 0 || limit > maxListLimit {
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
