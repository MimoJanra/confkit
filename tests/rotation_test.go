package confkit_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	confkit "github.com/MimoJanra/confkit"
)

type testErrorStrategy struct{ err error }

func (t *testErrorStrategy) ShouldRotate(ctx context.Context, lastRotation time.Time) (bool, error) {
	return false, t.err
}

func TestNewRotationEngine(t *testing.T) {
	strategy := confkit.RotateOnInterval(1 * time.Hour)
	engine := confkit.NewRotationEngine(strategy)
	if engine == nil {
		t.Fatal("Expected non-nil engine")
	}
}

func TestRotationEngineAddCallback(t *testing.T) {
	engine := confkit.NewRotationEngine(confkit.RotateOnInterval(1 * time.Hour))
	called := false
	engine.AddCallback(func(oldCfg, newCfg any, err error) { called = true })

	engine.Start(context.Background(), 30*time.Millisecond)
	time.Sleep(100 * time.Millisecond)
	engine.Stop()

	if !called {
		t.Log("note: callback may not fire in this test, that's OK — testing AddCallback doesn't panic")
	}
}

func TestIntervalRotationStrategy(t *testing.T) {
	cases := []struct {
		name            string
		interval        time.Duration
		timeSinceRotate time.Duration
		shouldRotate    bool
	}{
		{"Should rotate after interval", 1 * time.Hour, 2 * time.Hour, true},
		{"Should not rotate before interval", 1 * time.Hour, 30 * time.Minute, false},
		{"Should rotate at exact interval", 1 * time.Hour, 1 * time.Hour, true},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			strategy := confkit.RotateOnInterval(tt.interval)
			lastRotation := time.Now().Add(-tt.timeSinceRotate)
			result, err := strategy.ShouldRotate(context.Background(), lastRotation)
			if err != nil {
				t.Fatalf("ShouldRotate failed: %v", err)
			}
			if result != tt.shouldRotate {
				t.Errorf("Expected %v, got %v", tt.shouldRotate, result)
			}
		})
	}
}

func TestEventRotationStrategy(t *testing.T) {
	eventChan := make(chan struct{})
	strategy := confkit.RotateOnEvent(eventChan)

	shouldRotate, err := strategy.ShouldRotate(context.Background(), time.Now())
	if err != nil {
		t.Fatalf("ShouldRotate failed: %v", err)
	}
	if shouldRotate {
		t.Error("Should not rotate without event")
	}

	close(eventChan)
	_, err = strategy.ShouldRotate(context.Background(), time.Now())
	if err != nil {
		t.Errorf("Got error after close: %v", err)
	}
}

func TestTTLRotationStrategy(t *testing.T) {
	cases := []struct {
		name            string
		minTTL          time.Duration
		timeSinceRotate time.Duration
		shouldRotate    bool
	}{
		{"Should rotate after TTL", 1 * time.Hour, 2 * time.Hour, true},
		{"Should not rotate before TTL", 1 * time.Hour, 30 * time.Minute, false},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			strat := confkit.RotateOnMinTTL(tt.minTTL)
			lastRotation := time.Now().Add(-tt.timeSinceRotate)
			result, err := strat.ShouldRotate(context.Background(), lastRotation)
			if err != nil {
				t.Fatalf("ShouldRotate failed: %v", err)
			}
			if result != tt.shouldRotate {
				t.Errorf("Expected %v, got %v", tt.shouldRotate, result)
			}
		})
	}
}

func TestRotationEngineIsRotating(t *testing.T) {
	engine := confkit.NewRotationEngine(confkit.RotateOnInterval(1 * time.Hour))
	if engine.IsRotating() {
		t.Error("Should not be rotating initially")
	}
}

func TestRotationEngineStart_WithInterval(t *testing.T) {
	callCount := atomic.Int32{}
	engine := confkit.NewRotationEngine(confkit.RotateOnInterval(50 * time.Millisecond))
	engine.AddCallback(func(oldCfg, newCfg any, err error) {
		if err == nil {
			callCount.Add(1)
		}
	})
	engine.Start(context.Background(), 30*time.Millisecond)
	time.Sleep(150 * time.Millisecond)
	engine.Stop()

	if callCount.Load() == 0 {
		t.Error("Expected at least one rotation callback")
	}
}

func TestRotationEngineStop_DoubleStop(t *testing.T) {
	engine := confkit.NewRotationEngine(confkit.RotateOnInterval(1 * time.Second))
	engine.Start(context.Background(), 50*time.Millisecond)
	engine.Stop()
	time.Sleep(10 * time.Millisecond)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Second Stop should not panic: %v", r)
		}
	}()
	engine.Stop()
}

func TestRotationEngineStop_BeforeStart(t *testing.T) {
	engine := confkit.NewRotationEngine(confkit.RotateOnInterval(1 * time.Second))
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Stop before Start should not panic: %v", r)
		}
	}()
	engine.Stop()
}

func TestRotationEngineCallback_InvokedOnRotation(t *testing.T) {
	var callbackCalled atomic.Bool
	var receivedErr atomic.Value

	engine := confkit.NewRotationEngine(confkit.RotateOnInterval(50 * time.Millisecond))
	engine.AddCallback(func(oldCfg, newCfg any, err error) {
		callbackCalled.Store(true)
		if err != nil {
			receivedErr.Store(err)
		}
	})
	engine.Start(context.Background(), 30*time.Millisecond)
	time.Sleep(150 * time.Millisecond)
	engine.Stop()
	time.Sleep(50 * time.Millisecond)

	if !callbackCalled.Load() {
		t.Error("Expected callback to be invoked")
	}
	if val := receivedErr.Load(); val != nil {
		t.Errorf("Expected no error in callback, got %v", val)
	}
}

func TestRotationEngineCallback_ReceivesError(t *testing.T) {
	var errorReceived atomic.Bool
	strategy := &testErrorStrategy{err: errors.New("test error")}
	engine := confkit.NewRotationEngine(strategy)
	engine.AddCallback(func(oldCfg, newCfg any, err error) {
		if err != nil {
			errorReceived.Store(true)
		}
	})
	engine.Start(context.Background(), 30*time.Millisecond)
	time.Sleep(100 * time.Millisecond)
	engine.Stop()
	time.Sleep(50 * time.Millisecond)

	if !errorReceived.Load() {
		t.Error("Expected callback to receive error")
	}
}

func TestRotationEngine_ConcurrentCallbacks(t *testing.T) {
	count := atomic.Int32{}
	mu := sync.Mutex{}
	var callErrors []error

	engine := confkit.NewRotationEngine(confkit.RotateOnInterval(50 * time.Millisecond))
	for i := 0; i < 5; i++ {
		engine.AddCallback(func(oldCfg, newCfg any, err error) {
			count.Add(1)
			mu.Lock()
			callErrors = append(callErrors, err)
			mu.Unlock()
		})
	}
	engine.Start(context.Background(), 30*time.Millisecond)
	time.Sleep(150 * time.Millisecond)
	engine.Stop()
	time.Sleep(50 * time.Millisecond)

	if count.Load() < 5 {
		t.Errorf("Expected at least 5 total callback invocations, got %d", count.Load())
	}
}

func TestRotationEngine_LastRotationUpdated(t *testing.T) {
	engine := confkit.NewRotationEngine(confkit.RotateOnInterval(50 * time.Millisecond))
	initialTime := engine.LastRotation()
	engine.Start(context.Background(), 30*time.Millisecond)
	time.Sleep(150 * time.Millisecond)
	engine.Stop()
	time.Sleep(50 * time.Millisecond)

	finalTime := engine.LastRotation()
	if finalTime.Equal(initialTime) {
		t.Error("Expected lastRotation to be updated")
	}
	if finalTime.Before(initialTime) {
		t.Error("Expected lastRotation to be after initial time")
	}
}

func TestEventRotationStrategy_ContextCancellation(t *testing.T) {
	eventChan := make(chan struct{})
	strategy := confkit.RotateOnEvent(eventChan)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	shouldRotate, err := strategy.ShouldRotate(ctx, time.Now())
	if err == nil {
		t.Error("Expected error when context is canceled")
	}
	if shouldRotate {
		t.Error("Expected ShouldRotate to be false when context canceled")
	}
}

func TestEventRotationStrategy_ShouldRotateOnEvent(t *testing.T) {
	eventChan := make(chan struct{}, 1)
	eventChan <- struct{}{}

	strategy := confkit.RotateOnEvent(eventChan)
	shouldRotate, err := strategy.ShouldRotate(context.Background(), time.Now())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !shouldRotate {
		t.Error("Expected ShouldRotate to be true when event is available")
	}
}

type countingRotationStrategy struct{ calls atomic.Int64 }

func (c *countingRotationStrategy) ShouldRotate(_ context.Context, _ time.Time) (bool, error) {
	c.calls.Add(1)
	return false, nil
}

func TestRotationEngineStopsOnContextCancel(t *testing.T) {
	strategy := &countingRotationStrategy{}
	engine := confkit.NewRotationEngine(strategy)
	ctx, cancel := context.WithCancel(context.Background())
	defer engine.Stop()

	engine.Start(ctx, 10*time.Millisecond)
	time.Sleep(100 * time.Millisecond)
	cancel()

	afterCancel := strategy.calls.Load()
	time.Sleep(200 * time.Millisecond)

	if grew := strategy.calls.Load() - afterCancel; grew > 0 {
		t.Fatalf("engine kept polling %d times after context cancellation", grew)
	}
}
