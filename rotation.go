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

func (r *RotationEngine) Start(ctx context.Context, interval time.Duration) {
	r.ticker = time.NewTicker(interval)
	go r.run(ctx)
}

func (r *RotationEngine) Stop() {
	if r.ticker != nil {
		r.ticker.Stop()
	}
	close(r.stopChan)
}

func (r *RotationEngine) run(ctx context.Context) {
	for {
		select {
		case <-r.stopChan:
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
	callbacks := r.callbacks
	r.callbackMutex.RUnlock()

	for _, cb := range callbacks {
		go cb(oldCfg, newCfg, err)
	}
}

func (r *RotationEngine) IsRotating() bool {
	return r.isRotating.Load()
}

type IntervalRotationStrategy struct {
	interval time.Duration
}

func RotateOnInterval(interval time.Duration) RotationStrategy {
	return &IntervalRotationStrategy{interval: interval}
}

func (i *IntervalRotationStrategy) ShouldRotate(ctx context.Context, lastRotation time.Time) (bool, error) {
	return time.Since(lastRotation) >= i.interval, nil
}

type EventRotationStrategy struct {
	eventChan <-chan struct{}
}

func RotateOnEvent(eventChan <-chan struct{}) RotationStrategy {
	return &EventRotationStrategy{eventChan: eventChan}
}

func (e *EventRotationStrategy) ShouldRotate(ctx context.Context, lastRotation time.Time) (bool, error) {
	select {
	case <-e.eventChan:
		return true, nil
	case <-ctx.Done():
		return false, ctx.Err()
	default:
		return false, nil
	}
}

type TTLRotationStrategy struct {
	minTTL time.Duration
}

func RotateOnMinTTL(minTTL time.Duration) RotationStrategy {
	return &TTLRotationStrategy{minTTL: minTTL}
}

func (t *TTLRotationStrategy) ShouldRotate(ctx context.Context, lastRotation time.Time) (bool, error) {
	return time.Since(lastRotation) >= t.minTTL, nil
}
