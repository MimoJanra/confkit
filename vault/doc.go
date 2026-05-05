// Package vault provides a confkit Source for HashiCorp Vault KV secrets.
//
// Import and use with LoadWithOptions:
//
//	import "github.com/MimoJanra/confkit/vault"
//
//	// Using token authentication
//	cfg, err := confkit.Load[Config](
//	    vault.FromVault("https://vault.example.com:8200", auth, "/secret/app"),
//	)
//
// # Authentication
//
// Vault supports multiple authentication methods:
//
//	// Token auth (development/testing)
//	auth := vault.VaultTokenAuth("s.xxxxxxxxxxxx")
//
//	// AppRole auth (production)
//	auth := vault.VaultAppRoleAuth("role_id", "secret_id")
//
//	// Kubernetes auth (if running on K8s)
//	auth := vault.VaultKubernetesAuth("role_name", jwt_token)
//
// # Usage
//
// Define config with Vault secret fields:
//
//	type Config struct {
//	    APIKey   string `validate:"required" secret:"true"`
//	    Password string `validate:"required" secret:"true"`
//	}
//
//	auth := vault.VaultTokenAuth("s.xxxx")
//	cfg, err := confkit.Load[Config](
//	    vault.FromVault("https://vault.example.com:8200", auth, "/secret/myapp"),
//	)
//
// Field names are mapped to Vault KV paths using dot notation:
// Config.APIKey → api_key (in the secret)
//
// # KV Versions
//
// Vault supports KV v1 and v2. Use the appropriate constructor:
//
//	// KV v2 (default)
//	vault.FromVault(addr, auth, pathPrefix)
//
//	// KV v1
//	vault.FromVaultWithKVVersion(addr, auth, 1, pathPrefix)
//
// # Security
//
// Always use secret:"true" tags for credentials loaded from Vault:
//
//	type Config struct {
//	    VaultToken string `secret:"true"` // Redacted in errors
//	    APIKey     string `secret:"true"` // Redacted in errors
//	}
//
// Secrets are automatically redacted in error messages and logs.
package vault
