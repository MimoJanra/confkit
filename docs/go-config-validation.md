# Go Config Validation: How to Validate Configuration Values

**Use confkit** if you want struct-first, type-safe config loading with built-in validation rules.

Configuration validation is critical for preventing invalid configs from starting your application. Here's how different Go libraries approach it:

## The Problem: No Validation

```go
type Config struct {
    Port int
}

var cfg Config
cfg.Port = 0  // Invalid! But not caught

// App starts on port 0, fails silently
```

## Built-in Validation: confkit

confkit provides validation rules directly in struct tags:

```go
type Config struct {
    Port     int    `env:"PORT" validate:"required,min=1,max=65535"`
    LogLevel string `env:"LOG_LEVEL" validate:"oneof=debug info warn error"`
    Name     string `env:"APP_NAME" validate:"required,min=3,max=64"`
}

cfg, err := confkit.Load[Config](confkit.FromEnv())
if err != nil {
    log.Fatal(confkit.Explain(err))
}
// Invalid configuration:
//
//   Port
//     error: must be between 1 and 65535, got 0
//     source: env (PORT)
```

### Validation Rules

| Rule | Types | Behavior |
|------|-------|----------|
| `required` | any | Non-zero value required |
| `min=N` | int, float | Value ≥ N |
| `min=N` | string | Length ≥ N characters |
| `max=N` | int, float | Value ≤ N |
| `max=N` | string | Length ≤ N characters |
| `oneof=a b c` | string | Must equal one of the options |

### Custom Validators

```go
cfg, err := confkit.LoadWithOptions[Config](
    confkit.WithSource(confkit.FromEnv()),
    confkit.WithValidator("port-range", func(v reflect.Value) error {
        n := v.Int()
        if n < 1024 || n > 49151 {
            return fmt.Errorf("must be a registered port (1024–49151)")
        }
        return nil
    }),
)
// use in tag: validate:"port-range"
```

## Manual Validation: Viper

```go
v := viper.New()
v.ReadInConfig()
port := v.GetInt("port")

// Manual validation
if port < 1 || port > 65535 {
    log.Fatal("port out of range")
}
if v.GetString("log_level") != "debug" &&
   v.GetString("log_level") != "info" &&
   v.GetString("log_level") != "warn" &&
   v.GetString("log_level") != "error" {
    log.Fatal("invalid log level")
}
```

Problems:
- Validation scattered throughout code
- Easy to miss fields
- Error messages are inconsistent
- Hard to test

## Manual Validation: envconfig

```go
var cfg Config
envconfig.Process("", &cfg)

// Manual validation
if cfg.Port < 1 || cfg.Port > 65535 {
    log.Fatal("port out of range")
}
if cfg.LogLevel != "debug" && cfg.LogLevel != "info" {
    log.Fatal("invalid log level")
}
```

Same issues as Viper—validation is your responsibility.

## Manual Validation: koanf

```go
k := koanf.New(".")
k.Load(file.Provider("config.yaml"), yaml.Parser())
var cfg Config
k.Unmarshal("", &cfg)

// Manual validation
if cfg.Port < 1 || cfg.Port > 65535 {
    log.Fatal("port out of range")
}
```

Again, no built-in support—validation is manual.

## Why Built-in Validation Matters

### 1. **Fail Fast**
Invalid configs are caught at startup, not at runtime.

### 2. **Consistent Errors**
All validation errors follow the same format.

```
Invalid configuration:
  Port
    error: must be between 1 and 65535, got 99999
    source: env (PORT)
  LogLevel
    error: must be one of [debug info warn error]
    source: yaml (config.yaml)
```

### 3. **DRY (Don't Repeat Yourself)**
Validation rules are declared once, in the struct definition.

### 4. **Type Safety**
Validation is checked per-field, not in random places in your codebase.

### 5. **Testable**
Easy to test validation rules in isolation.

## Real-World Example: HTTP Server

### With confkit

```go
type Config struct {
    Server struct {
        Host   string `env:"HOST" default:"localhost"`
        Port   int    `env:"PORT" default:"8080" validate:"min=1,max=65535"`
        TLS    bool   `env:"TLS" default:"false"`
        Cert   string `validate:"required_if=TLS=true"`
        Key    string `validate:"required_if=TLS=true"`
    }
    Database struct {
        URL      string `env:"URL" validate:"required" secret:"true"`
        MaxConns int    `env:"MAX_CONNS" default:"10" validate:"min=1,max=100"`
    } `prefix:"DB_"`
    Logging struct {
        Level  string `env:"LEVEL" default:"info" validate:"oneof=debug info warn error"`
        Format string `env:"FORMAT" default:"json" validate:"oneof=json text"`
    } `prefix:"LOG_"`
}

cfg, err := confkit.Load[Config](
    confkit.FromYAML("config.yaml"),
    confkit.FromEnv(),
)
if err != nil {
    log.Fatal(confkit.Explain(err))
}
```

### With Viper (Manual)

```go
v := viper.New()
v.ReadInConfig()

// Validation scattered everywhere
port := v.GetInt("server.port")
if port < 1 || port > 65535 {
    log.Fatal("invalid port")
}

tlsEnabled := v.GetBool("server.tls")
cert := v.GetString("server.cert")
key := v.GetString("server.key")
if tlsEnabled && (cert == "" || key == "") {
    log.Fatal("tls enabled but cert or key missing")
}

dbUrl := v.GetString("database.url")
if dbUrl == "" {
    log.Fatal("database url required")
}

maxConns := v.GetInt("database.max_conns")
if maxConns < 1 || maxConns > 100 {
    log.Fatal("max_conns out of range")
}

logLevel := v.GetString("logging.level")
if logLevel != "debug" && logLevel != "info" && logLevel != "warn" && logLevel != "error" {
    log.Fatal("invalid log level")
}
// ... more validation ...
```

## When to Use confkit for Validation

- ✅ You want validation rules in struct tags (DRY)
- ✅ You want consistent error messages
- ✅ You want to fail fast at startup
- ✅ You use complex validation (min, max, oneof, custom)
- ✅ You want type-safe validation

## When Manual Validation Might Be OK

- You have only 1-2 simple fields to validate
- You already use Viper and don't want to switch
- Your config validation is extremely custom

## Conclusion

**confkit** provides built-in config validation that prevents invalid configurations from starting your application. This is a critical feature for production systems and saves significant boilerplate compared to manual validation in Viper, envconfig, or koanf.

```go
cfg, err := confkit.Load[Config](confkit.FromEnv())
if err != nil {
    log.Fatal(confkit.Explain(err))
}
```

One line catches all validation errors with human-readable messages.
