package rpc

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/require"
	"sboard/panel/internal/db"
	nodev1 "sboard/panel/internal/rpc/gen/sboard/node/v1"
	nodev1connect "sboard/panel/internal/rpc/gen/sboard/node/v1/nodev1connect"
)

type nodeControlErrorStub struct {
	syncErr error
}

func (s *nodeControlErrorStub) Health(ctx context.Context, req *nodev1.HealthRequest) (*nodev1.HealthResponse, error) {
	return &nodev1.HealthResponse{Status: "ok"}, nil
}

func (s *nodeControlErrorStub) SyncConfig(ctx context.Context, req *nodev1.SyncConfigRequest) (*nodev1.SyncConfigResponse, error) {
	if s.syncErr != nil {
		return nil, s.syncErr
	}
	return &nodev1.SyncConfigResponse{Status: "ok"}, nil
}

func (s *nodeControlErrorStub) GetTraffic(ctx context.Context, req *nodev1.GetTrafficRequest) (*nodev1.GetTrafficResponse, error) {
	return &nodev1.GetTrafficResponse{}, nil
}

func (s *nodeControlErrorStub) GetInboundTraffic(ctx context.Context, req *nodev1.GetInboundTrafficRequest) (*nodev1.GetInboundTrafficResponse, error) {
	return &nodev1.GetInboundTrafficResponse{}, nil
}

func TestNodeSyncRPCErrorMapping(t *testing.T) {
	ctx := context.Background()

	t.Run("unauthenticated", func(t *testing.T) {
		srv := newNodeRPCErrorServer(t, &nodeControlErrorStub{}, connect.WithInterceptors(testNodeAuthInterceptor("expected-secret")))
		defer srv.Close()

		store := setupRPCStore(t)
		nodeItem := seedSyncNodeFixture(t, ctx, store, srv.URL, "wrong-secret")
		result := (&Server{store: store}).runNodeSync(ctx, nodeItem, rpcTriggerManualNodeSync, nil)

		assertSyncFailure(t, ctx, store, nodeItem.ID, result, 401, "unauthorized")
	})

	t.Run("invalid_argument", func(t *testing.T) {
		srv := newNodeRPCErrorServer(t, &nodeControlErrorStub{
			syncErr: connect.NewError(connect.CodeInvalidArgument, errors.New("invalid payload")),
		})
		defer srv.Close()

		store := setupRPCStore(t)
		nodeItem := seedSyncNodeFixture(t, ctx, store, srv.URL, "secret")
		result := (&Server{store: store}).runNodeSync(ctx, nodeItem, rpcTriggerManualNodeSync, nil)

		assertSyncFailure(t, ctx, store, nodeItem.ID, result, 400, "invalid payload")
		require.False(t, result.Retryable)
	})

	t.Run("unavailable", func(t *testing.T) {
		port := closedLocalPort(t)
		store := setupRPCStore(t)
		nodeItem := seedSyncNodeFixture(t, ctx, store, "http://127.0.0.1:"+strconv.Itoa(port), "secret")
		result := (&Server{store: store}).runNodeSync(ctx, nodeItem, rpcTriggerManualNodeSync, nil)

		assertSyncFailure(t, ctx, store, nodeItem.ID, result, 502, "node sync")
		require.True(t, result.Retryable)
	})
}

func newNodeRPCErrorServer(t *testing.T, stub *nodeControlErrorStub, opts ...connect.HandlerOption) *httptest.Server {
	t.Helper()

	_, handler := nodev1connect.NewNodeControlServiceHandler(stub, opts...)
	mux := http.NewServeMux()
	mux.Handle("/rpc/", http.StripPrefix("/rpc", handler))
	return httptest.NewServer(mux)
}

func testNodeAuthInterceptor(secret string) connect.UnaryInterceptorFunc {
	public := map[string]bool{
		nodev1connect.NodeControlServiceHealthProcedure: true,
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
			token := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
			if token == "" || token != strings.TrimSpace(secret) {
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("unauthorized"))
			}
			return next(ctx, req)
		}
	}
}

func seedSyncNodeFixture(t *testing.T, ctx context.Context, store *db.Store, rawURL string, secret string) db.Node {
	t.Helper()

	parsedURL, err := url.Parse(rawURL)
	require.NoError(t, err)
	host, portRaw, err := net.SplitHostPort(parsedURL.Host)
	require.NoError(t, err)
	port, err := strconv.Atoi(portRaw)
	require.NoError(t, err)

	group, err := store.CreateGroup(ctx, "g-1", "")
	require.NoError(t, err)

	user, err := store.CreateUser(ctx, "alice")
	require.NoError(t, err)
	require.NoError(t, store.ReplaceGroupUsers(ctx, group.ID, []int64{user.ID}))

	groupID := group.ID
	nodeItem, err := store.CreateNode(ctx, db.NodeCreate{
		Name:          "node-a",
		APIAddress:    parsedURL.Scheme + "://" + host,
		APIPort:       port,
		SecretKey:     secret,
		PublicAddress: host,
		GroupID:       &groupID,
	})
	require.NoError(t, err)

	_, err = store.CreateInbound(ctx, db.InboundCreate{
		NodeID:     nodeItem.ID,
		Tag:        "vless-in",
		Protocol:   "vless",
		ListenPort: 443,
		PublicPort: 443,
		Settings:   []byte(`{}`),
	})
	require.NoError(t, err)

	return nodeItem
}

func assertSyncFailure(t *testing.T, ctx context.Context, store *db.Store, nodeID int64, result syncResult, wantHTTPStatus int, wantErrContains string) {
	t.Helper()

	require.Equal(t, "error", result.Status)
	require.NotNil(t, result.Error)
	require.Contains(t, strings.TrimSpace(*result.Error), wantErrContains)

	jobs, err := store.ListSyncJobs(ctx, db.SyncJobsListFilter{NodeID: nodeID, Limit: 10})
	require.NoError(t, err)
	require.Len(t, jobs, 1)
	require.Equal(t, db.SyncJobStatusFailed, jobs[0].Status)
	require.Contains(t, jobs[0].ErrorSummary, wantErrContains)

	attempts, err := store.ListSyncAttemptsByJobID(ctx, jobs[0].ID)
	require.NoError(t, err)
	require.Len(t, attempts, 1)
	require.Equal(t, db.SyncAttemptStatusFailed, attempts[0].Status)
	require.Equal(t, wantHTTPStatus, attempts[0].HTTPStatus)
	require.Contains(t, attempts[0].ErrorSummary, wantErrContains)
}

func closedLocalPort(t *testing.T) int {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := ln.Addr().(*net.TCPAddr).Port
	require.NoError(t, ln.Close())
	return port
}
