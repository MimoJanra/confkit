package confkit_test

import (
	"context"
	"testing"

	confkit "github.com/MimoJanra/confkit"
)

type memorySource struct {
	data map[string]string
}

func (m *memorySource) Name() string { return "memory" }

func (m *memorySource) Lookup(_ context.Context, field *confkit.FieldInfo) (any, bool, error) {
	v, ok := m.data[field.Name]
	return v, ok, nil
}

func TestRegisterSource(t *testing.T) {
	defer confkit.UnregisterSource("test_memory")

	err := confkit.RegisterSource("test_memory", func() confkit.Source {
		return &memorySource{data: make(map[string]string)}
	})
	if err != nil {
		t.Fatalf("Failed to register source: %v", err)
	}

	src, err := confkit.NewSource("test_memory")
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
	defer confkit.UnregisterSource("dup_test")

	_ = confkit.RegisterSource("dup_test", func() confkit.Source {
		return &memorySource{data: make(map[string]string)}
	})
	err := confkit.RegisterSource("dup_test", func() confkit.Source {
		return &memorySource{data: make(map[string]string)}
	})
	if err == nil {
		t.Fatal("Expected error for duplicate registration")
	}
}

func TestNewSourceNotFound(t *testing.T) {
	_, err := confkit.NewSource("nonexistent_source")
	if err == nil {
		t.Fatal("Expected error for nonexistent source")
	}
}

func TestLoadWithCustomSource(t *testing.T) {
	defer confkit.UnregisterSource("test_custom")

	_ = confkit.RegisterSource("test_custom", func() confkit.Source {
		return &memorySource{data: map[string]string{
			"Port": "9000",
			"Host": "custom.local",
		}}
	})

	type Config struct {
		Port int    `custom:"Port"`
		Host string `custom:"Host"`
	}
	src, err := confkit.NewSource("test_custom")
	if err != nil {
		t.Fatalf("Failed to create source: %v", err)
	}
	cfg, err := confkit.Load[Config](src)
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
	err := confkit.RegisterSource("", func() confkit.Source {
		return &memorySource{data: make(map[string]string)}
	})
	if err == nil {
		t.Fatal("Expected error for empty source name")
	}
}

func TestRegisterSourceNilFactory(t *testing.T) {
	err := confkit.RegisterSource("nil_test", nil)
	if err == nil {
		t.Fatal("Expected error for nil factory")
	}
}
