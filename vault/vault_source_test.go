package vault

import (
	"context"
	"testing"

	"github.com/hashicorp/vault/api"
)

func TestVaultTokenAuth(t *testing.T) {
	auth := VaultTokenAuth("test-token")
	if auth == nil {
		t.Fatal("Expected non-nil auth")
	}

	token, err := auth.Authenticate(context.Background(), &api.Client{})
	if err != nil {
		t.Fatalf("Authenticate failed: %v", err)
	}

	if token != "test-token" {
		t.Errorf("Expected test-token, got %s", token)
	}
}

func TestVaultAppRoleAuth(t *testing.T) {
	auth := VaultAppRoleAuth("role-123", "secret-456")
	if auth == nil {
		t.Fatal("Expected non-nil auth")
	}

	_, ok := auth.(*vaultAppRoleAuth)
	if !ok {
		t.Error("Expected *vaultAppRoleAuth type")
	}
}

func TestVaultKubernetesAuth(t *testing.T) {
	auth := VaultKubernetesAuth("my-role", "jwt-token")
	if auth == nil {
		t.Fatal("Expected non-nil auth")
	}

	_, ok := auth.(*vaultKubernetesAuth)
	if !ok {
		t.Error("Expected *vaultKubernetesAuth type")
	}
}

func TestVaultSourceName(t *testing.T) {
	src := &VaultSource{}
	name := src.Name()
	if name != "vault" {
		t.Errorf("Expected name 'vault', got %q", name)
	}
}

func TestFromVault(t *testing.T) {
	auth := VaultTokenAuth("test-token")
	src := FromVault("http://localhost:8200", auth, "myapp")

	if src == nil {
		t.Fatal("Expected non-nil source")
	}

	if src.Name() != "vault" && src.Name() != "error" {
		t.Errorf("Expected vault or error source, got %q", src.Name())
	}
}

func TestFromVaultWithKVVersion(t *testing.T) {
	tests := []struct {
		kvVersion int
		expected  int
	}{
		{1, 1},
		{2, 2},
		{3, 2},
		{0, 2},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.kvVersion)), func(t *testing.T) {
			auth := VaultTokenAuth("test-token")
			src := FromVaultWithKVVersion("http://localhost:8200", auth, tt.kvVersion, "myapp")

			if src == nil {
				t.Fatal("Expected non-nil source")
			}

			if vsrc, ok := src.(*VaultSource); ok {
				if vsrc.kvVersion != tt.expected {
					t.Errorf("Expected kvVersion=%d, got %d", tt.expected, vsrc.kvVersion)
				}
			}
		})
	}
}

func TestVaultSourceBuildPath(t *testing.T) {
	tests := []struct {
		kvVersion int
		fieldPath string
		expected  string
	}{
		{1, "port", "secret/myapp/port"},
		{2, "port", "secret/data/myapp/port"},
		{1, "database/host", "secret/myapp/database/host"},
		{2, "database/host", "secret/data/myapp/database/host"},
	}

	for _, tt := range tests {
		t.Run(tt.fieldPath, func(t *testing.T) {
			src := &VaultSource{kvVersion: tt.kvVersion}
			result := src.buildPath(tt.fieldPath)
			if result != tt.expected {
				t.Errorf("buildPath(%q) = %q, want %q", tt.fieldPath, result, tt.expected)
			}
		})
	}
}
