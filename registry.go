package confkit

import (
	"fmt"
	"sync"
)

// SourceFactory constructs a Source on demand. It is called afresh by every
// NewSource, so each caller receives an independent instance.
type SourceFactory func() Source

var (
	sourceRegistry = make(map[string]SourceFactory)
	registryMutex  sync.RWMutex
)

// RegisterSource registers factory under name in the process-wide registry, so a
// source can be selected by a string from configuration or a CLI flag.
//
// It returns an error if name is empty, factory is nil, or name is already taken;
// re-registering requires UnregisterSource first. Safe for concurrent use.
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

// NewSource builds the Source registered under name, or reports an error if no such
// name is registered. Safe for concurrent use.
func NewSource(name string) (Source, error) {
	registryMutex.RLock()
	factory, exists := sourceRegistry[name]
	registryMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("source %q not registered", name)
	}

	return factory(), nil
}

// UnregisterSource removes name from the registry. Unregistering a name that was
// never registered is not an error. Safe for concurrent use.
func UnregisterSource(name string) {
	registryMutex.Lock()
	defer registryMutex.Unlock()
	delete(sourceRegistry, name)
}
