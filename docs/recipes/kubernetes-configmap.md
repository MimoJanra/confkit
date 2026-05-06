---
layout: default
title: "Recipe: Kubernetes ConfigMap"
---

# Recipe: Kubernetes ConfigMap

Load configuration from a Kubernetes ConfigMap mounted as a volume.

## Use Case

- Kubernetes deployments
- ConfigMaps for non-secret configuration
- Development and staging environments
- Dynamic configuration updates

## Code

```go
package main

import (
    "log"
    "github.com/MimoJanra/confkit"
    "github.com/MimoJanra/confkit/k8s"
)

type Config struct {
    App struct {
        Name string `env:"NAME" default:"myapp"`
    }
    Server struct {
        Port int `env:"PORT" default:"8080"`
    }
    Database struct {
        Host     string `env:"HOST" default:"postgres"`
        Port     int    `env:"PORT" default:"5432"`
        Name     string `env:"NAME" validate:"required"`
        User     string `env:"USER" validate:"required"`
        Password string `env:"PASSWORD" validate:"required" secret:"true"`
    } `prefix:"DB_"`
}

func main() {
    cfg, err := confkit.Load[Config](
        confkit.FromEnv(),  // Highest priority — env vars override ConfigMap
        k8s.FromKubernetesConfigMap("default", "app-config"),
    )
    if err != nil {
        log.Fatal(confkit.Explain(err))
    }
    
    log.Printf("App: %s", cfg.App.Name)
    log.Printf("Server: :%d", cfg.Server.Port)
    log.Printf("Database: %s@%s:%d/%s", cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
}
```

## Kubernetes Manifests

### ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: default
data:
  APP_NAME: myapp
  SERVER_PORT: "8080"
  DB_HOST: postgres
  DB_PORT: "5432"
  DB_NAME: myapp_db
  DB_USER: postgres
```

### Secret (for sensitive data)

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: app-secrets
  namespace: default
type: Opaque
stringData:
  DB_PASSWORD: very-secret-password
```

### Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
  namespace: default
spec:
  replicas: 2
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
        volumeMounts:
        - name: config
          mountPath: /etc/app
        - name: secrets
          mountPath: /etc/app/secrets
      volumes:
      - name: config
        configMap:
          name: app-config
      - name: secrets
        secret:
          secretName: app-secrets
```

## Mounting as Files

If you prefer to load from files instead of environment variables:

```go
cfg, err := confkit.Load[Config](
    confkit.FromEnv(),
    confkit.FromYAML("/etc/app/config.yaml"),  // Mounted ConfigMap — fallback
)
```

### ConfigMap with File

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  config.yaml: |
    app:
      name: myapp
    server:
      port: 8080
    database:
      host: postgres
      port: 5432
      name: myapp_db
      user: postgres
```

### Mount File

```yaml
volumeMounts:
- name: config
  mountPath: /etc/app
volumes:
- name: config
  configMap:
    name: app-config
```

## Namespace Support

Load from different namespaces:

```go
// Current namespace
cfg, _ := confkit.Load[Config](
    k8s.FromKubernetesConfigMap("default", "app-config"),
)

// Specific namespace
cfg, _ := confkit.Load[Config](
    k8s.FromKubernetesConfigMap("production", "app-config"),
)
```

## ConfigMap with Custom Path

If the ConfigMap is mounted at a custom path:

```go
cfg, err := confkit.Load[Config](
    k8s.FromKubernetesConfigMapWithPath("default", "app-config", "/custom/path"),
)
```

## Multi-Environment Setup

### Development ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config-dev
  namespace: development
data:
  APP_NAME: myapp-dev
  DB_HOST: postgres-dev
  LOG_LEVEL: debug
```

### Production ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config-prod
  namespace: production
data:
  APP_NAME: myapp-prod
  DB_HOST: postgres-prod.internal
  LOG_LEVEL: info
```

### Deployment with ConfigMap Selection

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
  namespace: production
spec:
  template:
    spec:
      containers:
      - name: app
        envFrom:
        - configMapRef:
            name: app-config-prod  # Reference prod ConfigMap
```

## With Hot Reload

Restart the pod when ConfigMap changes:

```bash
# Force pod restart
kubectl rollout restart deployment/app -n default
```

Or use a sidecar to watch for ConfigMap changes.

## Combining Multiple ConfigMaps

```yaml
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: app
        envFrom:
        - configMapRef:
            name: app-config       # Base config
        - configMapRef:
            name: app-config-prod  # Environment-specific
        - secretRef:
            name: app-secrets      # Secrets
```

In code, combine sources. The first source to provide a value wins; later sources fill in only unset fields:

```go
cfg, err := confkit.Load[Config](
    confkit.FromEnv(),                                           // Highest priority — checked first
    k8s.FromKubernetesConfigMap("default", "app-config-prod"), // Environment-specific overrides
    k8s.FromKubernetesConfigMap("default", "app-config"),      // Base defaults
)
```

## RBAC Permissions

ConfigMap reader needs these permissions:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: app-reader
rules:
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: app-reader-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: app-reader
subjects:
- kind: ServiceAccount
  name: default
  namespace: default
```

## Complete Example

### ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: myapp-config
data:
  APP_NAME: "myapp"
  APP_ENV: "production"
  SERVER_PORT: "8080"
  DB_HOST: "postgres"
  DB_PORT: "5432"
  DB_NAME: "myapp_db"
  DB_USER: "appuser"
---
apiVersion: v1
kind: Secret
metadata:
  name: myapp-secrets
type: Opaque
stringData:
  DB_PASSWORD: "db-secret-password"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  replicas: 2
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
    spec:
      containers:
      - name: myapp
        image: myapp:1.0.0
        ports:
        - containerPort: 8080
        envFrom:
        - configMapRef:
            name: myapp-config
        - secretRef:
            name: myapp-secrets
```

### Go Code

```go
package main

import (
    "log"
    "github.com/MimoJanra/confkit"
    "github.com/MimoJanra/confkit/k8s"
)

type Config struct {
    App struct {
        Name string `env:"APP_NAME" default:"myapp"`
        Env  string `env:"APP_ENV" validate:"oneof=dev staging prod"`
    }
    Server struct {
        Port int `env:"SERVER_PORT" default:"8080"`
    }
    Database struct {
        Host     string `env:"DB_HOST" default:"localhost"`
        Port     int    `env:"DB_PORT" default:"5432"`
        Name     string `env:"DB_NAME" validate:"required"`
        User     string `env:"DB_USER" validate:"required"`
        Password string `env:"DB_PASSWORD" validate:"required" secret:"true"`
    } `prefix:""`
}

func main() {
    cfg, err := confkit.Load[Config](
        confkit.FromEnv(),
        k8s.FromKubernetesConfigMap("default", "myapp-config"),
    )
    if err != nil {
        log.Fatal(confkit.Explain(err))
    }
    
    log.Printf("App: %s (env=%s)", cfg.App.Name, cfg.App.Env)
    log.Printf("Listening on :%d", cfg.Server.Port)
}
```

## Best Practices

1. **Separate config and secrets**
   ```yaml
   # ConfigMap for non-sensitive
   kind: ConfigMap
   
   # Secret for sensitive data
   kind: Secret
   ```

2. **Use namespaces for isolation**
   ```bash
   kubectl create configmap app-config -n production
   kubectl create configmap app-config -n development
   ```

3. **Version your ConfigMaps**
   ```yaml
   name: app-config-v2
   ```

4. **Document required keys**
   ```yaml
   data:
     APP_NAME: "myapp"    # Required
     LOG_LEVEL: "info"    # Optional (has default)
   ```

5. **Use RBAC to restrict access**
   ```yaml
   kind: Role
   rules:
   - resources: ["configmaps"]
     verbs: ["get"]
   ```

## Troubleshooting

### ConfigMap not found

```
Error: configmap "app-config" not found in namespace "default"
```

Check the ConfigMap exists:

```bash
kubectl get configmaps -n default
kubectl describe configmap app-config -n default
```

### Permission denied

```
Error: configmaps is forbidden
```

Check RBAC permissions:

```bash
kubectl auth can-i get configmaps --as=system:serviceaccount:default:default
```

### Wrong namespace

Ensure ConfigMap is in the same namespace as the pod:

```bash
kubectl get configmap app-config -n production
# If not found, create it
kubectl create configmap app-config --from-literal=APP_NAME=myapp -n production
```

## See Also

- **[Getting Started](../docs/getting-started.md)** — Loading from environment
- **[Sources](../docs/sources.md)** — All configuration sources
- **[Recipes: YAML + Env](./yaml-env-overrides.md)** — Combining multiple sources
