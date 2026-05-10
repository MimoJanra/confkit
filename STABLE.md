# API Stability Contract

As of **v1.0.0**, the following symbols are **frozen**: no signature changes, no removals, no semantic changes.
Additive changes (new exported symbols) are always allowed.

## Interfaces (frozen)

```
Source { Name() string; Lookup(context.Context, *FieldInfo) (any, bool, error) }
Option  (apply is unexported; only With*-constructors are public)
```

## Types (frozen)

```
FieldInfo
ErrorReport
FieldError
ErrorKind  (ErrorKindParse, ErrorKindValidation, ErrorKindIO)

AuditEntry
AuditLogger
LoadHookFunc
MiddlewareFunc

ConfigWatcher
ConfigChangeListener
ConfigDelta
ConfigChangeListenerWithDelta

DumpFormat  (FormatJSON, FormatYAML)
DumpOption

RotationEngine
RotationStrategy
RotationCallback
```

## Functions (frozen)

```go
Load[T](sources ...Source) (*T, error)
LoadContext[T](ctx context.Context, sources ...Source) (*T, error)
LoadWithOptions[T](opts ...Option) (*T, error)
LoadWithOptionsContext[T](ctx context.Context, opts ...Option) (*T, error)
LoadWithWatcher[T](filePath string, sources ...Source) (*T, *ConfigWatcher, error)
MustLoad[T](sources ...Source) *T
MustLoadContext[T](ctx context.Context, sources ...Source) *T
ValidateOnly[T](ctx context.Context, opts ...Option) (*T, error)

Dump[T](cfg T, opts ...DumpOption) ([]byte, error)
DumpString[T](cfg T, opts ...DumpOption) string
DumpYAML[T](cfg T, opts ...DumpOption) ([]byte, error)

Explain(err error) string
ScanFields(v any) []FieldInfo

FindFile(name string, dirs ...string) (string, bool)
FindSource(name string, dirs ...string) Source
DefaultSearchDirs(appName string) []string
OverlayPath(basePath, env string) string

// Source constructors
FromYAML(path string) Source
FromYAMLOptional(path string) Source
FromJSON(path string) Source
FromTOML(path string) Source
FromYAMLFiles(paths ...string) Source
FromJSONFiles(paths ...string) Source
FromTOMLFiles(paths ...string) Source
FromEnv() Source
FromFlags(flags *flag.FlagSet) Source
FromOverlay(base Source, env string) Source

// Option constructors
WithSource(source Source) Option
WithValidator(name string, fn func(reflect.Value) error) Option
WithModelValidator[T any](fn func(*T) error) Option
WithMiddleware(fn MiddlewareFunc) Option
WithInterpolationMaxDepth(depth int) Option
WithAuditLogger(fn AuditLogger) Option
WithLoadHook(fn LoadHookFunc) Option
WithContext(ctx context.Context) Option
WithDumpFormat(f DumpFormat) DumpOption
WithDumpRedactSecrets(redact bool) DumpOption

// ErrorReport methods
(*ErrorReport).Error() string
(*ErrorReport).Format() string
(*ErrorReport).AddError(fe FieldError)
(*ErrorReport).IsEmpty() bool
(*ErrorReport).FirstError() *FieldError
(*ErrorReport).Unwrap() []error

// ConfigWatcher methods
(*ConfigWatcher).AddListener(listener ConfigChangeListener)
(*ConfigWatcher).AddDeltaListener(listener ConfigChangeListenerWithDelta)
(*ConfigWatcher).Start()
(*ConfigWatcher).Stop()
(*ConfigWatcher).SetPollInterval(interval time.Duration)
```

## Struct tags (frozen)

```
env   yaml   json   toml   flag
default   validate   secret
prefix   short   hidden   desc
```

## NOT frozen

- Internal helpers (`anyToString`, `splitPath`, `lookupNested`, `splitMapPairs`, etc.)
- `schema/` subpackage — frozen in spirit but versioned independently
- Cloud submodules (`aws`, `vault`, `consul`, `etcd`, `k8s`) — each has its own `go.mod`
- `prometheus/`, `otel/` submodules — independent `go.mod`
- `structtags/` — internal utility, not a public contract

## Versioning policy after v1.0

```
1.0.x  — bug fixes, docs, tests only
1.x.0  — new additive symbols only
2.0.0  — any breaking change to a frozen symbol
         (requires deprecation notice in the preceding minor)
```
