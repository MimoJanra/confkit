package confkit

import (
	"testing"
	"time"
)

func TestFromAWSSecretsManager(t *testing.T) {
	src := FromAWSSecretsManager("myapp/database")
	if src == nil {
		t.Fatal("Expected non-nil source")
	}

	if src.Name() != "aws-secrets-manager" && src.Name() != "file" {
		t.Errorf("Expected aws-secrets-manager or file, got %q", src.Name())
	}
}

func TestFromAWSSecretsManagerWithRegion(t *testing.T) {
	src := FromAWSSecretsManagerWithRegion("myapp/database", "us-east-1")
	if src == nil {
		t.Fatal("Expected non-nil source")
	}

	if src.Name() != "aws-secrets-manager" && src.Name() != "file" {
		t.Errorf("Expected aws-secrets-manager or file, got %q", src.Name())
	}
}

func TestFromAWSSecretsManagerWithOptions(t *testing.T) {
	src := FromAWSSecretsManagerWithOptions("myapp/database", "us-west-2", 10*time.Minute)
	if src == nil {
		t.Fatal("Expected non-nil source")
	}

	if src.Name() != "aws-secrets-manager" && src.Name() != "file" {
		t.Errorf("Expected aws-secrets-manager or file, got %q", src.Name())
	}
}

func TestAWSSecretsManagerSourceName(t *testing.T) {
	src := &AWSSecretsManagerSource{}
	name := src.Name()
	if name != "aws-secrets-manager" {
		t.Errorf("Expected name 'aws-secrets-manager', got %q", name)
	}
}

func TestAWSSecretsManagerSourceCacheTTL(t *testing.T) {
	src := &AWSSecretsManagerSource{
		cacheTTL:    100 * time.Millisecond,
		lastCacheAt: time.Now(),
	}

	if time.Since(src.lastCacheAt) >= src.cacheTTL {
		t.Error("Cache should not be expired immediately")
	}

	src.lastCacheAt = time.Now().Add(-200 * time.Millisecond)
	if time.Since(src.lastCacheAt) < src.cacheTTL {
		t.Error("Cache should be expired after TTL")
	}
}

func TestAWSSecretsManagerSourceDefaultTTL(t *testing.T) {
	src, err := NewAWSSecretsManagerSource("myapp/db", "", 5*time.Minute)
	if err == nil {
		if src.cacheTTL != 5*time.Minute {
			t.Errorf("Expected cacheTTL=5m, got %v", src.cacheTTL)
		}
	}
}

func TestAWSSecretsManagerSourceRegionConfiguration(t *testing.T) {
	tests := []struct {
		region   string
		expected string
	}{
		{"us-east-1", "us-east-1"},
		{"eu-west-1", "eu-west-1"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.region, func(t *testing.T) {
			src, err := NewAWSSecretsManagerSource("test", tt.region, 5*time.Minute)
			if err == nil {
				if src.region != tt.expected {
					t.Errorf("Expected region=%q, got %q", tt.expected, src.region)
				}
			}
		})
	}
}
