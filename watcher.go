package confkit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

type ConfigDelta struct {
	Added   []string
	Removed []string
	Changed []string
}

type ConfigChangeListener func(oldCfg, newCfg any, err error)

type ConfigChangeListenerWithDelta func(delta ConfigDelta, oldSnap, newSnap map[string]any, err error)

type ConfigWatcher struct {
	filePath           string
	lastModTime        time.Time
	listeners          []ConfigChangeListener
	listenerMutex      sync.RWMutex
	deltaListeners     []ConfigChangeListenerWithDelta
	deltaListenerMutex sync.RWMutex
	lastSnapshot       map[string]any
	snapshotMu         sync.Mutex
	stopChan           chan struct{}
	done               chan struct{}
	pollInterval       atomic.Value
	mu                 sync.RWMutex
	ticker             *time.Ticker
	once               sync.Once
}

func NewConfigWatcher(filePath string) (*ConfigWatcher, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot watch file: %w", err)
	}

	snap, _ := parseFileToFlatMap(filePath)
	if snap == nil {
		snap = make(map[string]any)
	}

	watcher := &ConfigWatcher{
		filePath:       filePath,
		lastModTime:    fileInfo.ModTime(),
		listeners:      make([]ConfigChangeListener, 0),
		deltaListeners: make([]ConfigChangeListenerWithDelta, 0),
		lastSnapshot:   snap,
		stopChan:       make(chan struct{}),
		done:           make(chan struct{}),
	}
	watcher.pollInterval.Store(500 * time.Millisecond)
	return watcher, nil
}

func (cw *ConfigWatcher) AddListener(listener ConfigChangeListener) {
	cw.listenerMutex.Lock()
	defer cw.listenerMutex.Unlock()
	cw.listeners = append(cw.listeners, listener)
}

func (cw *ConfigWatcher) AddDeltaListener(listener ConfigChangeListenerWithDelta) {
	cw.deltaListenerMutex.Lock()
	defer cw.deltaListenerMutex.Unlock()
	cw.deltaListeners = append(cw.deltaListeners, listener)
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

func (cw *ConfigWatcher) SetPollInterval(interval time.Duration) {
	if interval <= 0 {
		return
	}
	cw.pollInterval.Store(interval)
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
				statErr := fmt.Errorf("cannot stat file: %w", err)
				cw.notifyListeners(nil, nil, statErr)
				cw.notifyDeltaListeners(ConfigDelta{}, nil, nil, statErr)
				continue
			}

			if fileInfo.ModTime() != cw.lastModTime {
				cw.lastModTime = fileInfo.ModTime()
				oldSnap, newSnap, delta, parseErr := cw.computeFileChange()
				if parseErr != nil {
					cw.notifyDeltaListeners(ConfigDelta{}, oldSnap, nil, parseErr)
					cw.notifyListeners(nil, nil, parseErr)
				} else {
					cw.notifyDeltaListeners(delta, oldSnap, newSnap, nil)
					cw.notifyListeners(nil, nil, nil)
				}
			}
		}
	}
}

func (cw *ConfigWatcher) computeFileChange() (oldSnap, newSnap map[string]any, delta ConfigDelta, err error) {
	cw.snapshotMu.Lock()
	defer cw.snapshotMu.Unlock()

	oldSnap = cw.lastSnapshot
	newSnap, err = parseFileToFlatMap(cw.filePath)
	if err != nil {

		return oldSnap, nil, ConfigDelta{}, err
	}
	if newSnap == nil {
		newSnap = make(map[string]any)
	}
	cw.lastSnapshot = newSnap
	delta = computeDelta(oldSnap, newSnap)
	return oldSnap, newSnap, delta, nil
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

func (cw *ConfigWatcher) notifyDeltaListeners(delta ConfigDelta, oldSnap, newSnap map[string]any, err error) {
	cw.deltaListenerMutex.RLock()
	listeners := make([]ConfigChangeListenerWithDelta, len(cw.deltaListeners))
	copy(listeners, cw.deltaListeners)
	cw.deltaListenerMutex.RUnlock()

	for _, listener := range listeners {
		listener(delta, oldSnap, newSnap, err)
	}
}

func parseFileToFlatMap(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		err = json.Unmarshal(data, &m)
	case ".toml":
		err = toml.Unmarshal(data, &m)
	default:
		err = yaml.Unmarshal(data, &m)
	}
	if err != nil {
		return nil, err
	}
	return flattenAnyMap("", m), nil
}

func flattenAnyMap(prefix string, m map[string]any) map[string]any {
	result := make(map[string]any)
	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		if nested, ok := v.(map[string]any); ok {
			for nk, nv := range flattenAnyMap(key, nested) {
				result[nk] = nv
			}
		} else {
			result[key] = v
		}
	}
	return result
}

func computeDelta(oldSnap, newSnap map[string]any) ConfigDelta {
	var delta ConfigDelta
	for k, nv := range newSnap {
		if ov, exists := oldSnap[k]; !exists {
			delta.Added = append(delta.Added, k)
		} else if !reflect.DeepEqual(ov, nv) {
			delta.Changed = append(delta.Changed, k)
		}
	}
	for k := range oldSnap {
		if _, exists := newSnap[k]; !exists {
			delta.Removed = append(delta.Removed, k)
		}
	}
	sort.Strings(delta.Added)
	sort.Strings(delta.Changed)
	sort.Strings(delta.Removed)
	return delta
}
