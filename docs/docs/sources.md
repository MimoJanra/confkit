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

Use `envPrefix` tag to automatically add a prefix to all env vars in a struct:

```go
type DatabaseConfig struct {
    Host string `env:"HOST"`
    Port int    `env:"PORT"`
}

type Config struct {
    DB DatabaseConfig `envPrefix:"DB_"`
}

// Reads from DB_HOST, DB_PORT
cfg, err := confkit.Load[Config](confkit.FromEnv())
```

---

## YAML Files

**Function:** `FromYAML(path string)`

Load configuration from YAML files. Field names use snake_case by default.

```go
type Config struct {
    Port     int    `yaml:"port"`
    Database string `yaml:"database"`
}

cfg, err := confkit.Load[Config](confkit.FromYAML("config.yaml"))
```

**config.yaml:**
```yaml
port: 8080
database: postgres://localhost/db
```

---

## JSON Files

**Function:** `FromJSON(path string)`

Load configuration from JSON files.

```go
cfg, err := confkit.Load[Config](confkit.FromJSON("config.json"))
```

**config.json:**
```json
{
  "port": 8080,
  "database": "postgres://localhost/db"
}
```

---

## TOML Files

**Function:** `FromTOML(path string)`

Load configuration from TOML files.

```go
cfg, err := confkit.Load[Config](confkit.FromTOML("config.toml"))
```

**config.toml:**
```toml
port = 8080
database = "postgres://localhost/db"
```

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

**Function:** `FromK8sConfigMap(namespace, name string)`

Load configuration from a Kubernetes ConfigMap.

```go
cfg, err := confkit.Load[Config](
    confkit.FromK8sConfigMap("default", "app-config"),
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

### With Secrets

Use `FromK8sSecret()` for Kubernetes Secrets:

```go
cfg, err := confkit.Load[Config](
    confkit.FromK8sSecret("default", "app-secrets"),
)
```

---

## AWS Systems Manager Parameter Store

**Function:** `FromAWSSSM(parameterPath string)`

Load configuration from AWS Systems Manager Parameter Store.

```go
cfg, err := confkit.Load[Config](
    confkit.FromAWSSSM("/prod/app/config"),
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
    confkit.FromAWSSSM("/app"),
)
```

---

## HashiCorp Vault

**Function:** `FromVault(addr, secretPath string)`

Load secrets from HashiCorp Vault.

```go
cfg, err := confkit.Load[Config](
    confkit.FromVault("https://vault.example.com:8200", "/secret/app"),
)
```

**Vault Secret:**
```
path: secret/app
data:
  port: 8080
  database: postgres://...
  api_key: ***
```

### With Authentication

```go
vault := confkit.NewVaultSource("https://vault.example.com", "/secret/app")
vault.SetToken("s.xxxxxxxxxxxx")

cfg, err := confkit.Load[Config](vault)
```

---

## AWS Secrets Manager

**Function:** `FromAWSSecretsManager(secretName string)`

Load secrets from AWS Secrets Manager.

```go
cfg, err := confkit.Load[Config](
    confkit.FromAWSSecretsManager("prod/app-secrets"),
)
```

---

## Consul KV

**Function:** `FromConsul(addr string)`

Load configuration from Consul KV store.

```go
cfg, err := confkit.Load[Config](
    confkit.FromConsul("consul.example.com:8500"),
)
```

### With Options

```go
consul := confkit.FromConsulWithOptions(
    "consul.example.com:8500",
    "your-token",
    "dc1", // datacenter
)

cfg, err := confkit.Load[Config](consul)
```

---

## etcd

**Function:** `FromEtcd(endpoints []string)`

Load configuration from etcd v3.

```go
cfg, err := confkit.Load[Config](
    confkit.FromEtcd([]string{
        "etcd1.example.com:2379",
        "etcd2.example.com:2379",
    }),
)
```

### With Prefix

```go
etcd := confkit.FromEtcdWithPrefix(
    []string{"etcd.example.com:2379"},
    "/myapp",
)

cfg, err := confkit.Load[Config](etcd)
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

When using multiple sources, they are evaluated left-to-right. Later sources override earlier ones:

```go
cfg, err := confkit.Load[Config](
    confkit.FromYAML("defaults.yaml"),     // 1st: base defaults
    confkit.FromYAML("config.yaml"),       // 2nd: env-specific
    confkit.FromEnv(),                     // 3rd: runtime overrides
)
```

**Example:**
- `defaults.yaml` has `port: 8080`
- `config.yaml` has `port: 3000`
- Environment has `PORT=9000`
- Result: `port = 9000` (highest precedence)

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

func (s *MySource) Lookup(field *confkit.FieldInfo) (confkit.Value, bool, error) {
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

1. **Start with defaults** — Files or defaults for base configuration
2. **Add environment-specific** — Different configs for dev/staging/prod
3. **Override with runtime** — Environment variables or flags for final tweaks

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
- Use `envPrefix` for organized namespacing
- Use ALL_CAPS naming convention
- Document available variables in README

---

## See Also

- [Getting Started](/docs/getting-started/) — Quick setup guide
- [API Reference](/api/) — Complete function reference
- [Examples](/examples/) — More source usage examples
