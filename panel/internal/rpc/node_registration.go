package rpc

import (
	"context"
	"crypto/subtle"
	"errors"
	"net"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"
	"sboard/panel/internal/db"
	panelv1 "sboard/panel/internal/rpc/gen/sboard/panel/v1"
)

func (s *Server) Heartbeat(ctx context.Context, req *panelv1.NodeHeartbeatRequest) (*panelv1.NodeHeartbeatResponse, error) {
	node, err := s.store.GetNodeByUUID(ctx, req.GetUuid())
	if err != nil {
		if !errors.Is(err, db.ErrNotFound) {
			return nil, connect.NewError(connect.CodeInternal, errors.New("failed to look up node"))
		}

		// Unknown node -- create a pending record.
		host, port := parseAPIAddr(req.GetApiAddr())
		_, createErr := s.store.CreatePendingNode(ctx, db.PendingNodeParams{
			UUID:       req.GetUuid(),
			SecretKey:  req.GetSecretKey(),
			APIAddress: host,
			APIPort:    port,
		})
		if createErr != nil {
			return nil, connect.NewError(connect.CodeInternal, errors.New("failed to register pending node"))
		}
		return &panelv1.NodeHeartbeatResponse{
			Status:  panelv1.NodeHeartbeatStatus_NODE_HEARTBEAT_STATUS_PENDING,
			Message: "node registered as pending",
		}, nil
	}

	// Known node -- verify secret key using constant-time comparison.
	if subtle.ConstantTimeCompare([]byte(node.SecretKey), []byte(req.GetSecretKey())) != 1 {
		return &panelv1.NodeHeartbeatResponse{
			Status:  panelv1.NodeHeartbeatStatus_NODE_HEARTBEAT_STATUS_REJECTED,
			Message: "invalid secret key",
		}, nil
	}

	// Secret key matches -- update last seen and return recognized.
	if err := s.store.UpdateNodeLastSeen(ctx, node.UUID, time.Now()); err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to update last seen"))
	}
	return &panelv1.NodeHeartbeatResponse{
		Status:  panelv1.NodeHeartbeatStatus_NODE_HEARTBEAT_STATUS_RECOGNIZED,
		Message: "ok",
	}, nil
}

// parseAPIAddr splits an address string into host and port.
// It accepts "host:port" or a bare host. The default port is 3000.
func parseAPIAddr(addr string) (string, int) {
	const defaultPort = 3000
	if addr == "" {
		return "", defaultPort
	}
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		// No port in the address -- use the default.
		return addr, defaultPort
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return host, defaultPort
	}
	return host, port
}

func (s *Server) ApproveNode(ctx context.Context, req *panelv1.ApproveNodeRequest) (*panelv1.ApproveNodeResponse, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("name is required"))
	}

	// Check node exists first to distinguish "not found" from "not pending".
	_, err := s.store.GetNodeByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("node not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	params := db.ApproveNodeParams{
		Name:          req.Name,
		GroupID:       req.GroupId,
		PublicAddress: req.GetPublicAddress(),
	}

	node, err := s.store.ApproveNode(ctx, req.Id, params)
	if err != nil {
		if errors.Is(err, db.ErrNotPending) {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("node is not pending"))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return &panelv1.ApproveNodeResponse{Data: mapDBNode(node)}, nil
}

func (s *Server) RejectNode(ctx context.Context, req *panelv1.RejectNodeRequest) (*panelv1.RejectNodeResponse, error) {
	// Check node exists and is pending.
	node, err := s.store.GetNodeByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("node not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if node.Status != "pending" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("node is not pending"))
	}

	if err := s.store.DeleteNode(ctx, req.Id); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return &panelv1.RejectNodeResponse{}, nil
}
