package node

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/require"
	"sboard/panel/internal/db"
	nodev1 "sboard/panel/internal/rpc/gen/sboard/node/v1"
	nodev1connect "sboard/panel/internal/rpc/gen/sboard/node/v1/nodev1connect"
)

type fakeDoer struct {
	do func(req *http.Request) (*http.Response, error)
}

func (f fakeDoer) Do(req *http.Request) (*http.Response, error) {
	return f.do(req)
}

type nodeControlServiceTestServer struct {
	health         func(context.Context, *nodev1.HealthRequest) (*nodev1.HealthResponse, error)
	syncConfig     func(context.Context, *nodev1.SyncConfigRequest) (*nodev1.SyncConfigResponse, error)
	getTraffic     func(context.Context, *nodev1.GetTrafficRequest) (*nodev1.GetTrafficResponse, error)
	inboundTraffic func(context.Context, *nodev1.GetInboundTrafficRequest) (*nodev1.GetInboundTrafficResponse, error)
}

func (s nodeControlServiceTestServer) Health(ctx context.Context, req *nodev1.HealthRequest) (*nodev1.HealthResponse, error) {
	if s.health != nil {
		return s.health(ctx, req)
	}
	return &nodev1.HealthResponse{Status: "ok"}, nil
}

func (s nodeControlServiceTestServer) SyncConfig(ctx context.Context, req *nodev1.SyncConfigRequest) (*nodev1.SyncConfigResponse, error) {
	if s.syncConfig != nil {
		return s.syncConfig(ctx, req)
	}
	return &nodev1.SyncConfigResponse{Status: "ok"}, nil
}

func (s nodeControlServiceTestServer) GetTraffic(ctx context.Context, req *nodev1.GetTrafficRequest) (*nodev1.GetTrafficResponse, error) {
	if s.getTraffic != nil {
		return s.getTraffic(ctx, req)
	}
	return &nodev1.GetTrafficResponse{}, nil
}

func (s nodeControlServiceTestServer) GetInboundTraffic(ctx context.Context, req *nodev1.GetInboundTrafficRequest) (*nodev1.GetInboundTrafficResponse, error) {
	if s.inboundTraffic != nil {
		return s.inboundTraffic(ctx, req)
	}
	return &nodev1.GetInboundTrafficResponse{}, nil
}

func newNodeRPCServer(t *testing.T, svc nodeControlServiceTestServer, inspect func(*http.Request)) *httptest.Server {
	t.Helper()

	_, handler := nodev1connect.NewNodeControlServiceHandler(svc)
	mux := http.NewServeMux()
	mux.Handle("/rpc/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if inspect != nil {
			inspect(r)
		}
		http.StripPrefix("/rpc", handler).ServeHTTP(w, r)
	}))
	return httptest.NewServer(mux)
}

func TestNewClient_DefaultDoer(t *testing.T) {
	c := NewClient(nil)
	require.NotNil(t, c)
	require.NotNil(t, c.doer)
}

func TestBuildURL(t *testing.T) {
	n := db.Node{APIAddress: "10.0.0.2", APIPort: 3900, PublicAddress: "pub.example"}
	require.Equal(t, "http://10.0.0.2:3900/rpc", buildRPCBaseURL(n))

	n = db.Node{APIAddress: "", PublicAddress: "pub.example", APIPort: 0}
	require.Equal(t, "http://pub.example:3000/rpc", buildRPCBaseURL(n))

	n = db.Node{}
	require.Equal(t, "http://127.0.0.1:3000/rpc", buildRPCBaseURL(n))
}

func TestClientHealth(t *testing.T) {
	t.Run("request error", func(t *testing.T) {
		c := NewClient(fakeDoer{do: func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("network down")
		}})

		err := c.Health(context.Background(), db.Node{APIAddress: "127.0.0.1", APIPort: 3000})
		require.ErrorContains(t, err, "network down")
	})

	t.Run("rpc error", func(t *testing.T) {
		srv := newNodeRPCServer(t, nodeControlServiceTestServer{
			health: func(ctx context.Context, req *nodev1.HealthRequest) (*nodev1.HealthResponse, error) {
				return nil, connect.NewError(connect.CodeUnavailable, errors.New("bad gateway"))
			},
		}, nil)
		defer srv.Close()

		node := nodeFromServerURL(t, srv.URL)
		err := NewClient(srv.Client()).Health(context.Background(), node)
		require.ErrorContains(t, err, "node health status 502")
		require.ErrorContains(t, err, "bad gateway")
	})

	t.Run("ok", func(t *testing.T) {
		inspected := false
		srv := newNodeRPCServer(t, nodeControlServiceTestServer{}, func(req *http.Request) {
			inspected = true
			require.Equal(t, http.MethodPost, req.Method)
		})
		defer srv.Close()

		node := nodeFromServerURL(t, srv.URL)
		require.NoError(t, NewClient(srv.Client()).Health(context.Background(), node))
		require.True(t, inspected)
	})
}

func TestClientTraffic(t *testing.T) {
	srv := newNodeRPCServer(t, nodeControlServiceTestServer{
		getTraffic: func(ctx context.Context, req *nodev1.GetTrafficRequest) (*nodev1.GetTrafficResponse, error) {
			return &nodev1.GetTrafficResponse{
				Interface: "eth0",
				RxBytes:   123,
				TxBytes:   456,
				At:        "2026-02-10T12:00:00Z",
			}, nil
		},
	}, nil)
	defer srv.Close()

	out, err := NewClient(srv.Client()).Traffic(context.Background(), nodeFromServerURL(t, srv.URL))
	require.NoError(t, err)
	require.Equal(t, "eth0", out.Interface)
	require.Equal(t, uint64(123), out.RxBytes)
	require.Equal(t, uint64(456), out.TxBytes)
	require.Equal(t, time.Date(2026, 2, 10, 12, 0, 0, 0, time.UTC), out.At)
}

func TestClientInboundTraffic(t *testing.T) {
	srv := newNodeRPCServer(t, nodeControlServiceTestServer{
		inboundTraffic: func(ctx context.Context, req *nodev1.GetInboundTrafficRequest) (*nodev1.GetInboundTrafficResponse, error) {
			require.True(t, req.GetReset_())
			return &nodev1.GetInboundTrafficResponse{
				Data: []*nodev1.InboundTraffic{{
					Tag:      "ss-in",
					User:     "alice",
					Uplink:   10,
					Downlink: 20,
					At:       "2026-02-10T12:00:00Z",
				}},
				Reset_: true,
			}, nil
		},
	}, nil)
	defer srv.Close()

	rows, err := NewClient(srv.Client()).InboundTraffic(context.Background(), nodeFromServerURL(t, srv.URL), true)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, "ss-in", rows[0].Tag)
}

func TestClientInboundTrafficWithMeta(t *testing.T) {
	srv := newNodeRPCServer(t, nodeControlServiceTestServer{
		inboundTraffic: func(ctx context.Context, req *nodev1.GetInboundTrafficRequest) (*nodev1.GetInboundTrafficResponse, error) {
			return &nodev1.GetInboundTrafficResponse{
				Data: []*nodev1.InboundTraffic{{
					Tag:      "ss-in",
					User:     "alice",
					Uplink:   10,
					Downlink: 20,
					At:       "2026-02-10T12:00:00Z",
				}},
				Meta: &nodev1.InboundTrafficMeta{TrackedTags: 3, TcpConns: 8, UdpConns: 9},
			}, nil
		},
	}, nil)
	defer srv.Close()

	rows, meta, err := NewClient(srv.Client()).InboundTrafficWithMeta(context.Background(), nodeFromServerURL(t, srv.URL), false)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.NotNil(t, meta)
	require.Equal(t, 3, meta.TrackedTags)
	require.Equal(t, int64(8), meta.TCPConns)
	require.Equal(t, int64(9), meta.UDPConns)
}

func TestClientSyncConfig(t *testing.T) {
	t.Run("marshal error", func(t *testing.T) {
		c := NewClient(fakeDoer{do: func(req *http.Request) (*http.Response, error) {
			t.Fatal("doer should not be called when marshal fails")
			return nil, nil
		}})
		err := c.SyncConfig(context.Background(), db.Node{APIAddress: "127.0.0.1", APIPort: 3000}, map[string]any{"bad": make(chan int)})
		require.Error(t, err)
	})

	t.Run("rpc error", func(t *testing.T) {
		srv := newNodeRPCServer(t, nodeControlServiceTestServer{
			syncConfig: func(ctx context.Context, req *nodev1.SyncConfigRequest) (*nodev1.SyncConfigResponse, error) {
				require.JSONEq(t, `{"inbounds":[]}`, string(req.GetPayloadJson()))
				return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid payload"))
			},
		}, nil)
		defer srv.Close()

		err := NewClient(srv.Client()).SyncConfig(context.Background(), nodeFromServerURL(t, srv.URL), map[string]any{"inbounds": []any{}})
		require.ErrorContains(t, err, "node sync status 400")
	})

	t.Run("ok", func(t *testing.T) {
		srv := newNodeRPCServer(t, nodeControlServiceTestServer{}, nil)
		defer srv.Close()

		require.NoError(t, NewClient(srv.Client()).SyncConfig(context.Background(), nodeFromServerURL(t, srv.URL), map[string]any{"inbounds": []any{}}))
	})

	t.Run("rpc not found", func(t *testing.T) {
		c := NewClient(fakeDoer{do: func(req *http.Request) (*http.Response, error) {
			require.True(t, strings.HasPrefix(req.URL.Path, "/rpc/"))
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Header:     http.Header{"Content-Type": []string{"text/plain"}},
				Body:       ioNopCloser("not found"),
			}, nil
		}})

		err := c.SyncConfig(context.Background(), db.Node{APIAddress: "127.0.0.1", APIPort: 3000, SecretKey: "secret"}, map[string]any{"inbounds": []any{}})
		require.Error(t, err)

		var syncErr *SyncError
		require.ErrorAs(t, err, &syncErr)
		require.Equal(t, http.StatusNotImplemented, syncErr.HTTPStatus())
		require.Equal(t, connect.CodeUnimplemented, syncErr.Code())
		require.ErrorContains(t, err, "node sync status 501")
	})
}

func nodeFromServerURL(t *testing.T, rawURL string) db.Node {
	t.Helper()
	trimmed := strings.TrimPrefix(rawURL, "http://")
	parts := strings.Split(trimmed, ":")
	require.Len(t, parts, 2)
	port, err := strconv.Atoi(parts[1])
	require.NoError(t, err)
	return db.Node{APIAddress: parts[0], APIPort: port, SecretKey: "secret"}
}

func ioNopCloser(body string) *readCloser { return &readCloser{Reader: strings.NewReader(body)} }

type readCloser struct{ *strings.Reader }

func (r *readCloser) Close() error { return nil }
