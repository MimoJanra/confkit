package confkit

import (
	"testing"
)

type MemorySource struct {
	data map[string]string
}

func (m *MemorySource) Name() string {
	return "memory"
}

func (m *MemorySource) Lookup(field *FieldInfo) (any, bool, error) {
	v, ok := m.data[field.Name]
	return v, ok, nil
}

func TestRegisterSource(t *testing.T) {
	defer UnregisterSource("test_memory")

	err := RegisterSource("test_memory", func() Source {
		return &MemorySource{data: make(map[string]string)}
	})
	if err != nil {
		t.Fatalf("Failed to register source: %v", err)
	}

	src, err := NewSource("test_memory")
	if err != nil {
		t.Fatalf("Failed to get source: %v", err)
	}
	if src == nil {
		t.Fatal("Source is nil")
	}
	if src.Name() != "memory" {
		t.Errorf("Expected source name 'memory', got '%s'", src.Name())
	}
}

func TestRegisterSourceDuplicate(t *testing.T) {
	defer UnregisterSource("dup_test")

	RegisterSource("dup_test", func() Source {
		return &MemorySource{data: make(map[string]string)}
	})

	err := RegisterSource("dup_test", func() Source {
		return &MemorySource{data: make(map[string]string)}
	})
	if err == nil {
		t.Fatal("Expected error for duplicate registration")
	}
}

func TestNewSourceNotFound(t *testing.T) {
	_, err := NewSource("nonexistent_source")
	if err == nil {
		t.Fatal("Expected error for nonexistent source")
	}
}

func TestLoadWithCustomSource(t *testing.T) {
	defer UnregisterSource("test_custom")

	RegisterSource("test_custom", func() Source {
		return &MemorySource{
			data: map[string]string{
				"Port": "9000",
				"Host": "custom.local",
			},
		}
	})

	type Config struct {
		Port int    `custom:"Port"`
		Host string `custom:"Host"`
	}

	src, err := NewSource("test_custom")
	if err != nil {
		t.Fatalf("Failed to create source: %v", err)
	}

	cfg, err := Load[Config](src)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Port != 9000 {
		t.Errorf("Expected Port=9000, got %d", cfg.Port)
	}
	if cfg.Host != "custom.local" {
		t.Errorf("Expected Host='custom.local', got '%s'", cfg.Host)
	}
}

func TestRegisterSourceInvalidName(t *testing.T) {
	err := RegisterSource("", func() Source {
		return &MemorySource{data: make(map[string]string)}
	})
	if err == nil {
		t.Fatal("Expected error for empty source name")
	}
}

func TestRegisterSourceNilFactory(t *testing.T) {
	err := RegisterSource("nil_test", nil)
	if err == nil {
		t.Fatal("Expected error for nil factory")
	}
}
