package confkit

import (
	"testing"
)

func TestFromEtcd(t *testing.T) {
	endpoints := []string{"localhost:2379"}
	src := FromEtcd(endpoints)
	if src == nil {
		t.Fatal("Expected non-nil source")
	}

	if src.Name() != "etcd" && src.Name() != "file" {
		t.Errorf("Expected etcd or file source, got %q", src.Name())
	}
}

func TestFromEtcdWithPrefix(t *testing.T) {
	endpoints := []string{"localhost:2379"}
	src := FromEtcdWithPrefix(endpoints, "/myapp/")
	if src == nil {
		t.Fatal("Expected non-nil source")
	}

	if src.Name() != "etcd" && src.Name() != "file" {
		t.Errorf("Expected etcd or file source, got %q", src.Name())
	}
}

func TestEtcdSourceName(t *testing.T) {
	src := &EtcdSource{}
	name := src.Name()
	if name != "etcd" {
		t.Errorf("Expected name 'etcd', got %q", name)
	}
}

func TestEtcdSourceBuildKey(t *testing.T) {
	tests := []struct {
		prefix    string
		fieldPath string
		expected  string
	}{
		{"/myapp/", "port", "/myapp/port"},
		{"/myapp/", "Database.Host", "/myapp/database.host"},
		{"/config/", "port", "/config/port"},
	}

	for _, tt := range tests {
		t.Run(tt.fieldPath, func(t *testing.T) {
			src := &EtcdSource{prefix: tt.prefix}
			result := src.buildKey(tt.fieldPath)
			if result != tt.expected {
				t.Errorf("buildKey(%q) = %q, want %q", tt.fieldPath, result, tt.expected)
			}
		})
	}
}

func TestEtcdSourceSetTimeout(t *testing.T) {
	src := &EtcdSource{timeout: 5}
	src.SetTimeout(10)

	if src.timeout != 10 {
		t.Errorf("Expected timeout=10, got %d", src.timeout)
	}
}

func TestFromEtcdWithPrefixNormalization(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/myapp", "/myapp/"},
		{"/myapp/", "/myapp/"},
		{"/config", "/config/"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			endpoints := []string{"localhost:2379"}
			src := FromEtcdWithPrefix(endpoints, tt.input)

			if esrc, ok := src.(*EtcdSource); ok {
				if esrc.prefix != tt.expected {
					t.Errorf("prefix = %q, want %q", esrc.prefix, tt.expected)
				}
			}
		})
	}
}
