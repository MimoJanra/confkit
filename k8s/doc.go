// Package k8s provides confkit Source for Kubernetes ConfigMaps.
//
// Import and use with Load:
//
//	import "github.com/MimoJanra/confkit/k8s"
//
//	cfg, err := confkit.Load[Config](
//	    k8s.FromKubernetesConfigMap("default", "app-config"),
//	)
//
// # Kubernetes ConfigMap
//
// Load configuration from Kubernetes ConfigMaps mounted as volumes:
//
//	// Default mount path: /var/run/secrets/config/{namespace}/{configmap}
//	src := k8s.FromKubernetesConfigMap("default", "app-config")
//
//	// Custom mount path
//	src := k8s.FromKubernetesConfigMapWithPath("default", "app-config", "/etc/config")
//
// ConfigMap keys are mapped to struct fields:
//
//	type Config struct {
//	    Port     int    `env:"PORT"`
//	    LogLevel string `env:"LOG_LEVEL"`
//	}
//
// With namespace "default" and configmap "app-config", the source looks for files:
// • /var/run/secrets/config/default/app-config/PORT
// • /var/run/secrets/config/default/app-config/LOG_LEVEL
//
// Each file contains the corresponding field value.
//
// # ConfigMap Creation
//
// Create a ConfigMap in Kubernetes:
//
//	kubectl create configmap app-config \
//	  --from-literal=PORT=8080 \
//	  --from-literal=LOG_LEVEL=info
//
// Mount it as a volume in your Pod:
//
//	spec:
//	  containers:
//	  - name: app
//	    volumeMounts:
//	    - name: config
//	      mountPath: /var/run/secrets/config
//	  volumes:
//	  - name: config
//	    configMap:
//	      name: app-config
//
// # Field Resolution
//
// The source resolves fields using multiple strategies:
//
// 1. Field name (exact)
// 2. env tag (e.g., env:"PORT")
// 3. yaml tag (e.g., yaml:"port")
// 4. json tag (e.g., json:"port")
// 5. Snake case (e.g., LogLevel → LOG_LEVEL)
//
// The first matching file is used.
//
// # Secrets
//
// For sensitive configuration, use Kubernetes Secrets with the same interface:
//
//	// Secrets work identically to ConfigMaps
//	spec:
//	  containers:
//	  - name: app
//	    volumeMounts:
//	    - name: secrets
//	      mountPath: /var/run/secrets/config
//	  volumes:
//	  - name: secrets
//	    secret:
//	      secretName: app-secrets
//
// Always mark sensitive fields with secret:"true":
//
//	type Config struct {
//	    APIKey   string `secret:"true"`
//	    Password string `secret:"true"`
//	}
//
// Secrets are automatically redacted in error messages.
//
// # Security
//
// ConfigMaps are not encrypted by default. For sensitive data:
// • Use Kubernetes Secrets instead
// • Enable encryption at rest in etcd
// • Use proper RBAC to restrict access
// • Mark sensitive fields with secret:"true"
package k8s
