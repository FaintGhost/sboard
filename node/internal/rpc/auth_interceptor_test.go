package rpc

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/require"
	nodev1 "sboard/node/internal/rpc/gen/sboard/node/v1"
	nodev1connect "sboard/node/internal/rpc/gen/sboard/node/v1/nodev1connect"
)

type authTestApplier struct{}

func (authTestApplier) ApplySyncPayload(ctx context.Context, body []byte) error {
	return nil
}

func TestNodeControlAuth(t *testing.T) {
	srv := httptest.NewServer(NewHandler("expected-secret", authTestApplier{}, inboundProviderStub{}))
	defer srv.Close()

	t.Run("health is public", func(t *testing.T) {
		client := nodev1connect.NewNodeControlServiceClient(srv.Client(), srv.URL)
		resp, err := client.Health(context.Background(), &nodev1.HealthRequest{})
		require.NoError(t, err)
		require.Equal(t, "ok", resp.GetStatus())
	})

	t.Run("missing bearer token", func(t *testing.T) {
		client := nodev1connect.NewNodeControlServiceClient(srv.Client(), srv.URL)
		_, err := client.SyncConfig(context.Background(), &nodev1.SyncConfigRequest{PayloadJson: []byte(`{"inbounds":[]}`)})
		require.Error(t, err)

		var connErr *connect.Error
		require.True(t, errors.As(err, &connErr))
		require.Equal(t, connect.CodeUnauthenticated, connErr.Code())
	})

	t.Run("wrong bearer token", func(t *testing.T) {
		auth := connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
			return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
				req.Header().Set("Authorization", "Bearer wrong-secret")
				return next(ctx, req)
			}
		})
		client := nodev1connect.NewNodeControlServiceClient(srv.Client(), srv.URL, connect.WithInterceptors(auth))
		_, err := client.SyncConfig(context.Background(), &nodev1.SyncConfigRequest{PayloadJson: []byte(`{"inbounds":[]}`)})
		require.Error(t, err)

		var connErr *connect.Error
		require.True(t, errors.As(err, &connErr))
		require.Equal(t, connect.CodeUnauthenticated, connErr.Code())
	})
}
