package confkit

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type RotationStrategy interface {
	ShouldRotate(ctx context.Context, lastRotation time.Time) (bool, error)
}

type RotationCallback func(oldCfg, newCfg any, err error)

type RotationEngine struct {
	strategy      RotationStrategy
	callbacks     []RotationCallback
	callbackMutex sync.RWMutex
	lastRotation  time.Time
	ticker        *time.Ticker
	stopChan      chan struct{}
	mu            sync.RWMutex
	isRotating    atomic.Bool
	startOnce     sync.Once
	stopOnce      sync.Once
}

func NewRotationEngine(strategy RotationStrategy) *RotationEngine {
	return &RotationEngine{
		strategy:     strategy,
		callbacks:    make([]RotationCallback, 0),
		lastRotation: time.Now(),
		stopChan:     make(chan struct{}),
	}
}

func (r *RotationEngine) AddCallback(cb RotationCallback) {
	r.callbackMutex.Lock()
	defer r.callbackMutex.Unlock()
	r.callbacks = append(r.callbacks, cb)
}

// Start begins the rotation loop. Subsequent calls are no-ops.
// The loop stops when Stop is called or when ctx is cancelled.
func (r *RotationEngine) Start(ctx context.Context, interval time.Duration) {
	r.startOnce.Do(func() {
		r.ticker = time.NewTicker(interval)
		go r.run(ctx)
	})
}

func (r *RotationEngine) Stop() {
	r.stopOnce.Do(func() {
		if r.ticker != nil {
			r.ticker.Stop()
		}
		close(r.stopChan)
	})
}

func (r *RotationEngine) run(ctx context.Context) {
	// Also stop the ticker when the loop exits via ctx cancellation, where
	// Stop() may never be called. Stopping a ticker twice is safe.
	defer r.ticker.Stop()

	for {
		select {
		case <-r.stopChan:
			r.isRotating.Store(false)
			return
		case <-ctx.Done():
			r.isRotating.Store(false)
			return
		case <-r.ticker.C:
			r.mu.RLock()
			lastRot := r.lastRotation
			r.mu.RUnlock()

			shouldRotate, err := r.strategy.ShouldRotate(ctx, lastRot)
			if err != nil {
				r.notifyCallbacks(nil, nil, fmt.Errorf("rotation check failed: %w", err))
				continue
			}

			if shouldRotate {
				r.isRotating.Store(true)

				r.mu.Lock()
				r.lastRotation = time.Now()
				r.mu.Unlock()

				r.notifyCallbacks(nil, nil, nil)

				r.isRotating.Store(false)
			}
		}
	}
}

func (r *RotationEngine) notifyCallbacks(oldCfg, newCfg any, err error) {
	r.callbackMutex.RLock()
	callbacks := make([]RotationCallback, len(r.callbacks))
	copy(callbacks, r.callbacks)
	r.callbackMutex.RUnlock()

	for _, cb := range callbacks {
		go cb(oldCfg, newCfg, err)
	}
}

func (r *RotationEngine) IsRotating() bool {
	return r.isRotating.Load()
}

func (r *RotationEngine) LastRotation() time.Time {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.lastRotation
}

type IntervalRotationStrategy struct {
	interval time.Duration
}

func RotateOnInterval(interval time.Duration) RotationStrategy {
	return &IntervalRotationStrategy{interval: interval}
}

func (i *IntervalRotationStrategy) ShouldRotate(_ context.Context, lastRotation time.Time) (bool, error) {
	return time.Since(lastRotation) >= i.interval, nil
}

type EventRotationStrategy struct {
	eventChan <-chan struct{}
}

func RotateOnEvent(eventChan <-chan struct{}) RotationStrategy {
	return &EventRotationStrategy{eventChan: eventChan}
}

func (e *EventRotationStrategy) ShouldRotate(ctx context.Context, _ time.Time) (bool, error) {
	select {
	case <-e.eventChan:
		return true, nil
	case <-ctx.Done():
		return false, ctx.Err()
	default:
		return false, nil
	}
}

func RotateOnMinTTL(minTTL time.Duration) RotationStrategy {
	return RotateOnInterval(minTTL)
}
