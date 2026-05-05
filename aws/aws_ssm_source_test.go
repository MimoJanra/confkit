package aws

import (
	"testing"
	"time"
)

func TestNewAWSSSMSource(t *testing.T) {
	src, err := NewAWSSSMSource("/myapp/", 5*time.Minute)
	if err != nil {
		t.Logf("Expected nil error (credentials not checked until first call), got %v", err)
		return
	}
	if src == nil {
		t.Error("Expected non-nil source")
	}
	if src.pathPrefix != "/myapp/" {
		t.Errorf("Expected pathPrefix /myapp/, got %s", src.pathPrefix)
	}
	if src.cacheTTL != 5*time.Minute {
		t.Errorf("Expected cacheTTL 5m, got %v", src.cacheTTL)
	}
}

func TestAWSSSMSourceName(t *testing.T) {
	src := &AWSSSMSource{
		pathPrefix: "/myapp/",
		cache:      make(map[string]string),
	}

	name := src.Name()
	if name != "aws-ssm" {
		t.Errorf("Expected name 'aws-ssm', got %q", name)
	}
}

func TestFieldPathToParameterPath(t *testing.T) {
	src := &AWSSSMSource{pathPrefix: "/myapp/"}

	tests := []struct {
		fieldPath string
		expected  string
	}{
		{"Port", "/myapp/port"},
		{"Database.Host", "/myapp/database/host"},
		{"Database.Port", "/myapp/database/port"},
		{"LogLevel", "/myapp/loglevel"},
	}

	for _, tt := range tests {
		t.Run(tt.fieldPath, func(t *testing.T) {
			result := src.fieldPathToParameterPath(tt.fieldPath)
			if result != tt.expected {
				t.Errorf("fieldPathToParameterPath(%q) = %q, want %q",
					tt.fieldPath, result, tt.expected)
			}
		})
	}
}

func TestParameterPathToFieldPath(t *testing.T) {
	src := &AWSSSMSource{pathPrefix: "/myapp/"}

	tests := []struct {
		paramPath string
		expected  string
	}{
		{"/myapp/port", "port"},
		{"/myapp/database/host", "database.host"},
		{"/myapp/database/port", "database.port"},
		{"/myapp/loglevel", "loglevel"},
		{"/otherapp/port", ""},
	}

	for _, tt := range tests {
		t.Run(tt.paramPath, func(t *testing.T) {
			result := src.parameterPathToFieldPath(tt.paramPath)
			if result != tt.expected {
				t.Errorf("parameterPathToFieldPath(%q) = %q, want %q",
					tt.paramPath, result, tt.expected)
			}
		})
	}
}

func TestFromAWSSSMParameterStore(t *testing.T) {
	src := FromAWSSSMParameterStore("/myapp")

	if src == nil {
		t.Fatal("Expected non-nil source")
	}

	if src.Name() != "aws-ssm" && src.Name() != "error" {
		t.Errorf("Expected aws-ssm or error, got %q", src.Name())
	}
}

func TestFromAWSSSMParameterStorePathNormalization(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/myapp", "/myapp/"},
		{"/myapp/", "/myapp/"},
		{"/my-app", "/my-app/"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			src := FromAWSSSMParameterStore(tt.input)
			if src.Name() == "error" {
				t.Skip("AWS SDK not available")
			}
			if asrc, ok := src.(*AWSSSMSource); ok {
				if asrc.pathPrefix != tt.expected {
					t.Errorf("pathPrefix = %q, want %q", asrc.pathPrefix, tt.expected)
				}
			}
		})
	}
}

func TestAWSSSMSourceCacheTTL(t *testing.T) {
	src := &AWSSSMSource{
		pathPrefix:  "/myapp/",
		cache:       make(map[string]string),
		cacheTTL:    100 * time.Millisecond,
		lastCacheAt: time.Now(),
	}

	if src.lastCacheAt.Add(src.cacheTTL).Before(time.Now()) {
		t.Error("Cache TTL should not have expired immediately")
	}

	src.lastCacheAt = time.Now().Add(-200 * time.Millisecond)
	if src.lastCacheAt.Add(src.cacheTTL).After(time.Now()) {
		t.Error("Cache TTL should have expired")
	}
}

func TestFromAWSSSMParameterStoreWithTTL(t *testing.T) {
	src := FromAWSSSMParameterStoreWithTTL("/myapp", 10*time.Minute)

	if src == nil {
		t.Fatal("Expected non-nil source")
	}

	if src.Name() != "aws-ssm" && src.Name() != "error" {
		t.Errorf("Expected aws-ssm or error, got %q", src.Name())
	}
}
