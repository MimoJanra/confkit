
# confkit vs koanf

confkit is a type-safe configuration toolkit for Go projects that want struct-based config loading, defaults, validation, source merging, and secret redaction. koanf is a configuration library focused on extreme modularity and support for many file formats.

## Key Differences

### Architecture

**confkit:** Struct-first — your config is a Go struct, everything flows from there.

```go
type Config struct {
    Port int    `env:"PORT" default:"8080" validate:"min=1,max=65535"`
    DSN  string `env:"DATABASE_URL" secret:"true"`
}

cfg, err := confkit.Load[Config](
    confkit.FromYAML("config.yaml"),
    confkit.FromEnv(),
)
// cfg is exactly type Config
```

**koanf:** Map-first — your config is a map, you unmarshal to struct afterward.

```go
k := koanf.New(".")
k.Load(file.Provider("config.yaml"), yaml.Parser())
k.Load(env.Provider("APP_", ".", nil), nil)

var cfg Config
k.Unmarshal("", &cfg)
// cfg is populated from the merged map
```

### Type Safety

**confkit:** Full compile-time type checking.

```go
cfg.Port  // type int, no casting
cfg.DSN   // type string, no casting
```

**koanf:** Runtime unmarshaling into a struct; loose initially, then typed.

```go
var port interface{} = k.Get("port")  // interface{}, could be anything
portInt := port.(int)  // explicit cast, can panic
```

### Built-in Defaults

**confkit:** Via struct tags.

```go
type Config struct {
    Port int `env:"PORT" default:"8080"`
}
// Defaults are built-in, no extra setup
```

**koanf:** Manual setup on the map.

```go
k.Set("port", 8080)
// Defaults must be set explicitly on the map
```

### Built-in Validation

**confkit:** Via struct tags with multiple rules.

```go
type Config struct {
    Port int `env:"PORT" validate:"min=1,max=65535"`
}
```

**koanf:** No built-in validation — you validate after unmarshaling.

```go
var cfg Config
k.Unmarshal("", &cfg)
// Validation is your responsibility
if cfg.Port < 1 || cfg.Port > 65535 {
    return fmt.Errorf("port out of range")
}
```

### Error Messages

**confkit:** Structured, human-readable errors with field context and source information.

```
Invalid configuration:

  Port
    error: must be between 1 and 65535, got 99999
    source: env (PORT)
```

**koanf:** Generic unmarshaling errors from encoding/json or similar.

```
json: cannot unmarshal string into Go struct field Config.Port of type int
```

### Secret Redaction

**confkit:** Automatic via `secret:"true"` tag.

```go
type Config struct {
    Token string `env:"API_TOKEN" secret:"true"`
}
// Errors and dumps automatically redact this field
```

**koanf:** No built-in secret redaction.

```go
// You must redact secrets manually when dumping or logging
```

### File Format Support

**confkit:** YAML, JSON, TOML, env, flags, cloud sources.

**koanf:** YAML, JSON, TOML, HCL, INI, JSONNET, and more via providers.

koanf's modularity means it can support many more formats, but confkit covers the most common ones.

### Configuration Composition

**confkit:** Explicit source precedence — last one wins per field.

```go
cfg, err := confkit.Load[Config](
    confkit.FromYAML("defaults.yaml"),    // base
    confkit.FromYAML("config.yaml"),      // overrides
    confkit.FromEnv(),                    // highest priority
)
```

**koanf:** Progressive loading with explicit merge order.

```go
k := koanf.New(".")
k.Load(file.Provider("defaults.yaml"), yaml.Parser())
k.Load(file.Provider("config.yaml"), yaml.Parser())
k.Load(env.Provider("APP_", ".", nil), nil)
```

Both are explicit, but koanf loads into a map first, then unmarshals. confkit loads directly into the struct.

### Cloud Sources

**confkit:** Optional modules for Vault, Consul, etcd, AWS.

```bash
go get github.com/MimoJanra/confkit/vault
```

**koanf:** Optional provider modules for Vault, Consul, etcd, AWS (Parameter Store, AppConfig), and S3.

```bash
go get github.com/knadh/koanf/v2/providers/vault/v2  # for Vault
go get github.com/knadh/koanf/v2/providers/consul/v2  # for Consul
go get github.com/knadh/koanf/v2/providers/aws  # for AWS services
```

### Nested Struct Prefixes

**confkit:** Via struct tags with unlimited nesting.

```go
type Config struct {
    DB struct {
        Host string `env:"HOST"`
    } `prefix:"DB_"`
}
// Reads DB_HOST
```

**koanf:** Via dotted keys in the map; nesting works but less intuitive for structs.

```go
k.Get("db.host")  // map-based access
// Struct unmarsal handles this but config definition is loose
```

## When to Use confkit

- You want a single struct to define your entire config
- You need defaults and validation built-in
- You care about type safety and compile-time checking
- You want human-readable error messages
- You use cloud sources (Vault, AWS, Kubernetes)
- You need secret redaction
- Your config structure is well-defined and stable

## When to Use koanf

- You need support for many file formats (HCL, INI, JSONNET, etc.)
- Your config is more dynamic (fields come and go)
- You prefer a map-based approach with loose typing
- You want extreme modularity and custom providers
- You don't need validation or defaults built-in
- You're building a tool that accepts arbitrary config formats

## Feature Comparison Table

|                          | confkit | koanf |
|--------------------------|:-------:|:-----:|
| **Typed `Load[T]`**       |    ✅    |   ❌   |
| **Defaults via tags**     |    ✅    |   ❌   |
| **Built-in validation**   |    ✅    |   ❌   |
| **Secret redaction**      |    ✅    |   ❌   |
| **Multi-source merging**  |    ✅    |   ✅   |
| **Map-based access**      |    ❌    |   ✅   |
| **Many file formats**     |    ⚠️    |   ✅   |
| **Custom providers**      |  limited |   ✅   |
| **Schema generation**     |    ✅    |   ❌   |
| **Error context**         |    ✅    |   ⚠️   |
| **Cloud integrations**    |  optional | optional |

## Migration Path

If you're moving from koanf to confkit:

1. Define a struct instead of working with maps:
   ```go
   type Config struct {
       Port int    `env:"PORT" default:"8080"`
       DSN  string `env:"DATABASE_URL" secret:"true"`
   }
   ```

2. Replace koanf's Load/Unmarshal with `confkit.Load[Config]`:
   ```go
   // Before: k.Load(...); cfg := &Config{}; k.Unmarshal("", cfg)
   // After:
   cfg, err := confkit.Load[Config](
       confkit.FromYAML("config.yaml"),
       confkit.FromEnv(),
   )
   ```

3. Add validation rules inline:
   ```go
   Port int `env:"PORT" validate:"min=1,max=65535"`
   ```

4. Add defaults inline:
   ```go
   Port int `env:"PORT" default:"8080"`
   ```

5. Mark secrets for redaction:
   ```go
   DSN string `env:"DATABASE_URL" secret:"true"`
   ```

## Code Complexity

**confkit:** Minimal. Define struct, call `Load[T]()`, done.

```go
cfg, err := confkit.Load[Config](confkit.FromYAML("config.yaml"), confkit.FromEnv())
```

**koanf:** More boilerplate. Create koanf instance, load sources, unmarshal.

```go
k := koanf.New(".")
k.Load(file.Provider("config.yaml"), yaml.Parser())
k.Load(env.Provider("APP_", ".", nil), nil)
var cfg Config
k.Unmarshal("", &cfg)
```

## Library Size & Dependencies

| Library | Dependencies | Binary overhead |
|---------|:------------:|:---------------:|
| confkit | yaml.v3, go-toml/v2 | ~500KB (core) |
| koanf   | yaml.v3, toml, etc. | ~600KB+ |

Both are similar; koanf may be larger if you use many providers.

## Use Case Example

**Microservice with typed config, validation, defaults, and cloud sources:** confkit.

```go
cfg, err := confkit.Load[Config](
    confkit.FromYAML("config.yaml"),
    confkit.FromEnv(),
    vault.FromVault(...),
)
```

**Tool that accepts many config formats and works with dynamic keys:** koanf.

```go
k := koanf.New(".")
k.Load(file.Provider("config.hcl"), hcl.Parser())
k.Load(env.Provider("APP_", ".", nil), nil)
value := k.Get("some.dynamic.key")
```

## Maturity

- **koanf:** Stable, widely used for flexible configurations
- **confkit:** Production-ready v0.5.0, focused on typed, validated, struct-first config
