package stats

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestReadNetDev_ValidateInput(t *testing.T) {
	_, _, err := ReadNetDev("")
	require.ErrorContains(t, err, "missing interface")

	_, _, err = ReadNetDev("definitely-not-a-real-iface-12345")
	require.Error(t, err)
	require.Contains(t, strings.ToLower(err.Error()), "not found")
}

func TestCurrentSample_InvalidInterface(t *testing.T) {
	_, err := CurrentSample("definitely-not-a-real-iface-12345")
	require.Error(t, err)
}

func TestDetectDefaultInterfaceAndCurrentSample(t *testing.T) {
	iface, err := DetectDefaultInterface()
	if err != nil {
		t.Skipf("skip: detect default iface failed in current env: %v", err)
	}

	require.NotEmpty(t, iface)
	sample, err := CurrentSample(iface)
	require.NoError(t, err)
	require.Equal(t, iface, sample.Interface)
	require.WithinDuration(t, time.Now().UTC(), sample.At, 3*time.Second)
}
