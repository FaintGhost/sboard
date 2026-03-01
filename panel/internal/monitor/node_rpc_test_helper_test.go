package monitor

import (
	"context"
	"net/http"
	"net/http/httptest"

	nodev1 "sboard/panel/internal/rpc/gen/sboard/node/v1"
	nodev1connect "sboard/panel/internal/rpc/gen/sboard/node/v1/nodev1connect"
)

type nodeRPCServiceStub struct {
	inboundTrafficFunc func(context.Context, *nodev1.GetInboundTrafficRequest) (*nodev1.GetInboundTrafficResponse, error)
}

func (s nodeRPCServiceStub) Health(context.Context, *nodev1.HealthRequest) (*nodev1.HealthResponse, error) {
	return &nodev1.HealthResponse{Status: "ok"}, nil
}

func (s nodeRPCServiceStub) SyncConfig(context.Context, *nodev1.SyncConfigRequest) (*nodev1.SyncConfigResponse, error) {
	return &nodev1.SyncConfigResponse{Status: "ok"}, nil
}

func (s nodeRPCServiceStub) GetTraffic(context.Context, *nodev1.GetTrafficRequest) (*nodev1.GetTrafficResponse, error) {
	return &nodev1.GetTrafficResponse{}, nil
}

func (s nodeRPCServiceStub) GetInboundTraffic(ctx context.Context, req *nodev1.GetInboundTrafficRequest) (*nodev1.GetInboundTrafficResponse, error) {
	if s.inboundTrafficFunc != nil {
		return s.inboundTrafficFunc(ctx, req)
	}
	return &nodev1.GetInboundTrafficResponse{}, nil
}

func serveNodeRPCRequest(req *http.Request, svc nodeRPCServiceStub) (*http.Response, error) {
	_, handler := nodev1connect.NewNodeControlServiceHandler(svc)
	mux := http.NewServeMux()
	mux.Handle("/rpc/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/rpc", handler).ServeHTTP(w, r)
	}))

	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, req)
	return recorder.Result(), nil
}
