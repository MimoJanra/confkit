package confkit

import (
	"context"
	"testing"
	"time"
)

func TestNewRotationEngine(t *testing.T) {
	strategy := RotateOnInterval(1 * time.Hour)
	engine := NewRotationEngine(strategy)

	if engine == nil {
		t.Fatal("Expected non-nil engine")
	}

	if len(engine.callbacks) != 0 {
		t.Errorf("Expected 0 callbacks, got %d", len(engine.callbacks))
	}
}

func TestRotationEngineAddCallback(t *testing.T) {
	engine := NewRotationEngine(RotateOnInterval(1 * time.Hour))

	engine.AddCallback(func(oldCfg, newCfg any, err error) {
	})

	if len(engine.callbacks) != 1 {
		t.Errorf("Expected 1 callback, got %d", len(engine.callbacks))
	}
}

func TestIntervalRotationStrategy(t *testing.T) {
	tests := []struct {
		name            string
		interval        time.Duration
		timeSinceRotate time.Duration
		shouldRotate    bool
	}{
		{"Should rotate after interval", 1 * time.Hour, 2 * time.Hour, true},
		{"Should not rotate before interval", 1 * time.Hour, 30 * time.Minute, false},
		{"Should rotate at exact interval", 1 * time.Hour, 1 * time.Hour, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := RotateOnInterval(tt.interval)
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
	strategy := RotateOnEvent(eventChan)

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
	tests := []struct {
		name            string
		minTTL          time.Duration
		timeSinceRotate time.Duration
		shouldRotate    bool
	}{
		{"Should rotate after TTL", 1 * time.Hour, 2 * time.Hour, true},
		{"Should not rotate before TTL", 1 * time.Hour, 30 * time.Minute, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strat := RotateOnMinTTL(tt.minTTL)
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
	engine := NewRotationEngine(RotateOnInterval(1 * time.Hour))

	if engine.IsRotating() {
		t.Error("Should not be rotating initially")
	}
}

func TestRotationEngineLastRotation(t *testing.T) {
	before := time.Now()
	engine := NewRotationEngine(RotateOnInterval(1 * time.Hour))

	if engine.lastRotation.Before(before) {
		t.Error("Last rotation should be recent")
	}
}

func TestMultipleCallbacks(t *testing.T) {
	engine := NewRotationEngine(RotateOnInterval(1 * time.Hour))

	count := 0
	engine.AddCallback(func(oldCfg, newCfg any, err error) {
		count++
	})
	engine.AddCallback(func(oldCfg, newCfg any, err error) {
		count++
	})

	if len(engine.callbacks) != 2 {
		t.Errorf("Expected 2 callbacks, got %d", len(engine.callbacks))
	}
}
