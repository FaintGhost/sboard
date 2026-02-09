package api

import (
	"github.com/gin-gonic/gin"
	"sboard/panel/internal/config"
	"sboard/panel/internal/db"
)

func NewRouter(cfg config.Config, store *db.Store) *gin.Engine {
	r := gin.New()
	r.Use(RequestLogger(cfg.LogRequests))
	r.Use(gin.Recovery())
	r.Use(CORSMiddleware(cfg.CORSAllowOrigins))
	r.GET("/api/health", Health)
	r.GET("/api/admin/bootstrap", AdminBootstrapGet(store))
	r.POST("/api/admin/bootstrap", AdminBootstrapPost(cfg, store))
	r.POST("/api/admin/login", AdminLogin(cfg, store))
	r.GET("/api/sub/:user_uuid", SubscriptionGet(store))

	singBoxTools := singBoxToolsFactory()

	auth := r.Group("/api")
	auth.Use(AuthMiddleware(cfg.JWTSecret))
	auth.GET("/users", UsersList(store))
	auth.POST("/users", UsersCreate(store))
	auth.GET("/users/:id", UsersGet(store))
	auth.PUT("/users/:id", UsersUpdate(store))
	auth.DELETE("/users/:id", UsersDelete(store))
	auth.GET("/users/:id/groups", UserGroupsGet(store))
	auth.PUT("/users/:id/groups", UserGroupsPut(store))
	auth.GET("/groups", GroupsList(store))
	auth.POST("/groups", GroupsCreate(store))
	auth.GET("/groups/:id", GroupsGet(store))
	auth.PUT("/groups/:id", GroupsUpdate(store))
	auth.DELETE("/groups/:id", GroupsDelete(store))
	auth.GET("/groups/:id/users", GroupUsersList(store))
	auth.PUT("/groups/:id/users", GroupUsersReplace(store))
	auth.GET("/nodes", NodesList(store))
	auth.POST("/nodes", NodesCreate(store))
	auth.GET("/nodes/:id", NodesGet(store))
	auth.PUT("/nodes/:id", NodesUpdate(store))
	auth.DELETE("/nodes/:id", NodesDelete(store))
	auth.GET("/nodes/:id/health", NodeHealth(store))
	auth.POST("/nodes/:id/sync", NodeSync(store))
	auth.GET("/nodes/:id/traffic", NodeTrafficList(store))
	auth.GET("/traffic/nodes/summary", TrafficNodesSummary(store))
	auth.GET("/traffic/total/summary", TrafficTotalSummary(store))
	auth.GET("/traffic/timeseries", TrafficTimeseries(store))
	auth.GET("/system/info", SystemInfoGet())
	auth.GET("/sync-jobs", SyncJobsList(store))
	auth.GET("/sync-jobs/:id", SyncJobsGet(store))
	auth.POST("/sync-jobs/:id/retry", SyncJobsRetry(store))
	auth.GET("/inbounds", InboundsList(store))
	auth.POST("/inbounds", InboundsCreate(store))
	auth.GET("/inbounds/:id", InboundsGet(store))
	auth.PUT("/inbounds/:id", InboundsUpdate(store))
	auth.DELETE("/inbounds/:id", InboundsDelete(store))
	auth.POST("/sing-box/format", SingBoxFormat(singBoxTools))
	auth.POST("/sing-box/check", SingBoxCheck(singBoxTools))
	auth.POST("/sing-box/generate", SingBoxGenerate(singBoxTools))

	if cfg.ServeWeb {
		ServeWebUI(r, cfg.WebDir)
	}
	return r
}
