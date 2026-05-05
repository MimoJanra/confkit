// Package etcd provides a confkit Source for etcd v3 configuration storage.
//
// Import and use with Load:
//
//	import "github.com/MimoJanra/confkit/etcd"
//
//	cfg, err := confkit.Load[Config](
//	    etcd.FromEtcd([]string{"etcd1.example.com:2379", "etcd2.example.com:2379"}),
//	)
//
// # Configuration
//
// Create an etcd source with optional key prefix:
//
//	// Basic usage (no prefix)
//	src := etcd.FromEtcd([]string{
//	    "etcd1.example.com:2379",
//	    "etcd2.example.com:2379",
//	})
//
//	// With key prefix
//	src := etcd.FromEtcdWithPrefix(
//	    []string{"etcd.example.com:2379"},
//	    "/myapp",
//	)
//
// # Usage
//
// Config fields are mapped to etcd keys using dot notation:
//
//	type Config struct {
//	    Host     string `validate:"required"`
//	    Port     int    `validate:"min=1,max=65535"`
//	    APIKey   string `secret:"true" validate:"required"`
//	}
//
//	cfg, err := confkit.Load[Config](
//	    etcd.FromEtcdWithPrefix(
//	        []string{"etcd.example.com:2379"},
//	        "/myapp",
//	    ),
//	)
//
// With prefix "/myapp", keys are:
// /myapp/host → Config.Host
// /myapp/port → Config.Port
// /myapp/api_key → Config.APIKey
//
// # Endpoints
//
// etcd sources require at least one endpoint. Multiple endpoints enable:
// • Load balancing
// • High availability
// • Automatic failover
//
// Endpoints should be in the format: hostname:2379 (default etcd port is 2379)
//
// # Secrets
//
// Always mark sensitive fields with secret:"true":
//
//	type Config struct {
//	    APIKey   string `secret:"true"`
//	    Token    string `secret:"true"`
//	}
//
// Secrets are automatically redacted in error messages and logs.
//
// # High Availability
//
// For production, use multiple etcd endpoints:
//
//	endpoints := []string{
//	    "etcd1.example.com:2379",
//	    "etcd2.example.com:2379",
//	    "etcd3.example.com:2379",
//	}
//	cfg, err := confkit.Load[Config](etcd.FromEtcd(endpoints))
//
// The client will automatically handle failover between endpoints.
package etcd
