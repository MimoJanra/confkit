---
layout: default
title: "Recipe: etcd"
---

# Recipe: etcd

Load configuration from etcd v3 distributed key-value store.

## Installation

```bash
go get github.com/MimoJanra/confkit/etcd@latest
```

## Code

```go
package main

import (
    "context"
    "log"
    "github.com/MimoJanra/confkit"
    "github.com/MimoJanra/confkit/etcd"
)

type Config struct {
    Database struct {
        Host     string `validate:"required"`
        Port     int    `default:"5432"`
        Name     string `validate:"required"`
        User     string `validate:"required"`
        Password string `secret:"true" validate:"required"`
    } `prefix:"DB_"`
}

func main() {
    cfg, err := confkit.Load[Config](
        confkit.FromEnv(),
        etcd.FromEtcd([]string{"localhost:2379"}),
    )
    if err != nil {
        log.Fatal(confkit.Explain(err))
    }
    
    log.Printf("Database: %s@%s:%d/%s", cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
}
```

## etcd Setup

### Start etcd

```bash
etcd --listen-client-urls=http://localhost:2379 --advertise-client-urls=http://localhost:2379
```

### Set Configuration

```bash
etcdctl put config/myapp/DB_HOST=postgres.internal
etcdctl put config/myapp/DB_PORT=5432
etcdctl put config/myapp/DB_NAME=myapp_db
etcdctl put config/myapp/DB_USER=appuser
etcdctl put config/myapp/DB_PASSWORD=secret
```

View stored values:

```bash
etcdctl get config/myapp/ --prefix
```

## Basic Usage

### Single Endpoint

```go
cfg, err := confkit.Load[Config](
    etcd.FromEtcd([]string{"localhost:2379"}),
)
```

### Multiple Endpoints

```go
cfg, err := confkit.Load[Config](
    etcd.FromEtcd([]string{
        "etcd-1.internal:2379",
        "etcd-2.internal:2379",
        "etcd-3.internal:2379",
    }),
)
```

### With Prefix

Load only keys with a specific prefix:

```go
cfg, err := confkit.Load[Config](
    etcd.FromEtcdWithPrefix([]string{"localhost:2379"}, "config/myapp/"),
)
```

## Kubernetes Integration

### Helm Install etcd

```bash
helm repo add bitnami https://charts.bitnami.com/bitnami
helm install etcd bitnami/etcd
```

### Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: etcd
spec:
  ports:
  - name: client
    port: 2379
  selector:
    app: etcd
```

### Application

```yaml
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
        env:
        - name: ETCD_ENDPOINTS
          value: "etcd:2379"
```

## Docker Compose

```yaml
version: '3.8'
services:
  etcd:
    image: quay.io/coreos/etcd:v3.5.0
    environment:
      - ETCD_LISTEN_CLIENT_URLS=http://0.0.0.0:2379
      - ETCD_ADVERTISE_CLIENT_URLS=http://etcd:2379
    ports:
      - "2379:2379"

  app:
    image: myapp:latest
    depends_on:
      - etcd
    environment:
      - ETCD_ENDPOINTS=etcd:2379
```

## Real-World Example

```go
package main

import (
    "log"
    "github.com/MimoJanra/confkit"
    "github.com/MimoJanra/confkit/etcd"
)

type Config struct {
    App struct {
        Name    string `default:"myapp"`
        Version string `validate:"required"`
    }
    Database struct {
        Host     string `validate:"required"`
        Port     int    `default:"5432"`
        Name     string `validate:"required"`
        User     string `validate:"required"`
        Password string `secret:"true" validate:"required"`
    } `prefix:"DB_"`
}

func main() {
    cfg, err := confkit.Load[Config](
        confkit.FromEnv(),
        etcd.FromEtcdWithPrefix(
            []string{"etcd.internal:2379"},
            "config/myapp/",
        ),
    )
    if err != nil {
        log.Fatal(confkit.Explain(err))
    }
    
    log.Printf("App: %s v%s", cfg.App.Name, cfg.App.Version)
    log.Printf("Connected to %s@%s", cfg.Database.User, cfg.Database.Host)
}
```

## Combining with Other Sources

The first source to provide a value wins; later sources fill in only unset fields:

```go
cfg, err := confkit.Load[Config](
    confkit.FromEnv(),                         // Highest priority — checked first
    etcd.FromEtcdWithPrefix(endpoints, prefix), // Shared config fills in what env did not set
    confkit.FromYAML("config.yaml"),           // Base fallback
)
```

## Updating etcd Keys

Update values dynamically:

```bash
# Update a single key
etcdctl put config/myapp/DB_HOST=new-host.internal

# Reload configuration
# (Your app picks up the change on next Load call)
```

## etcd with TLS

For production, enable TLS:

```bash
etcdctl \
  --endpoints=https://localhost:2379 \
  --cacert=/etc/etcd/ca.crt \
  --cert=/etc/etcd/client.crt \
  --key=/etc/etcd/client.key \
  put config/myapp/DB_HOST=...
```

In code:

```go
// Requires TLS configuration
cfg, err := confkit.Load[Config](
    etcd.FromEtcd([]string{"https://etcd.internal:2379"}),
)
```

## Watching for Changes

etcd supports watches for reactive updates:

```bash
# Watch for changes
etcdctl watch config/myapp/ --prefix
```

For automatic reloads in your application:

```go
// Reload config periodically
ticker := time.NewTicker(30 * time.Second)
defer ticker.Stop()

for range ticker.C {
    newCfg, err := confkit.Load[Config](
        etcd.FromEtcd([]string{"etcd:2379"}),
    )
    if err == nil {
        cfg = newCfg
    }
}
```

## Best Practices

1. **Use consistent key naming**
   ```bash
   etcdctl put config/myapp/DB_HOST=...
   etcdctl put config/myapp/DB_PASSWORD=...
   ```

2. **Use prefix for organization**
   ```go
   etcd.FromEtcdWithPrefix(endpoints, "config/myapp/")
   ```

3. **Mark secrets**
   ```go
   Password string `secret:"true"`
   ```

4. **Validate required values**
   ```go
   Password string `validate:"required" secret:"true"`
   ```

5. **Use multiple endpoints for HA**
   ```go
   etcd.FromEtcd([]string{
       "etcd-1:2379",
       "etcd-2:2379",
       "etcd-3:2379",
   })
   ```

## Troubleshooting

### Cannot connect to etcd

```
connection refused
```

Check etcd is running:

```bash
etcdctl member list
```

### Key not found

```
Key not found
```

Check key exists:

```bash
etcdctl get config/myapp/ --prefix
```

### Wrong prefix

Ensure prefix matches the keys you created:

```bash
# View all keys
etcdctl get --prefix ""

# Create with correct prefix
etcdctl put config/myapp/DB_HOST=...
```

## Use Cases

- **Kubernetes cluster config** — etcd already stores K8s state
- **Distributed config** — Share config across multiple services
- **Dynamic configuration** — Update config without restarting
- **Service discovery** — Store service locations and credentials

## See Also

- **[Consul](./consul.md)** — Consul KV alternative
- **[Vault](./vault.md)** — Secret management alternative
- **[Sources](../docs/sources.md)** — All configuration sources
