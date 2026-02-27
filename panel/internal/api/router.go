package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"sboard/panel/internal/config"
	"sboard/panel/internal/db"
)

func NewRouter(cfg config.Config, store *db.Store) *gin.Engine {
	return NewRouterWithRPC(cfg, store, nil)
}

func NewRouterWithRPC(cfg config.Config, store *db.Store, rpcHandler http.Handler) *gin.Engine {
	if store != nil {
		_ = InitSystemTimezone(context.Background(), store)
	}

	r := gin.New()
	r.Use(RequestLogger(cfg.LogRequests))
	r.Use(gin.Recovery())
	r.Use(CORSMiddleware(cfg.CORSAllowOrigins))

	if rpcHandler == nil {
		r.GET("/api/health", Health)
		r.GET("/api/admin/bootstrap", AdminBootstrapGet(store))
		r.POST("/api/admin/bootstrap", AdminBootstrapPost(cfg, store))
		r.POST("/api/admin/login", AdminLogin(cfg, store))
		r.GET("/api/sub/:user_uuid", SubscriptionGet(store))

		authed := r.Group("/api")
		authed.Use(AuthMiddleware(cfg.JWTSecret))
		{
			authed.GET("/admin/profile", AdminProfileGet(store))
			authed.PUT("/admin/profile", AdminProfilePut(store))

			authed.GET("/users", UsersList(store))
			authed.POST("/users", UsersCreate(store))
			authed.GET("/users/:id", UsersGet(store))
			authed.PUT("/users/:id", UsersUpdate(store))
			authed.DELETE("/users/:id", UsersDelete(store))
			authed.GET("/users/:id/groups", UserGroupsGet(store))
			authed.PUT("/users/:id/groups", UserGroupsPut(store))

			authed.GET("/groups", GroupsList(store))
			authed.POST("/groups", GroupsCreate(store))
			authed.GET("/groups/:id", GroupsGet(store))
			authed.PUT("/groups/:id", GroupsUpdate(store))
			authed.DELETE("/groups/:id", GroupsDelete(store))
			authed.GET("/groups/:id/users", GroupUsersList(store))
			authed.PUT("/groups/:id/users", GroupUsersReplace(store))

			authed.GET("/nodes", NodesList(store))
			authed.POST("/nodes", NodesCreate(store))
			authed.GET("/nodes/:id", NodesGet(store))
			authed.PUT("/nodes/:id", NodesUpdate(store))
			authed.DELETE("/nodes/:id", NodesDelete(store))
			authed.GET("/nodes/:id/health", NodeHealth(store))
			authed.POST("/nodes/:id/sync", NodeSync(store))
			authed.GET("/nodes/:id/traffic", NodeTrafficList(store))

			authed.GET("/traffic/nodes/summary", TrafficNodesSummary(store))
			authed.GET("/traffic/total/summary", LegacyTrafficTotalSummary(store))
			authed.GET("/traffic/timeseries", TrafficTimeseries(store))

			authed.GET("/inbounds", InboundsList(store))
			authed.POST("/inbounds", InboundsCreate(store))
			authed.GET("/inbounds/:id", InboundsGet(store))
			authed.PUT("/inbounds/:id", InboundsUpdate(store))
			authed.DELETE("/inbounds/:id", InboundsDelete(store))

			authed.GET("/sync-jobs", SyncJobsList(store))
			authed.GET("/sync-jobs/:id", SyncJobsGet(store))
			authed.POST("/sync-jobs/:id/retry", SyncJobsRetry(store))

			singBoxTools := singBoxToolsFactory()
			authed.POST("/singbox/format", SingBoxFormat(singBoxTools))
			authed.POST("/sing-box/format", SingBoxFormat(singBoxTools))
			authed.POST("/singbox/check", SingBoxCheck(singBoxTools))
			authed.POST("/sing-box/check", SingBoxCheck(singBoxTools))
			authed.POST("/singbox/generate", SingBoxGenerate(singBoxTools))
			authed.POST("/sing-box/generate", SingBoxGenerate(singBoxTools))

			authed.GET("/system/info", SystemInfoGet())
			authed.GET("/system/settings", SystemSettingsGet(store))
			authed.PUT("/system/settings", SystemSettingsPut(store))
		}
	} else if store != nil {
		// Keep legacy subscription endpoint for client compatibility.
		r.GET("/api/sub/:user_uuid", SubscriptionGet(store))
	}

	if rpcHandler != nil {
		r.Any("/rpc", gin.WrapH(http.StripPrefix("/rpc", rpcHandler)))
		r.Any("/rpc/*path", gin.WrapH(http.StripPrefix("/rpc", rpcHandler)))
	}

	if cfg.ServeWeb {
		ServeWebUI(r, cfg.WebDir)
	}
	return r
}
