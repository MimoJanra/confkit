---
layout: default
title: Installation — confkit
---

# Installation

## Core Library

```bash
go get github.com/MimoJanra/confkit@latest
```

Then import:

```go
import "github.com/MimoJanra/confkit"
```

## Optional Cloud Modules

confkit's cloud integrations are separate optional modules. Install only what you need:

### Kubernetes

```bash
go get github.com/MimoJanra/confkit/k8s@latest
```

```go
import "github.com/MimoJanra/confkit/k8s"
```

### HashiCorp Vault

```bash
go get github.com/MimoJanra/confkit/vault@latest
```

```go
import "github.com/MimoJanra/confkit/vault"
```

### Consul

```bash
go get github.com/MimoJanra/confkit/consul@latest
```

```go
import "github.com/MimoJanra/confkit/consul"
```

### etcd

```bash
go get github.com/MimoJanra/confkit/etcd@latest
```

```go
import "github.com/MimoJanra/confkit/etcd"
```

### AWS (SSM Parameter Store & Secrets Manager)

```bash
go get github.com/MimoJanra/confkit/aws@latest
```

```go
import "github.com/MimoJanra/confkit/aws"
```

## Minimum Go Version

confkit requires **Go 1.25.0 or later**.

**Why Go 1.25.0?**
- Generics for type-safe `Load[T]` API
- Improved `range` statement (iterate over integers and slices with range-only syntax)
- Enhanced iteration semantics

Check your Go version:

```bash
go version
```

If you have an older version, update it:

```bash
go install golang.org/dl/go1.25@latest
go1.25 download
```

## Dependencies

### Core Module

Core confkit depends only on:

- `gopkg.in/yaml.v3` — YAML parsing
- `github.com/pelletier/go-toml/v2` — TOML parsing
- Go standard library

No external dependencies for JSON parsing (uses `encoding/json` from stdlib).

### Cloud Modules

Each optional module brings only its necessary SDK:

- **k8s:** `client-go` and related Kubernetes libraries
- **vault:** HashiCorp Vault Go client
- **consul:** HashiCorp Consul Go client
- **etcd:** etcd Go client
- **aws:** AWS SDK for Go v2

Cloud modules are intentionally separated so your binary doesn't bloat if you don't need them.

## Verification

Verify successful installation:

```go
package main

import (
    "log"
    "github.com/MimoJanra/confkit"
)

type Config struct {
    Port int `env:"PORT" default:"8080"`
}

func main() {
    cfg, err := confkit.Load[Config](confkit.FromEnv())
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Port: %d", cfg.Port)  // cfg is *Config pointer (v1.0+)
}
```

Run it:

```bash
go run main.go
# Output: Port: 8080
```

## Version Notes

**v1.0.0 (current stable):**
- `Load[T]` returns `(*T, error)` instead of `(T, error)` — pointer return (breaking change from v0.9)
- Most code works unchanged due to Go auto-dereference behavior
- See [Migration Guide](./migration-v1.0.md) for details

**v0.9:**
- `Load[T]` returned `(T, error)` — value return
- If upgrading from v0.9, see the migration guide

## Updating confkit

To update to the latest version:

```bash
go get -u github.com/MimoJanra/confkit
go get -u github.com/MimoJanra/confkit/vault  # or any submodule
```

## Troubleshooting

### `go get` fails with "module not found"

Ensure your module has a valid `go.mod`:

```bash
go mod init myapp
go get github.com/MimoJanra/confkit
```

### Import errors after installation

Make sure you're using the correct import paths:

```go
// ✅ Correct
import "github.com/MimoJanra/confkit"

// ❌ Wrong
import "confkit"  // This will fail
```

### Cloud module import errors

For cloud sources, import the submodule:

```go
// ✅ Correct
import "github.com/MimoJanra/confkit/vault"

// ❌ Wrong
import "github.com/MimoJanra/confkit"  // Won't find vault functions
```

### Version conflicts

If you see version conflicts with other dependencies (especially AWS SDK):

```bash
go mod tidy
go get github.com/MimoJanra/confkit@v1.0.0  # pin explicit version
```

## Next Steps

- **[Getting Started](./getting-started.md)** — 5-minute quick start
- **[Struct Tags](../docs/defaults.md)** — `env`, `flag`, `default`, `validate`, `secret`
- **[Sources](./sources.md)** — All available configuration sources
- **[Recipes](../recipes/)** — Real-world examples
