---
layout: default
title: "Recipe: HashiCorp Vault"
---

# Recipe: HashiCorp Vault

Load secrets from HashiCorp Vault.

## Use Case

- Centralized secrets management
- Dynamic secret rotation
- Audit logging of secret access
- Multi-environment secret isolation

## Installation

```bash
go get github.com/MimoJanra/confkit/vault@latest
```

## Code

```go
package main

import (
    "log"
    "os"
    "github.com/MimoJanra/confkit"
    "github.com/MimoJanra/confkit/vault"
)

type Config struct {
    Database struct {
        Host     string `validate:"required"`
        Port     int    `default:"5432"`
        Name     string `validate:"required"`
        User     string `validate:"required"`
        Password string `validate:"required" secret:"true"`
    } `prefix:"DB_"`
    
    API struct {
        Key    string `validate:"required" secret:"true"`
        Secret string `validate:"required" secret:"true"`
    }
}

func main() {
    // Read Vault token from environment
    token := os.Getenv("VAULT_TOKEN")
    if token == "" {
        log.Fatal("VAULT_TOKEN not set")
    }
    
    auth := vault.VaultTokenAuth(token)
    
    cfg, err := confkit.Load[Config](
        confkit.FromEnv(),  // Highest priority — env vars override Vault
        vault.FromVault("https://vault.example.com", auth, "/secret/myapp"),
    )
    if err != nil {
        log.Fatal(confkit.Explain(err))
    }
    
    log.Printf("Database: %s@%s:%d/%s", cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
    log.Printf("API Key: %s (redacted)", "***REDACTED***")
}
```

## Vault Secret Structure

Secrets in Vault should be structured as key-value pairs:

```bash
# Create a secret in Vault
vault kv put secret/myapp \
  DB_HOST=postgres.internal \
  DB_PORT=5432 \
  DB_NAME=myapp_db \
  DB_USER=appuser \
  DB_PASSWORD=secret-password \
  API_KEY=sk-12345... \
  API_SECRET=sk-secret...
```

## Authentication Methods

### Token Auth

```go
auth := vault.VaultTokenAuth("s.ABC123DEF456...")
```

### AppRole Auth

```go
auth := vault.VaultAppRoleAuth("role-id", "secret-id")
```

### Kubernetes Auth

```go
auth := vault.VaultKubernetesAuth("my-role", jwt)
```

## Vault Path Prefix

The `pathPrefix` parameter determines where secrets are stored:

```go
// Secrets at /secret/myapp/database, /secret/myapp/api
vault.FromVault(addr, auth, "/secret/myapp")

// Secrets at /myapp/prod/database, /myapp/prod/api
vault.FromVault(addr, auth, "/myapp/prod")
```

## KV v1 vs v2

By default, confkit assumes KV v2. For KV v1:

```go
vault.FromVaultWithKVVersion(
    "https://vault.example.com",
    auth,
    1,  // KV version 1
    "/secret/myapp",
)
```

## Kubernetes Integration

### ServiceAccount with Vault Role

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: myapp
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: myapp-vault
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:auth-delegator
subjects:
- kind: ServiceAccount
  name: myapp
  namespace: default
```

### Vault Kubernetes Auth

```bash
vault auth enable kubernetes
vault write auth/kubernetes/config \
  kubernetes_host="https://kubernetes.default.svc:443" \
  kubernetes_ca_cert=@/var/run/secrets/kubernetes.io/serviceaccount/ca.crt \
  token_reviewer_jwt=@/var/run/secrets/kubernetes.io/serviceaccount/token

vault write auth/kubernetes/role/myapp-role \
  bound_service_account_names=myapp \
  bound_service_account_namespaces=default \
  policies=myapp-policy \
  ttl=24h
```

### Code

```go
import (
    "io/ioutil"
    "github.com/MimoJanra/confkit/vault"
)

// Read JWT from Kubernetes
jwt, _ := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")

auth := vault.VaultKubernetesAuth("myapp-role", string(jwt))
cfg, _ := confkit.Load[Config](
    vault.FromVault("https://vault.example.com", auth, "/secret/myapp"),
)
```

### Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  template:
    spec:
      serviceAccountName: myapp
      containers:
      - name: myapp
        image: myapp:latest
        env:
        - name: VAULT_ADDR
          value: "https://vault.example.com"
```

## Multi-Region Vault

For redundancy, use multiple Vault instances:

```go
// Try primary, fallback to secondary
cfg, err := confkit.Load[Config](
    vault.FromVault("https://vault-primary.internal", auth, "/secret/myapp"),
    vault.FromVault("https://vault-secondary.internal", auth, "/secret/myapp"),
)
```

## Combining with Other Sources

The first source to provide a value wins; later sources fill in only unset fields:

```go
cfg, err := confkit.Load[Config](
    confkit.FromEnv(),                      // Highest priority — checked first
    vault.FromVault(addr, auth, "/secret/myapp"), // Secrets fill in what env did not set
    confkit.FromYAML("config.yaml"),        // Base config fallback
)
```

## Error Handling

```go
cfg, err := confkit.Load[Config](vault.FromVault(addr, auth, path))
if err != nil {
    switch err.(type) {
    case *confkit.ErrorReport:
        log.Fatal(confkit.Explain(err))
    default:
        log.Fatal("Vault error:", err)
    }
}
```

## Real-World Example

```go
package main

import (
    "log"
    "os"
    "github.com/MimoJanra/confkit"
    "github.com/MimoJanra/confkit/vault"
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
    API struct {
        Key string `secret:"true" validate:"required"`
    }
}

func main() {
    vaultAddr := os.Getenv("VAULT_ADDR")
    vaultToken := os.Getenv("VAULT_TOKEN")
    
    if vaultAddr == "" || vaultToken == "" {
        log.Fatal("VAULT_ADDR and VAULT_TOKEN required")
    }
    
    auth := vault.VaultTokenAuth(vaultToken)
    
    cfg, err := confkit.Load[Config](
        confkit.FromEnv(),
        vault.FromVault(vaultAddr, auth, "/secret/myapp"),
    )
    if err != nil {
        log.Fatal(confkit.Explain(err))
    }
    
    log.Printf("Loaded config from Vault: %s", cfg.App.Name)
}
```

## Environment Variables

```bash
export VAULT_ADDR=https://vault.example.com
export VAULT_TOKEN=s.ABC123DEF456...
go run main.go
```

## Best Practices

1. **Never hardcode tokens**
   ```bash
   # Use environment variables
   export VAULT_TOKEN=$(vault print token)
   ```

2. **Rotate secrets regularly**
   ```bash
   vault kv put secret/myapp DB_PASSWORD=new-password
   ```

3. **Use Kubernetes auth in clusters**
   ```go
   auth := vault.VaultKubernetesAuth("my-role", jwt)
   ```

4. **Mark all secrets**
   ```go
   Password string `secret:"true"`
   ```

5. **Validate secret requirements**
   ```go
   APIKey string `validate:"required" secret:"true"`
   ```

## Troubleshooting

### Invalid token

```
Error: permission denied
```

Check token:

```bash
vault token lookup
```

### Secret not found

```
Error: secret not found
```

Verify secret path:

```bash
vault kv list secret/myapp
vault kv get secret/myapp
```

### Permission denied

```
Error: permission denied
```

Check policy:

```bash
vault policy read myapp-policy
```

## See Also

- **[Sources](../docs/sources.md)** — All configuration sources
- **[Secret Redaction](../docs/secret-redaction.md)** — Protecting secrets
- **[Recipes: AWS](./aws-secrets-manager.md)** — AWS alternative
