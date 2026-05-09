---
layout: default
title: Getting Started
---

# Getting Started with confkit

Install confkit and load your first configuration in 5 minutes.

## Installation

```bash
go get github.com/MimoJanra/confkit@latest
```

**Requirements:** Go 1.24.0 or later

## Your First Config

### Step 1: Define Your Config

Create a struct with tags describing where values come from:

```go
package main

import (
    "log"
    "github.com/MimoJanra/confkit"
)

type Config struct {
    // Load from PORT env var, default to 8080
    Port int `env:"PORT" default:"8080" validate:"min=1,max=65535"`
    
    // Load from DATABASE_URL env var, required
    DatabaseURL string `env:"DATABASE_URL" validate:"required" secret:"true"`
    
    // Load from LOG_LEVEL, with default
    LogLevel string `env:"LOG_LEVEL" default:"info"`
}
```

### Step 2: Load and Use

```go
func main() {
    cfg, err := confkit.Load[Config](
        confkit.FromEnv(),
    )
    if err != nil {
        log.Fatal(confkit.Explain(err))
    }
    
    // cfg is fully typed and validated
    log.Printf("Starting server on port %d\n", cfg.Port)
    log.Printf("Connecting to %s\n", cfg.DatabaseURL) // secret is redacted in logs
}
```

### Step 3: Set Environment Variables and Run

```bash
export PORT=3000
export DATABASE_URL="postgres://user:pass@localhost/db"
export LOG_LEVEL="debug"

go run main.go
# Output: Starting server on port 3000
# Output: Connecting to postgres://user:***@localhost/db
```

## Using Configuration Files

Load from YAML, JSON, or TOML files:

```go
cfg, err := confkit.Load[Config](
    confkit.FromEnv(),                 // env vars take priority
    confkit.FromYAML("config.yaml"),   // file provides fallback values
)
```

**config.yaml:**
```yaml
port: 3000
databaseURL: postgres://localhost/db
logLevel: debug
```

## Validation

confkit validates configuration at load time:

```go
type Config struct {
    Port int `validate:"min=1,max=65535"`
    Count int `validate:"min=0"`
    Status string `validate:"oneof=active,inactive,pending"`
}
```

**Error on validation failure:**
```
Invalid configuration:

  Port
    error: must be between 1 and 65535
    got: 99999
    source: env (PORT)

  Status
    error: must be one of: active, inactive, pending
    got: unknown
    source: yaml (config.yaml)
```

## Supported Types

confkit supports a wide range of types out of the box:

```go
type Config struct {
    // Primitives
    Port     int       `env:"PORT"`
    Enabled  bool      `env:"ENABLED"`
    Name     string    `env:"NAME"`
    Factor   float64   `env:"FACTOR"`
    
    // Time durations (e.g., "5s", "10m", "1h30m")
    Timeout  time.Duration `env:"TIMEOUT" default:"30s"`
    
    // Collections
    AllowedHosts []string `env:"ALLOWED_HOSTS"` // comma-separated
}
```

**Supported types:**
- `string`, `int`, `int8`–`int64`, `uint`, `uint8`–`uint64`
- `float32`, `float64`
- `bool` (accepts: true/1/yes/on or false/0/no/off)
- `time.Duration` (Go duration format: "5s", "10m", "1h30m")
- `time.Time` (RFC3339 format: "2026-01-01T00:00:00Z")
- `[]string`, `[]int` (from comma-separated values)
- `map[string]string`, `map[string]int`, etc. (KEY=val,KEY2=val2 format)
- Nested structs with `prefix:` tags
- Custom types with custom validators

### Time Duration Examples

```go
type Config struct {
    // All of these work:
    ShortTimeout  time.Duration `env:"SHORT_TIMEOUT" default:"5s"`
    MediumTimeout time.Duration `env:"MEDIUM_TIMEOUT" default:"30s"`
    LongTimeout   time.Duration `env:"LONG_TIMEOUT" default:"5m"`
    VeryLong      time.Duration `env:"VERY_LONG" default:"1h30m"`
}

// Load from env or defaults:
// SHORT_TIMEOUT="10s" → 10 seconds
// MEDIUM_TIMEOUT not set → 30 seconds (default)
// LONG_TIMEOUT="2m30s" → 2 minutes 30 seconds
```

## Secrets & Security

Mark sensitive fields with `secret:"true"` to redact them:

```go
type Config struct {
    // This will be redacted in errors, logs, and dumps
    APIKey string `env:"API_KEY" validate:"required" secret:"true"`
}
```

**In error messages:**
```
APIKey
  error: field is required
  source: env (API_KEY=***)
```

## Multiple Sources & Precedence

Sources use **first-wins** semantics: the first source that provides a value for a field wins. List your highest-priority source first.

```go
cfg, err := confkit.Load[Config](
    confkit.FromEnv(),                    // highest priority — runtime overrides
    confkit.FromYAML("config.yaml"),      // fallback
    confkit.FromYAML("defaults.yaml"),    // base defaults
)
```

## Real-World Examples

Check out the **[examples](https://github.com/MimoJanra/confkit/tree/main/examples)** directory for complete, production-ready examples:

- **Web Service** — Typical API with database, cache, logging
- **Microservice** — Enterprise setup with PostgreSQL, Redis, RabbitMQ, observability
- **CLI Tool** — Command-line tool with flags and file processing
- **Cloud-Native** — Kubernetes + AWS + Vault integration with health checks and mTLS
- **Full Setup** — Schema generation and comprehensive feature demo

Each example includes:
- Complete struct definitions with all field types
- Multiple configuration sources (env, YAML, defaults)
- Comprehensive test suite
- Example configuration files

Run tests to see them in action:
```bash
go test ./examples -v
```

## Next Steps

- [Configuration Sources](./sources/) — Learn about all available sources
- [Validation Rules](./validation/) — Deep dive into validation
- [Error Handling](./errors/) — Programmatically handle errors
- [Cloud Integrations](./cloud/) — Use Kubernetes, AWS, Vault, and more
- **[Examples](https://github.com/MimoJanra/confkit/tree/main/examples)** — Production-ready code samples
