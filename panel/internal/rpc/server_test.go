package rpc

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
	"sboard/panel/internal/config"
	panelv1 "sboard/panel/internal/rpc/gen/sboard/panel/v1"
	panelv1connect "sboard/panel/internal/rpc/gen/sboard/panel/v1/panelv1connect"
)

const testJWTSecret = "test-rpc-auth-secret"

// bearerInterceptor returns a connect client interceptor that injects the
// given Bearer token into every outgoing request.
func bearerInterceptor(token string) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("Authorization", "Bearer "+token)
			return next(ctx, req)
		}
	}
}

// mustJWT creates a valid admin JWT signed with the given secret.
func mustJWT(secret string) string {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   "admin",
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		panic(err)
	}
	return signed
}

// setupAuthTestServer creates an httptest.Server backed by the RPC handler
// with the standard auth interceptor. The store is real (migrated SQLite) so
// that service methods can execute without nil-pointer panics on public
// endpoints.
func setupAuthTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	store := setupRPCStore(t)
	cfg := config.Config{JWTSecret: testJWTSecret}
	handler := NewHandler(cfg, store)
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestRPCAuth_PublicEndpointsNoToken(t *testing.T) {
	srv := setupAuthTestServer(t)
	ctx := context.Background()

	// Health endpoint — no auth required.
	healthClient := panelv1connect.NewHealthServiceClient(srv.Client(), srv.URL)
	resp, err := healthClient.GetHealth(ctx, &panelv1.GetHealthRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Auth/GetBootstrapStatus — no auth required.
	authClient := panelv1connect.NewAuthServiceClient(srv.Client(), srv.URL)
	bsResp, err := authClient.GetBootstrapStatus(ctx, &panelv1.GetBootstrapStatusRequest{})
	require.NoError(t, err)
	require.NotNil(t, bsResp)

	// Auth/Login — no auth required (will fail with invalid creds but the
	// error should NOT be Unauthenticated from the interceptor; it should be
	// a business-level error).
	_, loginErr := authClient.Login(ctx, &panelv1.LoginRequest{
		Username: "nonexistent",
		Password: "wrong",
	})
	// Login may return an error because no admin exists, but it must NOT be
	// CodeUnauthenticated from the auth interceptor.  The interceptor skips
	// auth for this procedure, so any error comes from the service impl.
	if loginErr != nil {
		var connErr *connect.Error
		if errors.As(loginErr, &connErr) {
			require.NotEqual(t, connect.CodeUnauthenticated, connErr.Code(),
				"Login endpoint should not be blocked by auth interceptor")
		}
	}
}

func TestRPCAuth_ProtectedEndpointNoToken(t *testing.T) {
	srv := setupAuthTestServer(t)
	ctx := context.Background()

	// ListUsers is a protected endpoint. Call without any token.
	userClient := panelv1connect.NewUserServiceClient(srv.Client(), srv.URL)
	_, err := userClient.ListUsers(ctx, &panelv1.ListUsersRequest{})
	require.Error(t, err)

	var connErr *connect.Error
	require.ErrorAs(t, err, &connErr)
	require.Equal(t, connect.CodeUnauthenticated, connErr.Code())
}

func TestRPCAuth_ProtectedEndpointInvalidToken(t *testing.T) {
	srv := setupAuthTestServer(t)
	ctx := context.Background()

	// Create a client that sends an invalid token.
	opts := connect.WithInterceptors(bearerInterceptor("this-is-not-a-valid-jwt"))
	userClient := panelv1connect.NewUserServiceClient(srv.Client(), srv.URL, opts)
	_, err := userClient.ListUsers(ctx, &panelv1.ListUsersRequest{})
	require.Error(t, err)

	var connErr *connect.Error
	require.ErrorAs(t, err, &connErr)
	require.Equal(t, connect.CodeUnauthenticated, connErr.Code())
}

func TestRPCAuth_ProtectedEndpointValidToken(t *testing.T) {
	srv := setupAuthTestServer(t)
	ctx := context.Background()

	token := mustJWT(testJWTSecret)
	opts := connect.WithInterceptors(bearerInterceptor(token))
	userClient := panelv1connect.NewUserServiceClient(srv.Client(), srv.URL, opts)

	resp, err := userClient.ListUsers(ctx, &panelv1.ListUsersRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
}
