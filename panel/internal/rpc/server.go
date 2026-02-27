package rpc

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt/v5"
	"sboard/panel/internal/api"
	"sboard/panel/internal/config"
	"sboard/panel/internal/db"
	panelv1 "sboard/panel/internal/rpc/gen/sboard/panel/v1"
	panelv1connect "sboard/panel/internal/rpc/gen/sboard/panel/v1/panelv1connect"
)

var _ panelv1connect.AuthServiceHandler = (*Server)(nil)
var _ panelv1connect.HealthServiceHandler = (*Server)(nil)
var _ panelv1connect.SystemServiceHandler = (*Server)(nil)
var _ panelv1connect.UserServiceHandler = (*Server)(nil)
var _ panelv1connect.GroupServiceHandler = (*Server)(nil)
var _ panelv1connect.NodeServiceHandler = (*Server)(nil)
var _ panelv1connect.TrafficServiceHandler = (*Server)(nil)
var _ panelv1connect.InboundServiceHandler = (*Server)(nil)
var _ panelv1connect.SyncJobServiceHandler = (*Server)(nil)
var _ panelv1connect.SingBoxToolServiceHandler = (*Server)(nil)

type Server struct {
	cfg    config.Config
	store  *db.Store
	legacy *api.Server
}

func NewHandler(cfg config.Config, store *db.Store) http.Handler {
	mux := http.NewServeMux()
	s := &Server{cfg: cfg, store: store, legacy: api.NewServer(store, cfg, nil)}
	opts := []connect.HandlerOption{
		connect.WithInterceptors(authInterceptor(cfg.JWTSecret)),
	}

	path, h := panelv1connect.NewAuthServiceHandler(s, opts...)
	mux.Handle(path, h)
	path, h = panelv1connect.NewHealthServiceHandler(s, opts...)
	mux.Handle(path, h)
	path, h = panelv1connect.NewSystemServiceHandler(s, opts...)
	mux.Handle(path, h)
	path, h = panelv1connect.NewUserServiceHandler(s, opts...)
	mux.Handle(path, h)
	path, h = panelv1connect.NewGroupServiceHandler(s, opts...)
	mux.Handle(path, h)
	path, h = panelv1connect.NewNodeServiceHandler(s, opts...)
	mux.Handle(path, h)
	path, h = panelv1connect.NewTrafficServiceHandler(s, opts...)
	mux.Handle(path, h)
	path, h = panelv1connect.NewInboundServiceHandler(s, opts...)
	mux.Handle(path, h)
	path, h = panelv1connect.NewSyncJobServiceHandler(s, opts...)
	mux.Handle(path, h)
	path, h = panelv1connect.NewSingBoxToolServiceHandler(s, opts...)
	mux.Handle(path, h)

	return mux
}

func authInterceptor(secret string) connect.UnaryInterceptorFunc {
	public := map[string]bool{
		panelv1connect.AuthServiceGetBootstrapStatusProcedure: true,
		panelv1connect.AuthServiceBootstrapProcedure:          true,
		panelv1connect.AuthServiceLoginProcedure:              true,
		panelv1connect.HealthServiceGetHealthProcedure:        true,
	}

	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if public[req.Spec().Procedure] {
				return next(ctx, req)
			}

			auth := strings.TrimSpace(req.Header().Get("Authorization"))
			if !strings.HasPrefix(auth, "Bearer ") {
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("unauthorized"))
			}
			tokenStr := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
			claims := &jwt.RegisteredClaims{}
			token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
				if t.Method != jwt.SigningMethodHS256 {
					return nil, errors.New("unexpected signing method")
				}
				return []byte(secret), nil
			})
			if err != nil || !token.Valid || claims.Subject != "admin" {
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("unauthorized"))
			}

			return next(ctx, req)
		}
	}
}

func connectErrorFromHTTP(status int, msg string) error {
	m := strings.TrimSpace(msg)
	if m == "" {
		m = http.StatusText(status)
	}

	code := connect.CodeInternal
	switch status {
	case http.StatusBadRequest:
		code = connect.CodeInvalidArgument
	case http.StatusUnauthorized:
		code = connect.CodeUnauthenticated
	case http.StatusNotFound:
		code = connect.CodeNotFound
	case http.StatusConflict:
		code = connect.CodeFailedPrecondition
	case http.StatusPreconditionRequired:
		code = connect.CodeFailedPrecondition
	case http.StatusBadGateway:
		code = connect.CodeUnavailable
	case http.StatusInternalServerError:
		code = connect.CodeInternal
	}
	return connect.NewError(code, errors.New(m))
}

func limitParam(v int32) *api.LimitParam {
	if v <= 0 {
		return nil
	}
	n := api.LimitParam(v)
	return &n
}

func offsetParam(v int32) *api.OffsetParam {
	n := api.OffsetParam(v)
	return &n
}

func mapUser(u api.User) *panelv1.User {
	out := &panelv1.User{
		Id:              u.Id,
		Uuid:            u.Uuid,
		Username:        u.Username,
		GroupIds:        append([]int64{}, u.GroupIds...),
		TrafficLimit:    u.TrafficLimit,
		TrafficUsed:     u.TrafficUsed,
		TrafficResetDay: int32(u.TrafficResetDay),
		Status:          u.Status,
	}
	if u.ExpireAt != nil {
		v := u.ExpireAt.Format("2006-01-02T15:04:05Z07:00")
		out.ExpireAt = &v
	}
	return out
}

func mapGroup(g api.Group) *panelv1.Group {
	return &panelv1.Group{
		Id:          g.Id,
		Name:        g.Name,
		Description: g.Description,
		MemberCount: g.MemberCount,
	}
}

func mapNode(n api.Node) *panelv1.Node {
	out := &panelv1.Node{
		Id:            n.Id,
		Uuid:          n.Uuid,
		Name:          n.Name,
		ApiAddress:    n.ApiAddress,
		ApiPort:       int32(n.ApiPort),
		SecretKey:     n.SecretKey,
		PublicAddress: n.PublicAddress,
		Status:        n.Status,
		GroupId:       n.GroupId,
	}
	if n.LastSeenAt != nil {
		v := n.LastSeenAt.Format("2006-01-02T15:04:05Z07:00")
		out.LastSeenAt = &v
	}
	return out
}
