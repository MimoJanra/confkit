package confkit

import (
	"fmt"
	"sync"
)

type SourceFactory func() Source

var (
	sourceRegistry = make(map[string]SourceFactory)
	registryMutex  sync.RWMutex
)

func RegisterSource(name string, factory SourceFactory) error {
	if name == "" {
		return fmt.Errorf("source name cannot be empty")
	}
	if factory == nil {
		return fmt.Errorf("source factory cannot be nil")
	}

	registryMutex.Lock()
	defer registryMutex.Unlock()

	if _, exists := sourceRegistry[name]; exists {
		return fmt.Errorf("source %q already registered", name)
	}

	sourceRegistry[name] = factory
	return nil
}

func NewSource(name string) (Source, error) {
	registryMutex.RLock()
	factory, exists := sourceRegistry[name]
	registryMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("source %q not registered", name)
	}

	return factory(), nil
}

func UnregisterSource(name string) {
	registryMutex.Lock()
	defer registryMutex.Unlock()
	delete(sourceRegistry, name)
}
