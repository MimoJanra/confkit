---
layout: default
title: Examples
---

# Examples

Runnable code examples for common confkit use cases.

## Basic Configuration

```go
package main

import (
    "log"
    "github.com/MimoJanra/confkit"
)

type Config struct {
    Port     int    `env:"PORT" default:"8080" validate:"min=1,max=65535"`
    Host     string `env:"HOST" default:"localhost"`
    Database string `env:"DATABASE_URL" validate:"required" secret:"true"`
}

func main() {
    cfg, err := confkit.Load[Config](confkit.FromEnv())
    if err != nil {
        log.Fatal(confkit.Explain(err))
    }
    
    log.Printf("Starting on %s:%d\n", cfg.Host, cfg.Port)
}
```

**Run:**
```bash
export DATABASE_URL="postgres://user:pass@localhost/db"
go run main.go
```

---

## Multiple Sources with Precedence

```go
type Config struct {
    Port     int    `yaml:"port" env:"PORT" default:"8080"`
    Database string `yaml:"database" env:"DATABASE_URL"`
    LogLevel string `yaml:"log_level" env:"LOG_LEVEL" default:"info"`
}

func main() {
    // Precedence: env vars override YAML override defaults
    cfg, err := confkit.Load[Config](
        confkit.FromYAML("config.yaml"),    // Base config
        confkit.FromEnv(),                  // Runtime overrides
    )
    if err != nil {
        log.Fatal(confkit.Explain(err))
    }
}
```

**config.yaml:**
```yaml
port: 3000
database: postgres://localhost/mydb
log_level: debug
```

---

## Custom Validation

```go
func main() {
    cfg, err := confkit.LoadWithOptions[Config](
        confkit.WithSource(confkit.FromEnv()),
        confkit.WithValidator("email", func(val string) error {
            if !strings.Contains(val, "@") {
                return fmt.Errorf("must be a valid email address")
            }
            return nil
        }),
    )
}

type Config struct {
    AdminEmail string `env:"ADMIN_EMAIL" validate:"custom=email"`
}
```

---

## Nested Structures

```go
type DatabaseConfig struct {
    Host     string `yaml:"host" env:"DB_HOST" default:"localhost"`
    Port     int    `yaml:"port" env:"DB_PORT" default:"5432"`
    Username string `yaml:"user" env:"DB_USER"`
    Password string `yaml:"password" env:"DB_PASSWORD" secret:"true"`
}

type CacheConfig struct {
    Enabled bool   `yaml:"enabled" env:"CACHE_ENABLED" default:"true"`
    TTL     int    `yaml:"ttl" env:"CACHE_TTL" default:"3600"`
}

type Config struct {
    Server ServerConfig
    DB     DatabaseConfig
    Cache  CacheConfig
    LogLevel string `env:"LOG_LEVEL" default:"info"`
}

func main() {
    cfg, err := confkit.Load[Config](
        confkit.FromYAML("config.yaml"),
        confkit.FromEnv(),
    )
    // Access nested: cfg.DB.Host, cfg.Cache.TTL, etc.
}
```

---

## Environment Prefixes

```go
type DatabaseConfig struct {
    Host string `env:"HOST" default:"localhost"`
    Port int    `env:"PORT" default:"5432"`
}

type Config struct {
    Server ServerConfig
    // Fields in this struct automatically get DB_ prefix
    DB DatabaseConfig `envPrefix:"DB_"`
}

func main() {
    // Reads from: DB_HOST, DB_PORT (not just HOST, PORT)
    cfg, err := confkit.Load[Config](confkit.FromEnv())
}
```

**Environment:**
```bash
DB_HOST=db.example.com
DB_PORT=3306
```

---

## String Interpolation

```go
type Config struct {
    AppName    string `env:"APP_NAME" default:"MyApp"`
    AppVersion string `env:"APP_VERSION" default:"1.0.0"`
    // Will expand to "MyApp v1.0.0"
    AppTitle   string `env:"APP_TITLE" default:"${APP_NAME} v${APP_VERSION}"`
    
    BaseURL    string `env:"BASE_URL" default:"https://api.example.com"`
    // Will expand to full URL
    UsersURL   string `env:"USERS_URL" default:"${BASE_URL}/users"`
}
```

---

## Hot Reload / File Watching

```go
func main() {
    cfg, watcher, err := confkit.LoadWithWatcher[Config](
        "config.yaml",
        confkit.FromYAML("config.yaml"),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Watch for changes
    go watchConfigChanges(watcher)
    
    // Your application runs here
    startServer(cfg)
}

func watchConfigChanges(watcher *confkit.ConfigWatcher) {
    for newCfg := range watcher.Changes() {
        log.Printf("Config reloaded!\n")
        log.Printf("New settings: %+v\n", newCfg)
        // Apply new config to running services
    }
}
```

---

## Secret Redaction

Automatically redacts sensitive fields in errors and logs:

```go
type Config struct {
    APIKey   string `env:"API_KEY" validate:"required" secret:"true"`
    Password string `env:"PASSWORD" validate:"required" secret:"true"`
    Username string `env:"USERNAME"`
}

func main() {
    cfg, err := confkit.Load[Config](confkit.FromEnv())
    if err != nil {
        // Secrets are automatically redacted
        fmt.Println(confkit.Explain(err))
        // Output:
        // APIKey
        //   error: field is required
        //   source: env (API_KEY=***)
    }
}
```

---

## Kubernetes ConfigMap

```go
func main() {
    cfg, err := confkit.Load[Config](
        confkit.FromK8sConfigMap("default", "app-config"),
    )
}
```

**ConfigMap:**
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  port: "8080"
  log_level: "info"
```

---

## AWS Systems Manager Parameter Store

```go
func main() {
    cfg, err := confkit.Load[Config](
        confkit.FromAWSSSM("/prod/app/config"),
    )
}
```

---

## HashiCorp Vault

```go
func main() {
    cfg, err := confkit.Load[Config](
        confkit.FromVault("https://vault.example.com", "/secret/app"),
    )
}
```

---

## AWS Secrets Manager

```go
func main() {
    cfg, err := confkit.Load[Config](
        confkit.FromAWSSecretsManager("prod/app-secrets"),
    )
}
```

---

## Consul KV

```go
func main() {
    cfg, err := confkit.Load[Config](
        confkit.FromConsul("consul.example.com:8500"),
    )
}
```

---

## etcd

```go
func main() {
    cfg, err := confkit.Load[Config](
        confkit.FromEtcd([]string{"etcd1.example.com:2379"}),
    )
}
```

---

## Advanced: Custom Source

```go
type CustomSource struct {
    data map[string]string
}

func (s *CustomSource) Name() string {
    return "custom"
}

func (s *CustomSource) Lookup(field *confkit.FieldInfo) (confkit.Value, bool, error) {
    val, ok := s.data[field.Path]
    return val, ok, nil
}

func main() {
    source := &CustomSource{
        data: map[string]string{
            "Port":     "9000",
            "Database": "postgres://localhost/db",
        },
    }
    
    cfg, err := confkit.Load[Config](source)
}
```

---

## Full Production Setup

```go
package main

import (
    "context"
    "log"
    "github.com/MimoJanra/confkit"
)

type DatabaseConfig struct {
    Host     string        `yaml:"host" env:"DB_HOST" default:"localhost"`
    Port     int           `yaml:"port" env:"DB_PORT" default:"5432" validate:"min=1,max=65535"`
    Username string        `yaml:"username" env:"DB_USER" validate:"required"`
    Password string        `yaml:"password" env:"DB_PASSWORD" validate:"required" secret:"true"`
    SSL      bool          `yaml:"ssl" env:"DB_SSL" default:"true"`
    Timeout  time.Duration `yaml:"timeout" env:"DB_TIMEOUT" default:"30s"`
}

type ServerConfig struct {
    Host        string `yaml:"host" env:"SERVER_HOST" default:"0.0.0.0"`
    Port        int    `yaml:"port" env:"SERVER_PORT" default:"8080" validate:"min=1,max=65535"`
    ReadTimeout time.Duration `yaml:"read_timeout" env:"SERVER_READ_TIMEOUT" default:"10s"`
}

type Config struct {
    Database DatabaseConfig
    Server   ServerConfig
    LogLevel string `yaml:"log_level" env:"LOG_LEVEL" validate:"oneof=debug,info,warn,error" default:"info"`
    Environment string `yaml:"env" env:"ENVIRONMENT" validate:"oneof=dev,staging,prod" default:"dev"`
}

func main() {
    // Load with hot reload
    cfg, watcher, err := confkit.LoadWithWatcher[Config](
        "config.yaml",
        confkit.FromYAML("config.yaml"),
        confkit.FromEnv(),
    )
    if err != nil {
        log.Fatal(confkit.Explain(err))
    }
    
    log.Printf("Configuration loaded from %s\n", cfg.Environment)
    
    // Watch for changes
    go func() {
        for newCfg := range watcher.Changes() {
            log.Printf("Configuration reloaded\n")
            // Reconfigure running services
            updateServices(newCfg)
        }
    }()
    
    // Start application
    startServer(cfg)
}

func updateServices(cfg Config) {
    // Apply new configuration
}

func startServer(cfg Config) {
    log.Printf("Starting server at %s:%d\n", cfg.Server.Host, cfg.Server.Port)
    // Your application logic
}
```

---

## More Examples

- See the `/examples` folder in the repository for complete runnable code
- Check the README.md for quick reference examples
- Read the iteration docs for feature-specific examples
