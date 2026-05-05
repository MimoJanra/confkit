package confkit

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type ConfigWatcher struct {
	filePath      string
	lastModTime   time.Time
	listeners     []ConfigChangeListener
	listenerMutex sync.RWMutex
	stopChan      chan struct{}
	done          chan struct{}
	pollInterval  atomic.Value
	mu            sync.RWMutex
	ticker        *time.Ticker
	once          sync.Once
}

type ConfigChangeListener func(oldCfg, newCfg any, err error)

func NewConfigWatcher(filePath string) (*ConfigWatcher, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot watch file: %w", err)
	}

	watcher := &ConfigWatcher{
		filePath:    filePath,
		lastModTime: fileInfo.ModTime(),
		listeners:   make([]ConfigChangeListener, 0),
		stopChan:    make(chan struct{}),
		done:        make(chan struct{}),
	}
	watcher.pollInterval.Store(500 * time.Millisecond)
	return watcher, nil
}

func (cw *ConfigWatcher) AddListener(listener ConfigChangeListener) {
	cw.listenerMutex.Lock()
	defer cw.listenerMutex.Unlock()
	cw.listeners = append(cw.listeners, listener)
}

func (cw *ConfigWatcher) Start() {
	go cw.watch()
}

func (cw *ConfigWatcher) Stop() {
	cw.once.Do(func() {
		close(cw.stopChan)
		<-cw.done
	})
}

func (cw *ConfigWatcher) watch() {
	defer close(cw.done)

	cw.mu.Lock()
	interval := cw.pollInterval.Load().(time.Duration)
	ticker := time.NewTicker(interval)
	cw.ticker = ticker
	cw.mu.Unlock()
	defer ticker.Stop()

	lastInterval := interval

	for {
		select {
		case <-cw.stopChan:
			return
		case <-ticker.C:
			currentInterval := cw.pollInterval.Load().(time.Duration)
			if currentInterval != lastInterval {
				ticker.Stop()
				ticker = time.NewTicker(currentInterval)
				lastInterval = currentInterval
				cw.mu.Lock()
				cw.ticker = ticker
				cw.mu.Unlock()
				continue
			}

			fileInfo, err := os.Stat(cw.filePath)
			if err != nil {
				cw.notifyListeners(nil, nil, fmt.Errorf("cannot stat file: %w", err))
				continue
			}

			if fileInfo.ModTime() != cw.lastModTime {
				cw.lastModTime = fileInfo.ModTime()
				cw.notifyListeners(nil, nil, nil)
			}
		}
	}
}

func (cw *ConfigWatcher) notifyListeners(oldCfg, newCfg any, err error) {
	cw.listenerMutex.RLock()
	listeners := make([]ConfigChangeListener, len(cw.listeners))
	copy(listeners, cw.listeners)
	cw.listenerMutex.RUnlock()

	for _, listener := range listeners {
		listener(oldCfg, newCfg, err)
	}
}

func (cw *ConfigWatcher) SetPollInterval(interval time.Duration) {
	if interval <= 0 {
		return
	}

	cw.pollInterval.Store(interval)
}
