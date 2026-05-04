package confkit

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type ConfigWatcher struct {
	filePath      string
	lastModTime   time.Time
	listeners     []ConfigChangeListener
	listenerMutex sync.RWMutex
	stopChan      chan struct{}
	done          chan struct{}
	pollInterval  time.Duration
}

type ConfigChangeListener func(oldCfg, newCfg any, err error)

func NewConfigWatcher(filePath string) (*ConfigWatcher, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot watch file: %w", err)
	}

	return &ConfigWatcher{
		filePath:     filePath,
		lastModTime:  fileInfo.ModTime(),
		listeners:    make([]ConfigChangeListener, 0),
		stopChan:     make(chan struct{}),
		done:         make(chan struct{}),
		pollInterval: 500 * time.Millisecond,
	}, nil
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
	close(cw.stopChan)
	<-cw.done
}

func (cw *ConfigWatcher) watch() {
	defer close(cw.done)

	ticker := time.NewTicker(cw.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-cw.stopChan:
			return
		case <-ticker.C:
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
	listeners := cw.listeners
	cw.listenerMutex.RUnlock()

	for _, listener := range listeners {
		listener(oldCfg, newCfg, err)
	}
}

func (cw *ConfigWatcher) SetPollInterval(interval time.Duration) {
	if interval > 0 {
		cw.pollInterval = interval
	}
}
