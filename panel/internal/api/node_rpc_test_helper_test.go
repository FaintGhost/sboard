package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"

	nodev1 "sboard/panel/internal/rpc/gen/sboard/node/v1"
	nodev1connect "sboard/panel/internal/rpc/gen/sboard/node/v1/nodev1connect"
)

type nodeRPCServiceStub struct {
	healthFunc         func(context.Context, *nodev1.HealthRequest) (*nodev1.HealthResponse, error)
	syncConfigFunc     func(context.Context, *nodev1.SyncConfigRequest) (*nodev1.SyncConfigResponse, error)
	getTrafficFunc     func(context.Context, *nodev1.GetTrafficRequest) (*nodev1.GetTrafficResponse, error)
	inboundTrafficFunc func(context.Context, *nodev1.GetInboundTrafficRequest) (*nodev1.GetInboundTrafficResponse, error)
}

func (s nodeRPCServiceStub) Health(ctx context.Context, req *nodev1.HealthRequest) (*nodev1.HealthResponse, error) {
	if s.healthFunc != nil {
		return s.healthFunc(ctx, req)
	}
	return &nodev1.HealthResponse{Status: "ok"}, nil
}

func (s nodeRPCServiceStub) SyncConfig(ctx context.Context, req *nodev1.SyncConfigRequest) (*nodev1.SyncConfigResponse, error) {
	if s.syncConfigFunc != nil {
		return s.syncConfigFunc(ctx, req)
	}
	return &nodev1.SyncConfigResponse{Status: "ok"}, nil
}

func (s nodeRPCServiceStub) GetTraffic(ctx context.Context, req *nodev1.GetTrafficRequest) (*nodev1.GetTrafficResponse, error) {
	if s.getTrafficFunc != nil {
		return s.getTrafficFunc(ctx, req)
	}
	return &nodev1.GetTrafficResponse{}, nil
}

func (s nodeRPCServiceStub) GetInboundTraffic(ctx context.Context, req *nodev1.GetInboundTrafficRequest) (*nodev1.GetInboundTrafficResponse, error) {
	if s.inboundTrafficFunc != nil {
		return s.inboundTrafficFunc(ctx, req)
	}
	return &nodev1.GetInboundTrafficResponse{}, nil
}

func serveNodeRPCRequest(req *http.Request, svc nodeRPCServiceStub, inspect func(*http.Request)) (*http.Response, error) {
	_, handler := nodev1connect.NewNodeControlServiceHandler(svc)
	mux := http.NewServeMux()
	mux.Handle("/rpc/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if inspect != nil {
			inspect(r)
		}
		http.StripPrefix("/rpc", handler).ServeHTTP(w, r)
	}))

	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, req)
	return recorder.Result(), nil
}
