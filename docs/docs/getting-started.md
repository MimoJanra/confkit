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

**Requirements:** Go 1.22 or later

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

## Next Steps

- [Configuration Sources](./sources/) — Learn about all available sources
- [Validation Rules](./validation/) — Deep dive into validation
- [Error Handling](./errors/) — Programmatically handle errors
- [Cloud Integrations](./cloud/) — Use Kubernetes, AWS, Vault, and more
