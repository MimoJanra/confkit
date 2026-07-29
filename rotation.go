package confkit

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// RotationStrategy decides when secrets should be reloaded. ShouldRotate is polled
// on the engine's interval and receives the time of the previous rotation.
type RotationStrategy interface {
	ShouldRotate(ctx context.Context, lastRotation time.Time) (bool, error)
}

// RotationCallback is notified when a rotation happens or when the strategy fails.
// Callbacks run concurrently in their own goroutines, so they must be safe for
// concurrent use.
type RotationCallback func(oldCfg, newCfg any, err error)

// RotationEngine polls a RotationStrategy and notifies callbacks when secrets need
// reloading.
//
// Create one with NewRotationEngine, register callbacks with AddCallback, then call
// Start. The engine stops when Stop is called or when the context passed to Start
// is cancelled.
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

// NewRotationEngine returns an engine driven by strategy, with the last rotation
// time set to now. The engine does nothing until Start is called.
func NewRotationEngine(strategy RotationStrategy) *RotationEngine {
	return &RotationEngine{
		strategy:     strategy,
		callbacks:    make([]RotationCallback, 0),
		lastRotation: time.Now(),
		stopChan:     make(chan struct{}),
	}
}

// AddCallback registers cb. It is safe to call while the engine is running.
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

// Stop halts the engine and its ticker. It is safe to call more than once, and
// safe on an engine that was never started. Stop does not wait for callbacks that
// are already running.
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

// IsRotating reports whether a rotation is in progress.
func (r *RotationEngine) IsRotating() bool {
	return r.isRotating.Load()
}

// LastRotation returns when the most recent rotation occurred, or the engine's
// creation time if none has happened yet.
func (r *RotationEngine) LastRotation() time.Time {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.lastRotation
}

// IntervalRotationStrategy rotates once a fixed period has elapsed since the last
// rotation. Construct it with RotateOnInterval.
type IntervalRotationStrategy struct {
	interval time.Duration
}

// RotateOnInterval returns a strategy that rotates every interval.
func RotateOnInterval(interval time.Duration) RotationStrategy {
	return &IntervalRotationStrategy{interval: interval}
}

// ShouldRotate reports whether interval has elapsed since lastRotation.
func (i *IntervalRotationStrategy) ShouldRotate(_ context.Context, lastRotation time.Time) (bool, error) {
	return time.Since(lastRotation) >= i.interval, nil
}

// EventRotationStrategy rotates whenever a value is received on a channel, which
// lets an external trigger such as a webhook or SIGHUP drive rotation. Construct it
// with RotateOnEvent.
type EventRotationStrategy struct {
	eventChan <-chan struct{}
}

// RotateOnEvent returns a strategy that rotates once per value received on
// eventChan.
func RotateOnEvent(eventChan <-chan struct{}) RotationStrategy {
	return &EventRotationStrategy{eventChan: eventChan}
}

// ShouldRotate reports whether an event is pending. It does not block: with no event
// waiting it returns false, and it reports ctx.Err() if the context is done.
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

// RotateOnMinTTL returns a strategy that rotates once minTTL has elapsed, for
// secrets that carry a time-to-live and must be renewed before it expires. It is
// currently equivalent to RotateOnInterval.
func RotateOnMinTTL(minTTL time.Duration) RotationStrategy {
	return RotateOnInterval(minTTL)
}
