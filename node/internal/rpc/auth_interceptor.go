package rpc

import (
	"context"
	"errors"
	"strings"

	"connectrpc.com/connect"
	nodev1connect "sboard/node/internal/rpc/gen/sboard/node/v1/nodev1connect"
)

func authInterceptor(secret string) connect.UnaryInterceptorFunc {
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
