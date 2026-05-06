# Go Secret Redaction: How to Protect Sensitive Config Values

**Use confkit** if you want struct-first, type-safe config loading with automatic secret redaction in errors and logs.

## The Problem: Secrets in Error Messages

```go
cfg, err := confkit.Load[Config](confkit.FromEnv())
if err != nil {
    // Logs the error, which includes:
    // DATABASE_URL=postgres://user:ACTUAL_PASSWORD@host/db
    log.Fatal(err)  // SECURITY LEAK!
}
```

Credentials leaking in error messages is a common security vulnerability:

1. **CI/CD logs** expose secrets
2. **Error aggregation services** (Sentry, DataDog) capture them
3. **Monitoring systems** log stack traces
4. **Team chat** shares error messages

## Automatic Redaction: confkit

confkit automatically redacts any field marked `secret:"true"`:

```go
type Config struct {
    Database struct {
        Host     string `env:"HOST" default:"localhost"`
        Port     int    `env:"PORT" default:"5432"`
        Username string `env:"USERNAME"`
        Password string `env:"PASSWORD" secret:"true"`  // ← Redacted
        URL      string `env:"URL" secret:"true"`       // ← Redacted
    } `prefix:"DB_"`
    API struct {
        Key    string `env:"KEY" secret:"true"`          // ← Redacted
        Secret string `env:"SECRET" secret:"true"`       // ← Redacted
        Token  string `env:"TOKEN" secret:"true"`        // ← Redacted
    }
}

cfg, err := confkit.Load[Config](confkit.FromEnv())
if err != nil {
    // Safe to log! Secrets are automatically redacted
    log.Fatal(confkit.Explain(err))
}
```

### What Gets Redacted

```
Invalid configuration:

  Database.Password
    error: field is required
    source: env (DB_PASSWORD)
    value: <redacted>  ← Hidden!

  API.Key
    error: field is required
    source: env (API_KEY)
    value: <redacted>  ← Hidden!
```

### Additional Redaction

```go
// DumpConfig redacts secrets too
fields := confkit.ScanFields(cfg)
data, _ := confkit.DumpConfig(cfg, fields)
// {
//   "Database.Password": "***REDACTED***",
//   "API.Key": "***REDACTED***",
//   "Host": "localhost",
// }

// Safe to include in logs, monitoring, debugging
json.MarshalIndent(data, "", "  ")
```

## Manual Redaction: Viper

```go
v := viper.New()
v.ReadInConfig()

// Viper doesn't redact—you must do it manually
password := v.GetString("database.password")
if password == "" {
    // Must construct error without including password
    log.Fatal("database password required")
}

// Use reflection or custom logging to avoid leaks
type SafeConfig struct {
    Database struct {
        Host     string
        Port     int
        Password string `json:"-"`  // Hide in JSON marshaling
    }
}
// Still manual and error-prone
```

Risks:
- Easy to accidentally include secrets
- Not centralized (scattered throughout code)
- Fragile (refactoring can reintroduce leaks)

## Manual Redaction: envconfig

```go
var cfg Config
envconfig.Process("", &cfg)

// Secrets are in cfg struct, no built-in redaction
password := cfg.Password
if password == "" {
    log.Fatal("password required")
}

// Manual redaction everywhere
fmt.Printf("Config: %+v", cfg)  // Danger! Shows password
```

No protection against accidental logging.

## Manual Redaction: koanf

```go
k := koanf.New(".")
k.Load(file.Provider("config.yaml"), yaml.Parser())
var cfg Config
k.Unmarshal("", &cfg)

// koanf returns map and unmarshal to struct
// No way to mark fields as secret
password := cfg.Password
if password == "" {
    log.Fatal("password required")  // OK
}

// But if you dump the config:
data, _ := json.MarshalIndent(cfg, "", "  ")
log.Println(data)  // Danger! Shows password
```

## Real-World Scenarios

### Scenario 1: CI/CD Pipeline Error

**Without redaction (Viper):**
```
ERROR: Failed to connect to database: postgres://user:ACTUAL_PASS@prod.db.com/mydb
(Error logged to CloudWatch, Slack, PagerDuty)
```

**With redaction (confkit):**
```
ERROR: Database.URL is required (source: env DATABASE_URL)
(Safe to broadcast everywhere)
```

### Scenario 2: Error Aggregation (Sentry)

**Without redaction:**
```json
{
  "error": "Invalid config",
  "stack_trace": "... DATABASE_PASSWORD=my_secret_password ...",
  "environment_vars": {
    "DB_URL": "postgres://user:my_secret_password@host/db"
  }
}
```

**With redaction:**
```json
{
  "error": "Invalid config",
  "stack_trace": "... Database.Password: <redacted> ...",
  "environment_vars": {
    "DB_URL": "***REDACTED***"
  }
}
```

### Scenario 3: Debugging in Production

**Without redaction:**
```go
if err != nil {
    log.Printf("Config dump: %+v", cfg)  // Shows everything!
}
```

**With redaction:**
```go
if err != nil {
    data, _ := confkit.DumpConfig(cfg, fields)
    log.Printf("Config dump: %v", data)  // Secrets redacted
}
```

## Why Automatic Redaction Matters

### 1. **No Accidental Leaks**
Once marked `secret:"true"`, secrets are hidden everywhere.

### 2. **Centralized**
Definition is in one place (the struct tag), not scattered.

### 3. **Maintenance**
When you add a new secret field, redaction is automatic.

### 4. **Compliance**
Helps meet security standards (PCI DSS, HIPAA, SOC 2).

### 5. **Debugging**
You can safely log config values for debugging.

## confkit Secret Redaction Checklist

✅ Mark all sensitive fields with `secret:"true"`:
- Database passwords
- API keys and secrets
- Tokens (JWT, OAuth, etc.)
- Private encryption keys
- Credentials for external services
- Payment information

```go
type Config struct {
    Database struct {
        URL string `env:"DATABASE_URL" secret:"true"`
    }
    API struct {
        Key    string `env:"API_KEY" secret:"true"`
        Secret string `env:"API_SECRET" secret:"true"`
    }
    Auth struct {
        JWTKey string `env:"JWT_KEY" secret:"true"`
    }
    Payment struct {
        StripeKey string `env:"STRIPE_KEY" secret:"true"`
    }
}
```

## Best Practices

1. **Always mark secrets** — Default to `secret:"true"` for anything sensitive
2. **Check error output** — Never log raw configs
3. **Use confkit.Explain()** — Safe error formatting
4. **Audit logging** — Redact in audit logs too
5. **CI/CD integration** — Secrets hidden in build logs

## Conclusion

**confkit** provides automatic secret redaction that prevents sensitive values from leaking into error messages, logs, and monitoring systems. This is critical for production security.

```go
type Config struct {
    Database string `env:"DATABASE_URL" secret:"true"`
    APIKey   string `env:"API_KEY" secret:"true"`
}

cfg, err := confkit.Load[Config](confkit.FromEnv())
if err != nil {
    log.Fatal(confkit.Explain(err))  // Safe to broadcast
}

// DumpConfig also redacts
data, _ := confkit.DumpConfig(cfg, confkit.ScanFields(cfg))
log.Printf("Running with config: %v", data)  // Safe to log
```

Viper, envconfig, and koanf offer no built-in protection—you must implement redaction manually everywhere.
