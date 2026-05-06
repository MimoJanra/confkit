# Security Policy

## Reporting Security Vulnerabilities

If you discover a security vulnerability in confkit, please **do not** open a public GitHub issue. Instead, email
security details to **mimojanra@gmail.com** with:

- Description of the vulnerability
- Steps to reproduce (if applicable)
- Potential impact
- Suggested fix (if you have one)

We will:

- Acknowledge receipt within 48 hours
- Investigate and assess severity
- Develop and test a fix
- Release a patched version
- Credit the reporter (unless they prefer anonymity)

**Please allow 30 days for a fix and release before public disclosure.**

---

## Security Features

### 1. Secret Redaction

confkit automatically redacts sensitive fields marked with `secret:"true"`:

```go
type Config struct {
    APIKey string `env:"API_KEY" secret:"true"`
    Password string `env:"DB_PASSWORD" secret:"true"`
}

// Errors and dumps will show: APIKey=***, Password=***
```

**Use this for:**

- Database passwords
- API keys and tokens
- OAuth secrets
- Private certificates
- Any credential or sensitive data

### 2. Validation at Load Time

Configuration is validated immediately, catching errors before the application starts:

```go
type Config struct {
    Port int `validate:"min=1,max=65535"`
}

// Invalid config fails at startup, not in production
cfg, err := confkit.Load[Config](confkit.FromEnv())
if err != nil {
    log.Fatal(confkit.Explain(err))  // Clear error message
}
```

### 3. Type Safety

Fully typed configuration eliminates stringly-typed configuration bugs:

```go
// ❌ Unsafe — typo at runtime
host := os.Getenv("HOST")
port := os.Getenv("PORT")  // Returns string, not int

// ✅ Safe — type-checked at compile time
type Config struct {
    Host string `env:"HOST"`
    Port int    `env:"PORT"`  // Type guaranteed
}
cfg, _ := confkit.Load[Config](confkit.FromEnv())
```

### 4. Minimal Dependencies

Core module depends only on:

- `gopkg.in/yaml.v3` — YAML parsing (widely audited)
- `github.com/pelletier/go-toml/v2` — TOML parsing (widely audited)
- Go stdlib

Enterprise sources (Vault, Consul, etcd, AWS) are optional, reducing attack surface for projects that don't use them.

---

## Security Best Practices

### 1. Mark All Secrets with `secret:"true"`

```go
type Config struct {
    // DO THIS
    APIKey    string `env:"API_KEY" secret:"true"`
    Password  string `env:"PASSWORD" secret:"true"`
    
    // NOT THIS
    APIKey2   string `env:"API_KEY2"`
}
```

### 2. Use Environment Variables for Secrets

Avoid embedding secrets in config files:

```go
// config.yaml — checked into version control
port: 8080
log_level: info

// Environment variables — NOT in version control
DATABASE_URL=postgres://user:PASSWORD@localhost/db
API_KEY=sk_live_xxx
```

### 3. Validate with Bounds and Enums

```go
type Config struct {
    // ✅ Bounded port range
    Port     int    `env:"PORT" validate:"min=1,max=65535"`
    
    // ✅ Enum validation
    Environment string `env:"ENV" validate:"oneof=dev,staging,prod"`
    
    // ✅ String length limits
    Name     string `env:"APP_NAME" validate:"min=1,max=128"`
}
```

### 4. Use Cloud Secret Managers

For production, use:

- **HashiCorp Vault** — `confkit/vault`
- **AWS Secrets Manager** — `confkit/aws`
- **Kubernetes Secrets** — built-in
- **Consul KV** — `confkit/consul`
- **etcd** — `confkit/etcd`

```go
cfg, err := confkit.Load[Config](
    confkit.FromVault("https://vault.example.com", auth, "/secret/app"),
    confkit.FromEnv(),  // Override with env if present
)
```

### 5. Validate Sources in Order

Sources are evaluated left-to-right. Put most-trusted sources last:

```go
cfg, err := confkit.Load[Config](
    confkit.FromYAML("defaults.yaml"),        // Base defaults
    confkit.FromYAML("environment.yaml"),    // Environment-specific
    confkit.FromEnv(),                        // Runtime overrides (highest trust)
)
```

### 6. Enable Hot Reload Only When Safe

Hot reload can pick up config changes without restart:

```go
cfg, watcher, err := confkit.LoadWithWatcher[Config](
    "config.yaml",
    confkit.FromYAML("config.yaml"),
)

go func() {
    for newCfg := range watcher.Changes() {
        // Validate before applying
        if validateNewConfig(newCfg) {
            applyConfig(newCfg)
        }
    }
}()
```

### 7. Log Safely

confkit's error messages automatically redact secrets, but always use them:

```go
cfg, err := confkit.Load[Config](confkit.FromEnv())
if err != nil {
    // ✅ SAFE — secrets are redacted
    log.Printf("Config error: %v\n", confkit.Explain(err))
    
    // ❌ UNSAFE — might leak secrets
    log.Printf("Config error: %v\n", err.Error())
}
```

---

## Dependency Security

### Core Dependencies

- **gopkg.in/yaml.v3** — v3.0.1+ required
    - Widely used, actively maintained
    - CVE monitoring via Go's vulnerability database

- **github.com/pelletier/go-toml/v2** — v2.0.0+ required
    - Maintained, CVE monitoring enabled

### Enterprise Dependencies

Optional modules bring their own dependencies:

- **Vault:** `github.com/hashicorp/vault/api`
- **Consul:** `github.com/hashicorp/consul/api`
- **etcd:** `go.etcd.io/etcd/client/v3`
- **AWS:** `github.com/aws/aws-sdk-go-v2`

These are only required if you use the corresponding source.

### Vulnerability Scanning

We use:

- **Go's built-in vulnerability scanner:** `go list -json -m all | nancy sleuth`
- **GitHub Dependabot:** Automated PR for dependency updates
- **Manual audits** before releases

Run locally:

```bash
go list -json -m all | nancy sleuth
```

---

## Code Review & Testing

- **Code review:** All changes reviewed by maintainers
- **Unit tests:** 80%+ coverage on core logic
- **Integration tests:** Multiple sources, precedence, nested structs
- **Example-based tests:** Every documented example is tested

Before release:

```bash
go test -v ./...
go test -cover ./...
```

---

## Supported Versions

Security fixes are released for:

| Version | Status | Support                |
|---------|--------|------------------------|
| v0.5.x  | Active | Current + bug fixes    |
| v0.4.x  | Stable | Security fixes only    |
| v0.3.x  | Legacy | Critical security only |
| < v0.3  | EOL    | No support             |

**Recommendation:** Upgrade to latest stable version regularly.

---

## Security Disclosure Timeline

Example timeline for a reported vulnerability:

- **Day 1:** Reporter notifies security team
- **Day 2:** Acknowledgment + severity assessment
- **Day 3-7:** Fix development and testing
- **Day 8:** Pre-release notification to major users
- **Day 9:** Release patched version + security advisory
- **Day 14:** Public disclosure (optional)

---

## PII & Data Protection

confkit does **not**:

- Store configuration in logs (except via explicit `Explain()`)
- Cache sensitive values unencrypted
- Transmit configuration anywhere
- Track usage or phone home

confkit **does**:

- Redact secrets in error messages
- Support secure sources (Vault, AWS Secrets Manager)
- Validate at startup (fail fast)

---

## Questions?

For security questions (non-vulnerability):

- Open a GitHub Discussion
- Email **mimojanra@gmail.com**
- Read [CONTRIBUTING.md](CONTRIBUTING.md)

For vulnerabilities:

- Email **mimojanra@gmail.com** with details
- Do **not** open a public issue
- Allow 30 days for fix + release
