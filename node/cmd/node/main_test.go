package main

import (
	"testing"

	"github.com/stretchr/testify/require"
	"sboard/node/internal/sync"
)

func TestCoreAdapterApplyRawConfig_InvalidPayload(t *testing.T) {
	a := &coreAdapter{sbctx: sync.NewSingboxContext()}
	err := a.applyRawConfig([]byte(`{`))
	require.Error(t, err)
}

func TestCoreAdapterApplyConfig_InvalidPayload(t *testing.T) {
	a := &coreAdapter{sbctx: sync.NewSingboxContext(), statePath: t.TempDir() + "/state.json"}
	err := a.ApplyConfig(nil, []byte(`not-json`))
	require.Error(t, err)
}
