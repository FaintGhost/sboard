package node

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/db"
)

type fakeDoer struct {
	do func(req *http.Request) (*http.Response, error)
}

func (f fakeDoer) Do(req *http.Request) (*http.Response, error) {
	return f.do(req)
}

func jsonResp(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestNewClient_DefaultDoer(t *testing.T) {
	c := NewClient(nil)
	require.NotNil(t, c)
	require.NotNil(t, c.doer)
}

func TestBuildURL(t *testing.T) {
	n := db.Node{APIAddress: "10.0.0.2", APIPort: 3900, PublicAddress: "pub.example"}
	require.Equal(t, "http://10.0.0.2:3900/api/health", buildURL(n, "/api/health"))
	require.Equal(t, "http://10.0.0.2:3900/api/health", buildURL(n, "api/health"))

	n = db.Node{APIAddress: "", PublicAddress: "pub.example", APIPort: 0}
	require.Equal(t, "http://pub.example:3000/api/health", buildURL(n, "/api/health"))

	n = db.Node{}
	require.Equal(t, "http://127.0.0.1:3000/api/health", buildURL(n, "/api/health"))
}

func TestClientHealth(t *testing.T) {
	node := db.Node{APIAddress: "127.0.0.1", APIPort: 3000}

	t.Run("request error", func(t *testing.T) {
		c := NewClient(fakeDoer{do: func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("network down")
		}})

		err := c.Health(context.Background(), node)
		require.ErrorContains(t, err, "network down")
	})

	t.Run("non-2xx", func(t *testing.T) {
		c := NewClient(fakeDoer{do: func(req *http.Request) (*http.Response, error) {
			require.Equal(t, http.MethodGet, req.Method)
			require.Equal(t, "http://127.0.0.1:3000/api/health", req.URL.String())
			return jsonResp(http.StatusBadGateway, "bad gateway"), nil
		}})

		err := c.Health(context.Background(), node)
		require.ErrorContains(t, err, "node health status 502")
		require.ErrorContains(t, err, "bad gateway")
	})

	t.Run("ok", func(t *testing.T) {
		c := NewClient(fakeDoer{do: func(req *http.Request) (*http.Response, error) {
			return jsonResp(http.StatusOK, `{}`), nil
		}})

		require.NoError(t, c.Health(context.Background(), node))
	})
}

func TestClientTraffic(t *testing.T) {
	node := db.Node{APIAddress: "127.0.0.1", APIPort: 3000, SecretKey: "secret"}

	t.Run("non-2xx", func(t *testing.T) {
		c := NewClient(fakeDoer{do: func(req *http.Request) (*http.Response, error) {
			require.Equal(t, "Bearer secret", req.Header.Get("Authorization"))
			return jsonResp(http.StatusUnauthorized, "unauthorized"), nil
		}})

		_, err := c.Traffic(context.Background(), node)
		require.ErrorContains(t, err, "node traffic status 401")
	})

	t.Run("invalid json", func(t *testing.T) {
		c := NewClient(fakeDoer{do: func(req *http.Request) (*http.Response, error) {
			return jsonResp(http.StatusOK, `{"rx_bytes":"bad"}`), nil
		}})

		_, err := c.Traffic(context.Background(), node)
		require.Error(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		c := NewClient(fakeDoer{do: func(req *http.Request) (*http.Response, error) {
			return jsonResp(http.StatusOK, `{"interface":"eth0","rx_bytes":123,"tx_bytes":456,"at":"2026-02-10T12:00:00Z"}`), nil
		}})

		out, err := c.Traffic(context.Background(), node)
		require.NoError(t, err)
		require.Equal(t, "eth0", out.Interface)
		require.Equal(t, uint64(123), out.RxBytes)
		require.Equal(t, uint64(456), out.TxBytes)
		require.Equal(t, time.Date(2026, 2, 10, 12, 0, 0, 0, time.UTC), out.At)
	})
}

func TestClientInboundTraffic(t *testing.T) {
	node := db.Node{APIAddress: "127.0.0.1", APIPort: 3000, SecretKey: "secret"}

	t.Run("with reset", func(t *testing.T) {
		c := NewClient(fakeDoer{do: func(req *http.Request) (*http.Response, error) {
			require.Equal(t, "reset=1", req.URL.RawQuery)
			return jsonResp(http.StatusOK, `{"data":[{"tag":"ss-in","user":"alice","uplink":10,"downlink":20,"at":"2026-02-10T12:00:00Z"}],"reset":true}`), nil
		}})

		rows, err := c.InboundTraffic(context.Background(), node, true)
		require.NoError(t, err)
		require.Len(t, rows, 1)
		require.Equal(t, "ss-in", rows[0].Tag)
	})

	t.Run("non-2xx", func(t *testing.T) {
		c := NewClient(fakeDoer{do: func(req *http.Request) (*http.Response, error) {
			return jsonResp(http.StatusInternalServerError, "sync failed"), nil
		}})

		_, err := c.InboundTraffic(context.Background(), node, false)
		require.ErrorContains(t, err, "node inbound traffic status 500")
	})

	t.Run("decode error", func(t *testing.T) {
		c := NewClient(fakeDoer{do: func(req *http.Request) (*http.Response, error) {
			return jsonResp(http.StatusOK, `{"data":"bad"}`), nil
		}})

		_, err := c.InboundTraffic(context.Background(), node, false)
		require.Error(t, err)
	})
}

func TestClientInboundTrafficWithMeta(t *testing.T) {
	node := db.Node{APIAddress: "127.0.0.1", APIPort: 3000, SecretKey: "secret"}

	c := NewClient(fakeDoer{do: func(req *http.Request) (*http.Response, error) {
		return jsonResp(http.StatusOK, `{"data":[{"tag":"ss-in","user":"alice","uplink":10,"downlink":20,"at":"2026-02-10T12:00:00Z"}],"meta":{"tracked_tags":3,"tcp_conns":8,"udp_conns":9}}`), nil
	}})

	rows, meta, err := c.InboundTrafficWithMeta(context.Background(), node, false)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.NotNil(t, meta)
	require.Equal(t, 3, meta.TrackedTags)
	require.Equal(t, int64(8), meta.TCPConns)
	require.Equal(t, int64(9), meta.UDPConns)
}

func TestClientSyncConfig(t *testing.T) {
	node := db.Node{APIAddress: "127.0.0.1", APIPort: 3000, SecretKey: "secret"}

	t.Run("marshal error", func(t *testing.T) {
		c := NewClient(fakeDoer{do: func(req *http.Request) (*http.Response, error) {
			t.Fatal("doer should not be called when marshal fails")
			return nil, nil
		}})
		err := c.SyncConfig(context.Background(), node, map[string]any{"bad": make(chan int)})
		require.Error(t, err)
	})

	t.Run("non-2xx", func(t *testing.T) {
		c := NewClient(fakeDoer{do: func(req *http.Request) (*http.Response, error) {
			require.Equal(t, http.MethodPost, req.Method)
			require.Equal(t, "application/json", req.Header.Get("Content-Type"))
			require.Equal(t, "Bearer secret", req.Header.Get("Authorization"))
			body, _ := io.ReadAll(req.Body)
			require.JSONEq(t, `{"inbounds":[]}`, string(body))
			return jsonResp(http.StatusBadRequest, "invalid payload"), nil
		}})

		err := c.SyncConfig(context.Background(), node, map[string]any{"inbounds": []any{}})
		require.ErrorContains(t, err, "node sync status 400")
	})

	t.Run("ok", func(t *testing.T) {
		c := NewClient(fakeDoer{do: func(req *http.Request) (*http.Response, error) {
			return jsonResp(http.StatusOK, `{"status":"ok"}`), nil
		}})

		require.NoError(t, c.SyncConfig(context.Background(), node, map[string]any{"inbounds": []any{}}))
	})
}
