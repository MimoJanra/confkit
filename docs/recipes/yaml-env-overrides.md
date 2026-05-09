---
layout: default
title: "Recipe: YAML + Environment Overrides"
---

# Recipe: YAML + Environment Overrides

Load a base config from YAML, then allow environment variables to override specific values.

## Use Case

This is the most common pattern: have sensible defaults in a config file, but allow operators to override them with environment variables.

## Code

```go
package main

import (
    "log"
    "github.com/MimoJanra/confkit"
)

type Config struct {
    Server struct {
        Host string `env:"HOST" default:"0.0.0.0"`
        Port int    `env:"PORT" default:"8080" validate:"min=1,max=65535"`
    }
    Database struct {
        Host     string `env:"HOST" default:"localhost"`
        Port     int    `env:"PORT" default:"5432"`
        User     string `env:"USER" validate:"required"`
        Password string `env:"PASSWORD" validate:"required" secret:"true"`
    } `prefix:"DB_"`
}

func main() {
    cfg, err := confkit.Load[Config](
        confkit.FromEnv(),                 // Highest priority — env vars checked first
        confkit.FromYAML("config.yaml"),  // Fallback: values from file fill in unset fields
    )
    if err != nil {
        log.Fatal(confkit.Explain(err))
    }
    
    log.Printf("Server: %s:%d", cfg.Server.Host, cfg.Server.Port)
    log.Printf("Database: %s@%s:%d", cfg.Database.User, cfg.Database.Host, cfg.Database.Port)
}
```

## config.yaml

```yaml
server:
  host: 0.0.0.0
  port: 8080

database:
  host: localhost
  port: 5432
  user: app
  password: initial-password-from-file
```

## Usage

### Default (from YAML)

```bash
go run main.go
# Output:
# Server: 0.0.0.0:8080
# Database: app@localhost:5432
```

### Override with Environment Variables

```bash
export HOST=127.0.0.1
export PORT=9090
export DB_HOST=db.prod.internal
export DB_USER=produser
export DB_PASSWORD=prodpass

go run main.go
# Output:
# Server: 127.0.0.1:9090
# Database: produser@db.prod.internal:5432
```

Port from file (5432) is used because `DB_PORT` was not set in env.

## Load Order

The first source to provide a value wins; later sources fill in only unset fields:

1. **Environment variables checked first** — Any env var matching a tag is used immediately
2. **YAML file fills in the rest** — Fields not set by env vars receive their YAML values
3. **Defaults applied** — Only if neither source provided a value
4. **Validation** — All values validated against rules

## Common Patterns

### Multiple Environment Overrides

```bash
export HOST=prod-api.example.com
export PORT=443
export DB_HOST=prod-db.internal
export DB_USER=produser
export DB_PASSWORD=very-secret-password
go run main.go
```

### Partial Overrides

```bash
# Only override the port, keep everything else from YAML
export PORT=3000
go run main.go
```

### No Overrides (Pure YAML)

```bash
# Use only YAML defaults
go run main.go
```

## Docker Integration

In a Dockerfile:

```dockerfile
FROM golang:1.24 AS builder
WORKDIR /app
COPY . .
RUN go build -o app main.go

FROM alpine:latest
COPY --from=builder /app/app /app
COPY --from=builder /app/config.yaml /config.yaml
CMD ["/app"]
```

Run with environment variable overrides:

```bash
docker run -e HOST=0.0.0.0 -e PORT=8080 -e DB_HOST=postgres myapp:latest
```

## Kubernetes Integration

Define config as ConfigMap, override with environment from Secret:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  config.yaml: |
    server:
      host: 0.0.0.0
      port: 8080
    database:
      host: postgres
      port: 5432
      user: app
---
apiVersion: v1
kind: Secret
metadata:
  name: app-secrets
type: Opaque
stringData:
  DB_PASSWORD: very-secret-password
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
spec:
  template:
    spec:
      containers:
      - name: app
        image: myapp:latest
        volumeMounts:
        - name: config
          mountPath: /etc/app
        env:
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: DB_PASSWORD
      volumes:
      - name: config
        configMap:
          name: app-config
```

## Best Practices

1. **Use YAML for defaults**
   - Keep sensible defaults in version control
   - Makes it easy for new developers

2. **Use env vars for secrets**
   - Never commit secrets to version control
   - Inject at runtime via CI/CD, containers, or secret stores

3. **Use env vars for environment-specific settings**
   ```bash
   export ENVIRONMENT=production
   export DB_HOST=prod-db.internal
   ```

4. **Validate after loading**
   - Always include validation rules
   - Fail fast on invalid config

5. **Log the loaded config safely**
   ```go
   fields := confkit.ScanFields(cfg)
   dump, _ := confkit.DumpConfig(cfg, fields)
   log.Printf("Loaded config: %s", dump)
   // Secrets are redacted
   ```

## Troubleshooting

### Environment variable not being picked up

Ensure the tag name matches:

```go
type Config struct {
    Host string `env:"HOST"`  // ✅ Correct
}

// ❌ Wrong — will not pick up $MY_HOST
```

### Prefix not working

For nested structs, include `prefix`:

```go
type Config struct {
    Database struct {
        Host string `env:"HOST"`
    } `prefix:"DB_"`
}

// env DB_HOST=localhost will be used
```

### Default not applied

Check if environment variable is set:

```go
type Config struct {
    Port int `env:"PORT" default:"8080"`
}

// If $PORT is set in env, it overrides the default
unset PORT  # Make sure it's not set
export PORT=9090  # Now it's set to 9090
```

## See Also

- **[Environment Variables](../docs/getting-started.md)** — More on env var loading
- **[Defaults](../docs/defaults.md)** — How defaults work
- **[Hot Reload](../docs/hot-reload.md)** — Reloading YAML without restart
