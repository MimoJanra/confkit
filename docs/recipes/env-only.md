---
layout: default
title: "Recipe: Environment Variables Only"
---

# Recipe: Environment Variables Only

Load all configuration exclusively from environment variables. No config files needed.

## Use Case

- 12-factor app methodology
- Serverless/container deployments
- CI/CD pipelines
- Environment-specific deployments

## Code

```go
package main

import (
    "log"
    "github.com/MimoJanra/confkit"
)

type Config struct {
    App struct {
        Name    string `env:"APP_NAME" validate:"required"`
        Version string `env:"APP_VERSION" validate:"required"`
        Debug   bool   `env:"APP_DEBUG" default:"false"`
    }
    Server struct {
        Host string `env:"HOST" default:"0.0.0.0"`
        Port int    `env:"PORT" default:"8080" validate:"min=1,max=65535"`
    }
    Database struct {
        Host     string `env:"HOST" default:"localhost"`
        Port     int    `env:"PORT" default:"5432"`
        Name     string `env:"NAME" validate:"required"`
        User     string `env:"USER" validate:"required"`
        Password string `env:"PASSWORD" validate:"required" secret:"true"`
    } `prefix:"DB_"`
}

func main() {
    cfg, err := confkit.Load[Config](
        confkit.FromEnv(),
    )
    if err != nil {
        log.Fatal(confkit.Explain(err))
    }
    
    log.Printf("App: %s v%s (debug=%v)", cfg.App.Name, cfg.App.Version, cfg.App.Debug)
    log.Printf("Server: %s:%d", cfg.Server.Host, cfg.Server.Port)
    log.Printf("Database: %s://%s@%s:%d/%s", "postgres", cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
}
```

## Required Environment Variables

```bash
export APP_NAME=myapp
export APP_VERSION=1.0.0
export DB_NAME=myapp_db
export DB_USER=postgres
export DB_PASSWORD=secret
```

## Optional Environment Variables

```bash
# These have defaults and are optional
export APP_DEBUG=true
export HOST=127.0.0.1
export PORT=3000
export DB_HOST=postgres.internal
export DB_PORT=5432
```

## Full Example

```bash
#!/bin/bash

# Required variables
export APP_NAME=myapp
export APP_VERSION=1.0.0
export DB_NAME=myapp_db
export DB_USER=postgres
export DB_PASSWORD=secret

# Optional variables (with defaults)
export APP_DEBUG=true
export HOST=0.0.0.0
export PORT=8080
export DB_HOST=localhost
export DB_PORT=5432

go run main.go
```

## Docker Integration

### Dockerfile

```dockerfile
FROM golang:1.25 AS builder
WORKDIR /app
COPY . .
RUN go build -o app main.go

FROM alpine:latest
COPY --from=builder /app/app /app
ENTRYPOINT ["/app"]
```

### docker-compose.yml

```yaml
version: '3.8'
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: myapp_db
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: secret

  app:
    build: .
    depends_on:
      - postgres
    environment:
      APP_NAME: myapp
      APP_VERSION: 1.0.0
      APP_DEBUG: "true"
      HOST: 0.0.0.0
      PORT: 8080
      DB_HOST: postgres
      DB_PORT: 5432
      DB_NAME: myapp_db
      DB_USER: postgres
      DB_PASSWORD: secret
    ports:
      - "8080:8080"
```

Run:

```bash
docker-compose up
```

## Kubernetes Integration

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  APP_NAME: myapp
  APP_VERSION: "1.0.0"
  APP_DEBUG: "false"
  HOST: "0.0.0.0"
  PORT: "8080"
  DB_HOST: postgres
  DB_PORT: "5432"
  DB_NAME: myapp_db

---
apiVersion: v1
kind: Secret
metadata:
  name: app-secrets
type: Opaque
stringData:
  DB_USER: postgres
  DB_PASSWORD: very-secret-password

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
    spec:
      containers:
      - name: app
        image: myapp:latest
        ports:
        - containerPort: 8080
        envFrom:
        - configMapRef:
            name: app-config
        - secretRef:
            name: app-secrets
```

## Validation

Missing required variables cause clear errors:

```bash
# Don't set DB_NAME
unset DB_NAME
go run main.go

# Error:
# Invalid configuration:
#
#   Database.Name
#     error: field is required
#     source: env (DB_NAME)
```

## Environment Variable Expansion

Use shell variable expansion:

```bash
# AWS example
export AWS_REGION=us-east-1
export APP_NAME=myapp-$AWS_REGION
# APP_NAME → myapp-us-east-1

go run main.go
```

## 12-Factor App Pattern

This recipe follows the 12-factor app methodology:

✅ Configuration is stored in environment variables  
✅ Secrets are not in code  
✅ Deployment is environment-agnostic  
✅ No config files needed  

Read more: https://12factor.net/config

## Best Practices

1. **Document required variables**
   ```bash
   # .env.example
   APP_NAME=myapp
   APP_VERSION=1.0.0
   DB_NAME=myapp_db
   DB_USER=postgres
   DB_PASSWORD=
   ```

2. **Use defaults for optional variables**
   ```go
   Port int `env:"PORT" default:"8080"`
   ```

3. **Validate required variables**
   ```go
   AppName string `env:"APP_NAME" validate:"required"`
   ```

4. **Mark secrets**
   ```go
   Password string `env:"DB_PASSWORD" secret:"true"`
   ```

5. **Provide helpful error messages**
   ```go
   if err != nil {
       log.Fatal(confkit.Explain(err))
   }
   ```

## Troubleshooting

### Variable not found

Check it's exported:

```bash
# ❌ Not exported
MY_VAR=value
go run main.go

# ✅ Exported
export MY_VAR=value
go run main.go
```

### Wrong prefix for nested structs

```go
type Config struct {
    Database struct {
        Host string `env:"HOST"`
    } `prefix:"DB_"`
}

# Use DB_HOST, not DATABASE_HOST
export DB_HOST=localhost
```

### Validation fails

Ensure the value matches validation rules:

```go
Port int `validate:"min=1,max=65535"`

# ❌ Invalid
export PORT=99999

# ✅ Valid
export PORT=8080
```

## See Also

- **[Getting Started](../docs/getting-started.md)** — Loading from env
- **[Validation](../docs/validation.md)** — Validation rules
- **[Secret Redaction](../docs/secret-redaction.md)** — Hiding secrets
