package api_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	sync2 "sboard/node/internal/sync"
)

type failingReadCloser struct{}

func (f *failingReadCloser) Read(_ []byte) (int, error) {
	return 0, errors.New("read failed")
}

func (f *failingReadCloser) Close() error {
	return nil
}

func TestConfigSyncBranches(t *testing.T) {
	t.Run("core nil", func(t *testing.T) {
		r := newTestRouter(nil, nil)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/config/sync", strings.NewReader(`{"inbounds":[]}`))
		req.Header.Set("Authorization", "Bearer secret")
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusInternalServerError, w.Code)
		require.Contains(t, w.Body.String(), "core not ready")
	})

	t.Run("invalid body", func(t *testing.T) {
		r := newTestRouter(&fakeCore{}, nil)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/config/sync", nil)
		req.Header.Set("Authorization", "Bearer secret")
		req.Body = &failingReadCloser{}
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusBadRequest, w.Code)
		require.Contains(t, w.Body.String(), "invalid body")
	})

	t.Run("bad request error -> 400", func(t *testing.T) {
		r := newTestRouter(&fakeCore{err: sync2.BadRequestError{Message: "bad config"}}, nil)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/config/sync", strings.NewReader(`{"inbounds":[]}`))
		req.Header.Set("Authorization", "Bearer secret")
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusBadRequest, w.Code)
		require.Contains(t, w.Body.String(), "bad config")
	})

	t.Run("generic error -> 500", func(t *testing.T) {
		r := newTestRouter(&fakeCore{err: errors.New("boom")}, nil)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/config/sync", strings.NewReader(`{"inbounds":[]}`))
		req.Header.Set("Authorization", "Bearer secret")
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusInternalServerError, w.Code)
		require.Contains(t, w.Body.String(), "boom")
	})

	t.Run("success", func(t *testing.T) {
		r := newTestRouter(&fakeCore{}, nil)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/config/sync", io.NopCloser(strings.NewReader(`{"inbounds":[]}`)))
		req.Header.Set("Authorization", "Bearer secret")
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Body.String(), "ok")
	})
}
