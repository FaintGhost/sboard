package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/api"
	"sboard/panel/internal/config"
	"sboard/panel/internal/singboxcli"
)

type fakeSingBoxTools struct {
	formatInput string
	checkInput  string
	genCommand  string

	formatOutput string
	checkOutput  string
	genOutput    string

	formatErr error
	checkErr  error
	genErr    error
}

func (f *fakeSingBoxTools) Format(_ context.Context, config string) (string, error) {
	f.formatInput = config
	if f.formatErr != nil {
		return "", f.formatErr
	}
	if f.formatOutput != "" {
		return f.formatOutput, nil
	}
	return config, nil
}

func (f *fakeSingBoxTools) Check(_ context.Context, config string) (string, error) {
	f.checkInput = config
	if f.checkErr != nil {
		return "", f.checkErr
	}
	if f.checkOutput != "" {
		return f.checkOutput, nil
	}
	return "ok", nil
}

func (f *fakeSingBoxTools) Generate(_ context.Context, kind string) (string, error) {
	f.genCommand = kind
	if f.genErr != nil {
		return "", f.genErr
	}
	if f.genOutput != "" {
		return f.genOutput, nil
	}
	return "generated", nil
}

func TestSingBoxFormat_WrapsInboundTemplateAndStripsPublicPort(t *testing.T) {
	fakeTools := &fakeSingBoxTools{}
	restore := api.SetSingBoxToolsFactoryForTest(func() api.SingBoxTools {
		return fakeTools
	})
	t.Cleanup(restore)

	cfg := config.Config{JWTSecret: "secret"}
	r := api.NewRouter(cfg, setupStore(t))
	token := mustToken(cfg.JWTSecret)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/sing-box/format", strings.NewReader(`{"config":"{\"type\":\"shadowsocks\",\"tag\":\"ss-in\",\"listen_port\":8388,\"public_port\":8388}","mode":"inbound"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, fakeTools.formatInput, `"inbounds"`)
	require.Contains(t, fakeTools.formatInput, `"outbounds"`)
	require.NotContains(t, fakeTools.formatInput, `"public_port"`)

	var resp struct {
		Data struct {
			Formatted string `json:"formatted"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Contains(t, resp.Data.Formatted, `"inbounds"`)
	require.Contains(t, resp.Data.Formatted, `"outbounds"`)
	require.Contains(t, resp.Data.Formatted, `"type":"shadowsocks"`)
}

func TestSingBoxFormat_FullConfigStripsPublicPortInInbounds(t *testing.T) {
	fakeTools := &fakeSingBoxTools{}
	restore := api.SetSingBoxToolsFactoryForTest(func() api.SingBoxTools {
		return fakeTools
	})
	t.Cleanup(restore)

	cfg := config.Config{JWTSecret: "secret"}
	r := api.NewRouter(cfg, setupStore(t))
	token := mustToken(cfg.JWTSecret)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/sing-box/format", strings.NewReader(`{"config":"{\"inbounds\":[{\"type\":\"shadowsocks\",\"tag\":\"ss-in\",\"listen_port\":8388,\"public_port\":8388}],\"outbounds\":[]}","mode":"inbound"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, fakeTools.formatInput, `"inbounds"`)
	require.NotContains(t, fakeTools.formatInput, `"public_port"`)
	require.Contains(t, fakeTools.formatInput, `"outbounds"`)
}

func TestSingBoxCheck_CommandErrorReturnsOkFalse(t *testing.T) {
	fakeTools := &fakeSingBoxTools{
		checkErr: errors.New("invalid config"),
	}
	restore := api.SetSingBoxToolsFactoryForTest(func() api.SingBoxTools {
		return fakeTools
	})
	t.Cleanup(restore)

	cfg := config.Config{JWTSecret: "secret"}
	r := api.NewRouter(cfg, setupStore(t))
	token := mustToken(cfg.JWTSecret)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/sing-box/check", strings.NewReader(`{"config":"{\"type\":\"vmess\",\"tag\":\"vmess-in\",\"listen_port\":443}"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Data struct {
			OK     bool   `json:"ok"`
			Output string `json:"output"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.False(t, resp.Data.OK)
	require.Contains(t, resp.Data.Output, "invalid config")
}

func TestSingBoxGenerate_ReturnsOutput(t *testing.T) {
	fakeTools := &fakeSingBoxTools{genOutput: "cbfd575f-5ff8-4f35-bad9-f2ee0a9335c8"}
	restore := api.SetSingBoxToolsFactoryForTest(func() api.SingBoxTools {
		return fakeTools
	})
	t.Cleanup(restore)

	cfg := config.Config{JWTSecret: "secret"}
	r := api.NewRouter(cfg, setupStore(t))
	token := mustToken(cfg.JWTSecret)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/sing-box/generate", strings.NewReader(`{"command":"uuid"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "uuid", fakeTools.genCommand)
	require.Contains(t, w.Body.String(), "cbfd575f-5ff8-4f35-bad9-f2ee0a9335c8")
}

func TestSingBoxGenerate_InvalidCommand(t *testing.T) {
	fakeTools := &fakeSingBoxTools{genErr: singboxcli.ErrInvalidGenerateKind}
	restore := api.SetSingBoxToolsFactoryForTest(func() api.SingBoxTools {
		return fakeTools
	})
	t.Cleanup(restore)

	cfg := config.Config{JWTSecret: "secret"}
	r := api.NewRouter(cfg, setupStore(t))
	token := mustToken(cfg.JWTSecret)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/sing-box/generate", strings.NewReader(`{"command":"tls-keypair"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Contains(t, w.Body.String(), "invalid generate command")
}
