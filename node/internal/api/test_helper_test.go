package api_test

import (
	"github.com/gin-gonic/gin"
	"sboard/node/internal/api"
)

func newTestRouter(core api.Core, provider api.InboundTrafficProvider) *gin.Engine {
	gin.SetMode(gin.TestMode)
	return api.NewRouter("secret", core, provider)
}
