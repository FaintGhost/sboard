package rpc

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"connectrpc.com/connect"
	nodev1 "sboard/node/internal/rpc/gen/sboard/node/v1"
	"sboard/node/internal/stats"
	syncsvc "sboard/node/internal/sync"
)

const maxSyncPayloadBytes = 4 << 20 // 4 MiB

func (s *Server) Health(ctx context.Context, req *nodev1.HealthRequest) (*nodev1.HealthResponse, error) {
	return &nodev1.HealthResponse{Status: "ok"}, nil
}

func (s *Server) SyncConfig(ctx context.Context, req *nodev1.SyncConfigRequest) (*nodev1.SyncConfigResponse, error) {
	if s.applier == nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("core not ready"))
	}
	payload := req.GetPayloadJson()
	if len(payload) == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid body"))
	}
	if len(payload) > maxSyncPayloadBytes {
		return nil, connect.NewError(connect.CodeResourceExhausted, errors.New("body too large"))
	}

	if err := s.applier.ApplySyncPayload(ctx, payload); err != nil {
		var bre syncsvc.BadRequestError
		if errors.As(err, &bre) {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return &nodev1.SyncConfigResponse{Status: "ok"}, nil
}

func (s *Server) GetTraffic(ctx context.Context, req *nodev1.GetTrafficRequest) (*nodev1.GetTrafficResponse, error) {
	iface := strings.TrimSpace(req.GetInterface())
	if iface == "" {
		iface = strings.TrimSpace(os.Getenv("NODE_TRAFFIC_INTERFACE"))
	}
	sample, err := stats.CurrentSample(iface)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return &nodev1.GetTrafficResponse{
		Interface: sample.Interface,
		RxBytes:   sample.RxBytes,
		TxBytes:   sample.TxBytes,
		At:        sample.At.UTC().Format(time.RFC3339),
	}, nil
}

func (s *Server) GetInboundTraffic(ctx context.Context, req *nodev1.GetInboundTrafficRequest) (*nodev1.GetInboundTrafficResponse, error) {
	if s.inbound == nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("stats not ready"))
	}
	rows := s.inbound.InboundTraffic(req.GetReset_())
	out := make([]*nodev1.InboundTraffic, 0, len(rows))
	for _, row := range rows {
		out = append(out, &nodev1.InboundTraffic{
			Tag:      row.Tag,
			User:     row.User,
			Uplink:   row.Uplink,
			Downlink: row.Downlink,
			At:       row.At.UTC().Format(time.RFC3339),
		})
	}

	meta := s.inbound.InboundTrafficMeta()
	return &nodev1.GetInboundTrafficResponse{
		Data:   out,
		Reset_: req.GetReset_(),
		Meta: &nodev1.InboundTrafficMeta{
			TrackedTags: int32(meta.TrackedTags),
			TcpConns:    meta.TCPConns,
			UdpConns:    meta.UDPConns,
		},
	}, nil
}
