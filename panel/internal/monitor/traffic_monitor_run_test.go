package monitor

import (
	"context"
	"testing"
	"time"
)

func TestTrafficMonitorRun_IntervalNonPositiveReturnsImmediately(t *testing.T) {
	m := NewTrafficMonitor(nil, nil)
	start := time.Now()
	m.Run(context.Background(), 0)
	if time.Since(start) > 100*time.Millisecond {
		t.Fatal("Run with non-positive interval should return immediately")
	}
}

func TestTrafficMonitorRun_StopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	m := NewTrafficMonitor(nil, nil)

	done := make(chan struct{})
	go func() {
		defer close(done)
		m.Run(ctx, 5*time.Millisecond)
	}()

	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case <-done:
		return
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Run did not stop after context cancellation")
	}
}
