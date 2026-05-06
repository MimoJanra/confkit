---
layout: default
title: "Recipe: Consul KV"
---

# Recipe: Consul KV

Load configuration from HashiCorp Consul KV store.

## Installation

```bash
go get github.com/MimoJanra/confkit/consul
```

## Code

```go
package main

import (
    "log"
    "github.com/MimoJanra/confkit"
    "github.com/MimoJanra/confkit/consul"
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
        consul.FromConsul("localhost:8500"),
        confkit.FromEnv(),
    )
    if err != nil {
        log.Fatal(confkit.Explain(err))
    }
    
    log.Printf("Database: %s@%s:%d/%s", cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
}
```

## Consul Setup

### Start Consul

```bash
consul agent -server -ui -bootstrap-expect=1 -data-dir=/tmp/consul
```

### Add Configuration

```bash
consul kv put config/myapp/DB_HOST=postgres.internal
consul kv put config/myapp/DB_PORT=5432
consul kv put config/myapp/DB_NAME=myapp_db
consul kv put config/myapp/DB_USER=appuser
consul kv put config/myapp/DB_PASSWORD=secret
```

Or use JSON:

```bash
consul kv put config/myapp '{
  "DB_HOST": "postgres.internal",
  "DB_PORT": "5432",
  "DB_NAME": "myapp_db",
  "DB_USER": "appuser",
  "DB_PASSWORD": "secret"
}'
```

## Basic Usage

### Default Address

```go
cfg, err := confkit.Load[Config](
    consul.FromConsul("localhost:8500"),
)
```

### With Token

```go
cfg, err := confkit.Load[Config](
    consul.FromConsulWithToken("consul.internal:8500", "token-123..."),
)
```

### With Datacenter

```go
cfg, err := confkit.Load[Config](
    consul.FromConsulWithOptions("consul.internal:8500", "token-123...", "us-west-1"),
)
```

## Key Naming

Keys in Consul follow the path structure:

```
config/myapp/DB_HOST
config/myapp/DB_PORT
config/myapp/DB_NAME
```

confkit loads all keys with the prefix and uses them as environment variables:

```
DB_HOST → postgres.internal
DB_PORT → 5432
DB_NAME → myapp_db
```

## Consul with ACL

```bash
# Create policy
consul acl policy create -name app-reader -rules '
node_prefix "" {
  policy = "read"
}
key_prefix "config/myapp/" {
  policy = "read"
}
'

# Create token
consul acl token create -policy-name app-reader
```

Use token in code:

```go
cfg, err := confkit.Load[Config](
    consul.FromConsulWithToken("consul:8500", "token-from-acl"),
)
```

## Docker Integration

```yaml
version: '3.8'
services:
  consul:
    image: consul:latest
    ports:
      - "8500:8500"
    command: agent -server -ui -bootstrap-expect=1 -client=0.0.0.0

  app:
    image: myapp:latest
    depends_on:
      - consul
    environment:
      - CONFIG_CONSUL_ADDR=consul:8500
```

## Kubernetes Integration

### Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: consul
spec:
  ports:
  - port: 8500
  selector:
    app: consul
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: consul
spec:
  selector:
    matchLabels:
      app: consul
  template:
    metadata:
      labels:
        app: consul
    spec:
      containers:
      - name: consul
        image: consul:latest
        ports:
        - containerPort: 8500
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
        - name: CONSUL_ADDR
          value: "consul:8500"
```

## Real-World Example

```go
package main

import (
    "log"
    "github.com/MimoJanra/confkit"
    "github.com/MimoJanra/confkit/consul"
)

type Config struct {
    App struct {
        Name string `default:"myapp"`
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
        consul.FromConsul("consul.internal:8500"),
        confkit.FromEnv(),
    )
    if err != nil {
        log.Fatal(confkit.Explain(err))
    }
    
    log.Printf("Loaded config from Consul: %s", cfg.App.Name)
}
```

## Combining with Other Sources

```go
cfg, err := confkit.Load[Config](
    confkit.FromYAML("config.yaml"),        // Base
    consul.FromConsul("consul:8500"),       // Consul KV
    confkit.FromEnv(),                      // Overrides
)
```

## Updating Consul KV

Update values in Consul:

```bash
consul kv put config/myapp/DB_HOST=new-host.internal
```

Your application picks up the change on next reload:

```go
// Reload config
newCfg, err := confkit.Load[Config](
    consul.FromConsul("consul:8500"),
)
```

## Consul Watches

For automatic reloads, use Consul watches:

```bash
consul watch -type=key -key=config/myapp/ curl -X POST http://localhost:8080/reload
```

## Best Practices

1. **Use descriptive key names**
   ```bash
   consul kv put config/myapp/DB_HOST=...
   consul kv put config/myapp/DB_PASSWORD=...
   ```

2. **Enable ACL**
   ```go
   consul.FromConsulWithToken(addr, token)
   ```

3. **Mark secrets**
   ```go
   Password string `secret:"true"`
   ```

4. **Validate required values**
   ```go
   Password string `validate:"required" secret:"true"`
   ```

5. **Use Consul for shared config**
   - Shared database settings
   - Feature flags
   - Non-sensitive settings

## Troubleshooting

### Cannot connect to Consul

```
Connection refused
```

Check Consul is running and accessible:

```bash
consul version
consul members
```

### Key not found

```
Key not found: config/myapp/DB_HOST
```

Check keys exist:

```bash
consul kv get config/myapp/
consul kv get config/myapp/DB_HOST
```

### ACL permission denied

```
ACL not found
```

Check token is valid:

```bash
consul acl token read -id=token-123...
```

## See Also

- **[Vault](./vault.md)** — HashiCorp Vault alternative
- **[etcd](./etcd.md)** — etcd alternative
- **[Sources](../docs/sources.md)** — All configuration sources
