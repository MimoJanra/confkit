---
layout: default
title: Secret Redaction — confkit
---

# Secret Redaction

confkit automatically redacts sensitive fields in error messages, config dumps, and logs. Mark any field as `secret:"true"` and its value will never be exposed.

## Marking Secrets

Use the `secret:"true"` struct tag:

```go
type Config struct {
    Username string `env:"DB_USER"`
    Password string `env:"DB_PASSWORD" secret:"true"`
    
    APIKey   string `env:"API_KEY" secret:"true"`
    APIName  string `env:"API_NAME"`
}
```

## Redaction in Error Messages

When a secret field fails validation or parsing, confkit shows `<redacted>` instead of the value:

```go
type Config struct {
    DatabaseURL string `env:"DATABASE_URL" validate:"required" secret:"true"`
}

cfg, err := confkit.Load[Config](confkit.FromEnv())
if err != nil {
    fmt.Println(confkit.Explain(err))
}
```

**Output (if DATABASE_URL is invalid):**
```
Invalid configuration:

  DatabaseURL
    error: field is required
    source: env (DATABASE_URL)
    value: <redacted>  ← Never shows actual value
```

**Without `secret:"true"` (insecure):**
```
Invalid configuration:

  DatabaseURL
    error: field is required
    source: env (DATABASE_URL)
    value: postgres://user:pass@localhost/db  ← EXPOSED!
```

## Redaction in Config Dumps

Use `confkit.DumpConfig()` to safely log your loaded configuration:

```go
cfg, err := confkit.Load[Config](confkit.FromEnv())
fields := confkit.ScanFields(cfg)
dump, _ := confkit.DumpConfig(cfg, fields)

log.Println("Loaded config:", dump)
// Safe to log — secrets are redacted
```

**Output:**
```json
{
  "Username": "admin",
  "Password": "***REDACTED***",
  "APIKey": "***REDACTED***",
  "APIName": "my-api"
}
```

## Redaction in Programmatic Error Inspection

When inspecting errors programmatically, the `Value` field is empty for secrets:

```go
cfg, err := confkit.Load[Config](confkit.FromEnv())
if err != nil {
    report := err.(*confkit.ErrorReport)
    for _, fieldErr := range report.Errors {
        fmt.Printf("Field: %s\n", fieldErr.Path)
        fmt.Printf("Secret: %v\n", fieldErr.Secret)
        fmt.Printf("Value: %s\n", fieldErr.Value)  // empty string if Secret=true
    }
}
```

**Output:**
```
Field: Password
Secret: true
Value:               ← empty, never shown
```

## Nested Struct Secrets

Mark secret fields in nested structs:

```go
type Config struct {
    Server struct {
        Host string `env:"HOST"`
        Port int    `env:"PORT"`
    }
    Database struct {
        Host     string `env:"HOST" default:"localhost"`
        Port     int    `env:"PORT" default:"5432"`
        User     string `env:"USER"`
        Password string `env:"PASSWORD" secret:"true"`
    } `prefix:"DB_"`
}

// DB_PASSWORD is redacted everywhere
```

## Use Cases for Secret Redaction

### API Keys and Tokens

```go
type Config struct {
    APIKey    string `env:"API_KEY" secret:"true"`
    AuthToken string `env:"AUTH_TOKEN" secret:"true"`
}
```

### Database Credentials

```go
type Config struct {
    Database struct {
        URL      string `env:"URL" secret:"true"`
        Password string `env:"PASSWORD" secret:"true"`
    } `prefix:"DB_"`
}
```

### SSH Keys and Certificates

```go
type Config struct {
    SSH struct {
        PrivateKey string `env:"PRIVATE_KEY" secret:"true"`
        PublicKey  string `env:"PUBLIC_KEY"` // not secret
    }
}
```

### OAuth Secrets

```go
type Config struct {
    OAuth struct {
        ClientID     string `env:"CLIENT_ID"`
        ClientSecret string `env:"CLIENT_SECRET" secret:"true"`
    }
}
```

### AWS Credentials

```go
type Config struct {
    AWS struct {
        AccessKeyID     string `env:"ACCESS_KEY_ID" secret:"true"`
        SecretAccessKey string `env:"SECRET_ACCESS_KEY" secret:"true"`
        Region          string `env:"REGION"`
    }
}
```

## Combining with Validation

Secrets can have validation rules:

```go
type Config struct {
    DatabaseURL string `env:"DATABASE_URL" validate:"required" secret:"true"`
    APIKey      string `env:"API_KEY" validate:"min=32" secret:"true"`
}
```

Validation errors show `<redacted>` for secret fields.

## Combining with Defaults

Secrets can have defaults (though unusual):

```go
type Config struct {
    // Don't use default for secrets in production!
    // But if you must:
    APIKey string `env:"API_KEY" default:"dev-key-12345" secret:"true"`
}
```

Defaults are still redacted everywhere.

## Safe Logging Practices

### ✅ Safe: Use confkit.Explain()

```go
if err != nil {
    // Safe to log — secrets are redacted
    log.Error(confkit.Explain(err))
}
```

### ✅ Safe: Use confkit.DumpConfig()

```go
fields := confkit.ScanFields(cfg)
dump, _ := confkit.DumpConfig(cfg, fields)
log.Info("Config loaded", "config", dump)
```

### ❌ Unsafe: Direct struct logging

```go
log.Info("Config", "cfg", cfg)  // EXPOSES SECRETS!
```

### ❌ Unsafe: fmt.Sprintf

```go
log.Infof("Config: %v", cfg)  // EXPOSES SECRETS!
```

## Querying Secret Status

Check if a field is marked as secret:

```go
fields := confkit.ScanFields(cfg)
for _, field := range fields {
    if field.IsSecret {
        fmt.Printf("%s is marked as secret\n", field.Name)
    }
}
```

## Redaction Guarantees

confkit redacts secrets in:

✅ Error messages (via `confkit.Explain()`)  
✅ Config dumps (via `confkit.DumpConfig()`)  
✅ Programmatic error inspection (`ErrorReport.Errors[*].Value`)  
✅ Help text generation (schemas don't expose values)  

However, confkit **does not**:

❌ Redact directly from logs you write (use `DumpConfig()` first)  
❌ Redact from `fmt.Printf("%v", cfg)`  
❌ Manage environment variables in the OS  
❌ Protect values after they're loaded into your application  

**Your application must respect secret fields once loaded.**

## Best Practices

1. **Mark all sensitive fields**
   ```go
   Password string `env:"DB_PASSWORD" secret:"true"`
   APIKey   string `env:"API_KEY" secret:"true"`
   ```

2. **Use `confkit.DumpConfig()` when logging**
   ```go
   fields := confkit.ScanFields(cfg)
   dump, _ := confkit.DumpConfig(cfg, fields)
   log.Info("Loaded config", dump)
   ```

3. **Avoid logging raw config struct**
   ```go
   // ❌ Never
   log.Printf("Config: %+v", cfg)
   
   // ✅ Always
   fields := confkit.ScanFields(cfg)
   dump, _ := confkit.DumpConfig(cfg, fields)
   log.Printf("Config: %s", dump)
   ```

4. **Use validation with secrets**
   ```go
   Password string `env:"DB_PASSWORD" validate:"required" secret:"true"`
   ```

5. **Document which fields are secrets**
   ```go
   type Config struct {
       // Public fields
       Port int `env:"PORT" default:"8080"`
       
       // Secret fields — never log these
       APIKey   string `env:"API_KEY" secret:"true"`
       Password string `env:"DB_PASSWORD" secret:"true"`
   }
   ```

## Redaction Examples

### Example: Error with Secret

```go
type Config struct {
    DSN string `env:"DATABASE_URL" validate:"required" secret:"true"`
}

cfg, err := confkit.Load[Config](confkit.FromEnv())
if err != nil {
    fmt.Println(confkit.Explain(err))
}
```

**Output:**
```
Invalid configuration:

  DSN
    error: field is required
    source: env (DATABASE_URL)
```

The actual database URL is never shown, even though it's required.

### Example: Dump with Secrets

```go
type Config struct {
    Host     string `env:"HOST"`
    Port     int    `env:"PORT"`
    Username string `env:"USERNAME"`
    Password string `env:"PASSWORD" secret:"true"`
}

cfg, _ := confkit.Load[Config](confkit.FromEnv())

fields := confkit.ScanFields(cfg)
dump, _ := confkit.DumpConfig(cfg, fields)
log.Println(dump)
```

**Output:**
```json
{
  "Host": "db.example.com",
  "Port": 5432,
  "Username": "admin",
  "Password": "***REDACTED***"
}
```

Safe to log or send to external systems.

## Examples in Action

See secret redaction in all production examples:

- **[Web Service Example](https://github.com/MimoJanra/confkit/tree/main/examples)** — Database password marked as secret
- **[Microservice Example](https://github.com/MimoJanra/confkit/tree/main/examples)** — Multiple secrets (JWT, Redis, RabbitMQ passwords)
- **[Cloud-Native Example](https://github.com/MimoJanra/confkit/tree/main/examples)** — TLS certificates and credentials

All examples show secure handling of sensitive configuration data.

## Next Steps

- **[Validation](./validation.md)** — Validation rules
- **[Defaults](./defaults.md)** — Default values
- **[Examples](https://github.com/MimoJanra/confkit/tree/main/examples)** — Production examples with secrets
