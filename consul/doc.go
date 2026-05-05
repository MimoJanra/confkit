// Package consul provides a confkit Source for HashiCorp Consul KV storage.
//
// Import and use with Load:
//
//	import "github.com/MimoJanra/confkit/consul"
//
//	cfg, err := confkit.Load[Config](
//	    consul.FromConsul("consul.example.com:8500"),
//	)
//
// # Configuration
//
// Create a Consul source with optional authentication:
//
//	// Basic usage
//	src := consul.FromConsul("consul.example.com:8500")
//
//	// With token
//	src := consul.FromConsulWithToken(
//	    "consul.example.com:8500",
//	    "your-token",
//	)
//
//	// With datacenter
//	src := consul.FromConsulWithOptions(
//	    "consul.example.com:8500",
//	    "token",
//	    "dc1",
//	)
//
// # Usage
//
// Config fields are mapped to Consul KV keys using dot notation:
//
//	type Config struct {
//	    Host     string `validate:"required"`
//	    Port     int    `validate:"min=1,max=65535"`
//	    APIKey   string `secret:"true" validate:"required"`
//	}
//
//	cfg, err := confkit.Load[Config](
//	    consul.FromConsul("consul.example.com:8500"),
//	)
//
// KV path mapping:
// Config.Host → host
// Config.Port → port
// Config.APIKey → api_key (stored as secret)
//
// # Secrets
//
// Always mark sensitive fields with secret:"true":
//
//	type Config struct {
//	    APIKey   string `secret:"true"`
//	    Password string `secret:"true"`
//	}
//
// Secrets are automatically redacted in error messages and logs.
//
// # Datacenter
//
// Consul supports multi-datacenter deployments. Specify the datacenter:
//
//	src := consul.FromConsulWithOptions(addr, token, "dc2")
//
// The source will query the specified datacenter's KV store.
package consul
