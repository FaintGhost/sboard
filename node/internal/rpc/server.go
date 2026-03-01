package rpc

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	nodev1connect "sboard/node/internal/rpc/gen/sboard/node/v1/nodev1connect"
	"sboard/node/internal/stats"
)

var _ nodev1connect.NodeControlServiceHandler = (*Server)(nil)

type SyncApplier interface {
	ApplySyncPayload(ctx context.Context, body []byte) error
}

type InboundTrafficProvider interface {
	InboundTraffic(reset bool) []stats.InboundTraffic
	InboundTrafficMeta() stats.InboundTrafficMeta
}

type Server struct {
	secret  string
	applier SyncApplier
	inbound InboundTrafficProvider
}

func NewHandler(secret string, applier SyncApplier, inbound InboundTrafficProvider) http.Handler {
	mux := http.NewServeMux()
	s := &Server{secret: secret, applier: applier, inbound: inbound}
	opts := []connect.HandlerOption{
		connect.WithInterceptors(authInterceptor(secret)),
	}
	path, h := nodev1connect.NewNodeControlServiceHandler(s, opts...)
	mux.Handle(path, h)
	return mux
}
