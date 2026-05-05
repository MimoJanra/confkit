package consul

import (
	"testing"
	"time"
)

func TestFromConsul(t *testing.T) {
	src := FromConsul("localhost:8500")
	if src == nil {
		t.Fatal("Expected non-nil source")
	}

	if src.Name() != "consul" && src.Name() != "error" {
		t.Errorf("Expected consul or error source, got %q", src.Name())
	}
}

func TestFromConsulWithToken(t *testing.T) {
	src := FromConsulWithToken("localhost:8500", "my-token")
	if src == nil {
		t.Fatal("Expected non-nil source")
	}

	if src.Name() != "consul" && src.Name() != "error" {
		t.Errorf("Expected consul or error source, got %q", src.Name())
	}
}

func TestFromConsulWithOptions(t *testing.T) {
	src := FromConsulWithOptions("localhost:8500", "my-token", "dc1")
	if src == nil {
		t.Fatal("Expected non-nil source")
	}

	if src.Name() != "consul" && src.Name() != "error" {
		t.Errorf("Expected consul or error source, got %q", src.Name())
	}
}

func TestConsulSourceName(t *testing.T) {
	src := &ConsulSource{}
	name := src.Name()
	if name != "consul" {
		t.Errorf("Expected name 'consul', got %q", name)
	}
}

func TestConsulSourceBuildKey(t *testing.T) {
	src := &ConsulSource{prefix: "myapp"}

	tests := []struct {
		fieldPath string
		expected  string
	}{
		{"port", "myapp/port"},
		{"Database.Host", "myapp/database.host"},
		{"Database.Port", "myapp/database.port"},
	}

	for _, tt := range tests {
		t.Run(tt.fieldPath, func(t *testing.T) {
			result := src.buildKey(tt.fieldPath)
			if result != tt.expected {
				t.Errorf("buildKey(%q) = %q, want %q", tt.fieldPath, result, tt.expected)
			}
		})
	}
}

func TestConsulSourceSetPrefix(t *testing.T) {
	src := &ConsulSource{prefix: "default"}
	src.SetPrefix("custom")

	if src.prefix != "custom" {
		t.Errorf("Expected prefix=custom, got %s", src.prefix)
	}
}

func TestConsulSourceSetWaitDuration(t *testing.T) {
	src := &ConsulSource{}
	duration := 10 * time.Second
	src.SetWaitDuration(duration)

	if src.waitDuration != duration {
		t.Errorf("Expected waitDuration=%v, got %v", duration, src.waitDuration)
	}
}

func TestConsulSourceDefaultValues(t *testing.T) {
	src := &ConsulSource{}

	if src.prefix != "" {
		t.Errorf("Expected default prefix to be empty, got %q", src.prefix)
	}

	if src.waitDuration != 0 {
		t.Errorf("Expected default waitDuration to be 0, got %v", src.waitDuration)
	}
}
