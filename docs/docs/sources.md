---
layout: default
title: Configuration Sources
---

# Configuration Sources

confkit supports loading configuration from multiple sources. Learn how to use each one.

## Environment Variables

**Function:** `FromEnv()`

Load configuration from environment variables. Field names are converted to uppercase.

```go
type Config struct {
    Port     int    `env:"PORT" default:"8080"`
    Database string `env:"DATABASE_URL" validate:"required"`
}

cfg, err := confkit.Load[Config](confkit.FromEnv())
```

**Environment:**
```bash
export PORT=3000
export DATABASE_URL="postgres://localhost/db"
```

### With Prefix

Use `prefix:` tag to automatically add a prefix to all env vars in a struct:

```go
type DatabaseConfig struct {
    Host string `env:"HOST"`
    Port int    `env:"PORT"`
}

type Config struct {
    DB DatabaseConfig `prefix:"DB_"`
}

// Reads from DB_HOST, DB_PORT
cfg, err := confkit.Load[Config](confkit.FromEnv())
```

---

## YAML Files

**Functions:** 
- `FromYAML(path string)` — File must exist, returns error if missing
- `FromYAMLOptional(path string)` — File is optional, returns empty source if missing

Load configuration from YAML files. Field names use snake_case by default, with **automatic snake_case ↔ CamelCase mapping**.

```go
type Config struct {
    Port     int    `yaml:"port"`
    Database string `yaml:"database"`
}

// File must exist
cfg, err := confkit.Load[Config](confkit.FromYAML("config.yaml"))

// File is optional (doesn't fail if missing)
cfg, err := confkit.Load[Config](confkit.FromYAMLOptional("config.yaml"))
```

**Automatic snake_case mapping (v0.10+):** If a field doesn't have an explicit `yaml:` tag, confkit automatically tries the snake_case version:

```go
type Config struct {
    ServerAddr    string  // Auto-matches "server_addr" in YAML
    ShutdownSecs  int     // Auto-matches "shutdown_secs" in YAML
    DatabaseHost  string  // Auto-matches "database_host" in YAML
    APIKey        string `yaml:"api_key"`  // Explicit tag still works
}
```

**How it works:**
1. First checks for explicit tag (e.g., `yaml:"api_key"`)
2. If no tag, tries snake_case version (e.g., `ServerAddr` → `server_addr`)
3. If field not found, skips and applies defaults

**config.yaml:**
```yaml
server_addr: localhost:8080
shutdown_secs: 30
database_host: db.example.com
api_key: secret123
```

### Optional vs Required Files

- **`FromYAML()`** — File must exist. Returns error if missing.
- **`FromYAMLOptional()`** — File is optional (v0.10+). Returns empty source if file missing. Useful for optional config files with environment variable overrides.

**Example: Development vs Production setup**
```go
cfg, err := confkit.Load[Config](
    confkit.FromEnv(),                                    // Highest priority - runtime overrides
    confkit.FromYAMLOptional("config.local.yaml"),       // Optional local overrides (git-ignored)
    confkit.FromYAML("config.yaml"),                     // Required base config
)
```

If `config.local.yaml` doesn't exist, loading continues without error.

---

## JSON Files

**Function:** `FromJSON(path string)`

Load configuration from JSON files. **Supports automatic snake_case mapping (v0.10+)**.

```go
cfg, err := confkit.Load[Config](confkit.FromJSON("config.json"))
```

**config.json:**
```json
{
  "port": 8080,
  "database_url": "postgres://localhost/db",
  "log_level": "info"
}
```

Go struct with `DatabaseURL` field automatically matches `database_url` in JSON.

---

## TOML Files

**Function:** `FromTOML(path string)`

Load configuration from TOML files. **Supports automatic snake_case mapping (v0.10+)**.

```go
cfg, err := confkit.Load[Config](confkit.FromTOML("config.toml"))
```

**config.toml:**
```toml
port = 8080
database_url = "postgres://localhost/db"
log_level = "info"
```

Go struct with `DatabaseURL` field automatically matches `database_url` in TOML.

---

## Command-Line Flags

**Function:** `FromFlags()`

Load configuration from command-line flags. Field names are converted to kebab-case.

```go
type Config struct {
    Port     int    `flag:"port"`
    Database string `flag:"database-url"`
}

cfg, err := confkit.Load[Config](confkit.FromFlags())
```

**Run:**
```bash
go run main.go --port=3000 --database-url="postgres://..."
```

---

## Kubernetes ConfigMap

**Function:** `k8s.FromKubernetesConfigMap(namespace, name string)`

Load configuration from a Kubernetes ConfigMap.

```go
import "github.com/MimoJanra/confkit/k8s"

cfg, err := confkit.Load[Config](
    k8s.FromKubernetesConfigMap("default", "app-config"),
)
```

**ConfigMap:**
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: default
data:
  port: "8080"
  database: "postgres://..."
```

---

## AWS Systems Manager Parameter Store

**Function:** `aws.FromAWSSSMParameterStore(parameterPath string)`

Load configuration from AWS Systems Manager Parameter Store.

```go
import "github.com/MimoJanra/confkit/aws"

cfg, err := confkit.Load[Config](
    aws.FromAWSSSMParameterStore("/prod/app/config"),
)
```

**Parameters:**
```
/prod/app/config/port = 8080
/prod/app/config/database = postgres://...
```

### Hierarchical Parameters

Use `/` for nested structures:

```go
// Loads from: /app/db/host, /app/db/port
cfg, err := confkit.Load[Config](
    aws.FromAWSSSMParameterStore("/app"),
)
```

---

## HashiCorp Vault

**Function:** `vault.FromVault(addr string, auth vault.AuthMethod, secretPath string)`

Load secrets from HashiCorp Vault.

```go
import (
    "os"
    "github.com/MimoJanra/confkit/vault"
)

auth := vault.VaultTokenAuth(os.Getenv("VAULT_TOKEN"))

cfg, err := confkit.Load[Config](
    vault.FromVault("https://vault.example.com", auth, "/secret/myapp"),
)
```

**Vault Secret:**
```
path: secret/myapp
data:
  port: 8080
  database: postgres://...
  api_key: ***
```

---

## AWS Secrets Manager

**Function:** `aws.FromAWSSecretsManager(secretName string)`

Load secrets from AWS Secrets Manager.

```go
import "github.com/MimoJanra/confkit/aws"

cfg, err := confkit.Load[Config](
    aws.FromAWSSecretsManager("prod/app-secrets"),
)
```

---

## Consul KV

**Function:** `consul.FromConsul(addr string)`

Load configuration from Consul KV store.

```go
import "github.com/MimoJanra/confkit/consul"

cfg, err := confkit.Load[Config](
    consul.FromConsul("consul.example.com:8500"),
)
```

### With Options

```go
cfg, err := confkit.Load[Config](
    consul.FromConsulWithOptions(
        "consul.example.com:8500",
        "your-token",
        "dc1", // datacenter
    ),
)
```

---

## etcd

**Function:** `etcd.FromEtcd(endpoints []string)`

Load configuration from etcd v3.

```go
import "github.com/MimoJanra/confkit/etcd"

cfg, err := confkit.Load[Config](
    etcd.FromEtcd([]string{
        "etcd1.example.com:2379",
        "etcd2.example.com:2379",
    }),
)
```

### With Prefix

```go
cfg, err := confkit.Load[Config](
    etcd.FromEtcdWithPrefix(
        []string{"etcd.example.com:2379"},
        "/myapp",
    ),
)
```

---

## Multi-File Sources

When you need to merge several files of the same format (e.g., a base config + environment overlay + local overrides), use the `*Files` variants. Later files override earlier ones; nested maps are merged recursively (deep merge).

### FromYAMLFiles

```go
cfg, err := confkit.Load[Config](
    confkit.FromYAMLFiles(
        "config/base.yaml",       // shipped defaults
        "config/production.yaml", // environment overlay
        "config/local.yaml",      // developer overrides (git-ignored)
    ),
    confkit.FromEnv(),
)
```

### FromJSONFiles

```go
cfg, err := confkit.Load[Config](
    confkit.FromJSONFiles("base.json", "override.json"),
)
```

### FromTOMLFiles

```go
cfg, err := confkit.Load[Config](
    confkit.FromTOMLFiles("base.toml", "local.toml"),
)
```

**Deep merge behaviour:** if both files define a nested map key, the maps are merged. Scalar values in the later file win.

---

## Source Precedence

Sources use **first-wins** semantics: the first source that provides a value for a field wins. Sources listed later are only consulted for fields not yet set by an earlier source. Put your highest-priority sources first.

```go
cfg, err := confkit.Load[Config](
    confkit.FromEnv(),                     // 1st: highest priority — runtime overrides
    confkit.FromYAML("config.yaml"),       // 2nd: env-specific config
    confkit.FromYAML("defaults.yaml"),     // 3rd: base defaults (fallback)
)
```

**Example:**
- `defaults.yaml` has `port: 8080`
- `config.yaml` has `port: 3000`
- Environment has `PORT=9000`
- Result: `port = 9000` — `PORT=9000` from env wins because it is the first source

---

## Custom Sources

Implement the `Source` interface to create custom sources:

```go
type MySource struct {
    data map[string]string
}

func (s *MySource) Name() string {
    return "custom"
}

func (s *MySource) Lookup(ctx context.Context, field *confkit.FieldInfo) (any, bool, error) {
    val, ok := s.data[field.Path]
    return val, ok, nil
}

cfg, err := confkit.Load[Config](&MySource{
    data: map[string]string{
        "Port": "8080",
        "Database": "postgres://...",
    },
})
```

---

## Best Practices

### Source Order

Sources are first-wins: list the highest-priority source first.

1. **Start with runtime overrides** — Environment variables or flags for highest-priority values
2. **Add environment-specific** — Different configs for dev/staging/prod
3. **End with defaults** — Files with base defaults as a final fallback

### Secrets

- Use cloud secret managers (Vault, AWS Secrets Manager)
- Mark sensitive fields with `secret:"true"`
- Never commit secrets to version control
- Use separate secrets source from configuration

### Files

- Keep config files in version control
- Use YAML for readability
- Use TOML for complex structures
- Use JSON for programmatic generation

### Environment Variables

- Use for runtime overrides
- Use `prefix:` for organized namespacing
- Use ALL_CAPS naming convention
- Document available variables in README

---

## See Also

- [Getting Started](/docs/getting-started/) — Quick setup guide
- [API Reference](/api/) — Complete function reference
- **[Examples](https://github.com/MimoJanra/confkit/tree/main/examples)** — Production-ready examples with all sources
  - Web service with database and cache
  - Microservice with Postgres, Redis, RabbitMQ
  - CLI tool with file I/O
  - Cloud-native with Kubernetes and AWS
