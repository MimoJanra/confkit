# confkit Examples

Complete, production-ready examples demonstrating different use cases of confkit.

## Examples Overview

### 1. Web Service (`web_service_example.go`)

A typical web service with database, cache, and logging configuration.

**Features:**
- Nested struct configuration with prefixes
- Time.Duration fields (timeouts, TTLs)
- Secret field redaction (passwords)
- Comprehensive validation (min/max, oneof)
- Multiple configuration sources (env, YAML)

**Run:**
```bash
# Using environment variables
export APP_NAME="myapp"
export SERVER_PORT=8080
export DB_HOST="localhost"
export DB_USER="postgres"
export DB_PASSWORD="secret"
export DB_NAME="mydb"
go run web_service_example.go

# Using config file
go run web_service_example.go web_service_config.yaml
```

**Config file:** `web_service_config.yaml`

---

### 2. Microservice (`microservice_example.go`)

Enterprise microservice with multiple integrations: PostgreSQL, Redis, message queue, auth, observability.

**Features:**
- Complex nested configuration
- Database connection pooling settings
- Message queue integration (RabbitMQ, Kafka, NATS)
- JWT authentication configuration
- Rate limiting setup
- Observability (metrics, tracing, logging)
- Feature flags

**Run:**
```bash
# Minimal required environment variables
export SERVICE_NAME="payment-service"
export POSTGRES_HOST="localhost"
export POSTGRES_USER="admin"
export POSTGRES_PASSWORD="password"
export POSTGRES_DATABASE="payments"
export REDIS_HOST="localhost"
export MQ_HOST="localhost"
export MQ_USER="guest"
export MQ_PASSWORD="guest"
export AUTH_JWT_SECRET="secret-key"

go run microservice_example.go
```

**Config file:** `microservice_config.yaml`

---

### 3. CLI Tool (`cli_tool_example.go`)

Command-line tool configuration with input/output, processing parameters, and performance options.

**Features:**
- Command-line flag support (e.g., `--input`, `--output`)
- Multiple file format support (JSON, YAML, CSV, XML)
- Processing parallelization settings
- Validation and error handling configuration
- Performance tuning options (caching, compression)
- Dry-run mode support

**Run:**
```bash
# Using environment variables
export INPUT_FILE="data.json"
export OUTPUT_FILE="result.json"
export PROCESSING_THREADS=8
go run cli_tool_example.go

# Using flags (when integrated with flag parsing)
go run cli_tool_example.go --input data.json --output result.json --threads 8
```

---

### 4. Cloud-Native (`cloud_native_example.go`)

Kubernetes and cloud-native deployment configuration supporting multiple secret backends.

**Features:**
- Kubernetes ConfigMap integration
- AWS Secrets Manager support
- HashiCorp Vault integration
- Health check and probe paths (Kubernetes)
- Container resource limits
- Horizontal Pod Autoscaler (HPA) settings
- TLS/mTLS configuration
- Service mesh integration

**Run:**
```bash
# In Kubernetes environment (with service account)
kubectl set env deployment/myapp \
  -e APP_NAME=myapp \
  -e SERVER_PORT=8080
go run cloud_native_example.go

# Locally with Vault (requires Vault running)
export VAULT_ADDR="http://localhost:8200"
export VAULT_TOKEN="hvs.CAESIMgR..."
go run cloud_native_example.go
```

---

### 5. Full Setup with Schema (`full_setup_example.go`)

Complete example demonstrating schema generation capabilities (JSON Schema, Markdown docs, CLI help).

**Features:**
- Full configuration structure with all field types
- JSON Schema generation for API documentation
- Markdown documentation generation
- CLI help text generation
- Custom validators
- Database URL validation

**Run:**
```bash
# Set required environment variables
export PORT=8080
export DATABASE_URL="postgres://localhost/myapp"
export LOG_LEVEL="info"

go run full_setup_example.go
```

This will generate and display:
- Configuration loaded successfully
- JSON Schema representation
- Markdown documentation
- CLI help text

**Config file:** `config.toml`

---

## Configuration Files

### Hierarchical Configuration Pattern

All examples support a three-level configuration hierarchy (in priority order):

1. **Environment Variables** (highest priority)
   - Override everything
   - Prefix-based naming (e.g., `DB_HOST`, `SERVICE_PORT`)

2. **Environment-Specific Config** (medium priority)
   - `web_service_config.yaml` for web service
   - `microservice_config.yaml` for microservice
   - Environment-based overrides

3. **Base Defaults** (lowest priority)
   - `defaults.yaml` - shared defaults for all examples
   - Fallback values for all configuration keys

**Loading pattern:**
```go
cfg, err := confkit.Load[Config](
    confkit.FromEnv(),                    // Priority 1
    confkit.FromYAML("config.yaml"),      // Priority 2
    confkit.FromYAML("defaults.yaml"),    // Priority 3
)
```

---

## Running the Tests

All examples have comprehensive tests in `examples_test.go`.

```bash
# Run all tests
go test ./examples -v

# Run specific test
go test ./examples -v -run TestWebServiceConfig

# Run with coverage
go test ./examples -cover
```

### Test Coverage

- ✅ Default values application
- ✅ Environment variable overrides
- ✅ Time.Duration parsing
- ✅ Validation enforcement
- ✅ Prefix mapping (nested structs)
- ✅ Secret field redaction
- ✅ Type coercion and parsing

---

## Key Patterns Demonstrated

### 1. Environment Variable Prefixing

```go
type DatabaseConfig struct {
    Host string `env:"HOST"`
    Port int    `env:"PORT"`
}

type Config struct {
    DB DatabaseConfig `prefix:"DB_"`
}

// Reads from: DB_HOST, DB_PORT
```

### 2. Time Duration Configuration

```go
type Config struct {
    Timeout   time.Duration `env:"TIMEOUT" default:"30s"`
    ReadTime  time.Duration `env:"READ_TIME" default:"15s"`
    WriteTime time.Duration `env:"WRITE_TIME" default:"10s"`
}
```

### 3. Secret Redaction

```go
type Config struct {
    Password string `env:"PASSWORD" secret:"true" validate:"required"`
    APIKey   string `env:"API_KEY" secret:"true"`
}

// Passwords appear as [REDACTED] in error messages and logs
```

### 4. Complex Validation

```go
type Config struct {
    Port     int `validate:"min=1,max=65535"`
    LogLevel string `validate:"oneof=debug,info,warn,error"`
    Count    int `validate:"min=0,max=100"`
}
```

### 5. Nested Struct Configuration

```go
type Config struct {
    Server struct {
        Host string
        Port int
    } `prefix:"SERVER_"`
    
    Database struct {
        Host string
        Port int
    } `prefix:"DB_"`
}

// Reads from: SERVER_HOST, SERVER_PORT, DB_HOST, DB_PORT
```

---

## Environment Setup

### For Web Service Example

```bash
export APP_NAME="my-app"
export SERVER_HOST="0.0.0.0"
export SERVER_PORT="8080"
export DB_HOST="localhost"
export DB_USER="postgres"
export DB_PASSWORD="password"
export DB_NAME="mydb"
export CACHE_ENABLED="true"
export LOG_LEVEL="info"
```

### For Microservice Example

```bash
export SERVICE_NAME="order-service"
export SERVICE_PORT="8000"
export AUTH_JWT_SECRET="your-secret-key"
export POSTGRES_HOST="localhost"
export POSTGRES_USER="postgres"
export POSTGRES_PASSWORD="password"
export POSTGRES_DATABASE="orders"
export REDIS_HOST="localhost"
export MQ_HOST="localhost"
export MQ_USER="guest"
export MQ_PASSWORD="guest"
export OBS_LOG_LEVEL="info"
export FEATURE_ENABLE_NEW_API="false"
```

### For CLI Tool Example

```bash
export INPUT_FILE="input.json"
export OUTPUT_FILE="output.json"
export PROCESSING_THREADS="4"
export VALIDATION_STRICT="true"
export PERFORMANCE_CACHE="true"
```

---

## Best Practices Shown

1. **Use snake_case in YAML/JSON** - confkit automatically converts to CamelCase
2. **Prefix nested structs** - improves env var organization
3. **Set reasonable defaults** - reduces configuration burden
4. **Validate early** - catch configuration errors at startup
5. **Mark secrets** - ensure sensitive data is redacted
6. **Use time.Duration** - instead of seconds/milliseconds as integers
7. **Document with tags** - use `help:` and `desc:` tags for clarity
8. **Support multiple sources** - env vars, files, and secret managers
9. **Test configuration** - unit test all examples
10. **Use optional files** - FromYAMLOptional for non-critical configs

---

## Further Reading

- [confkit Documentation](/docs/getting-started/)
- [Validation Rules](/docs/validation/)
- [Configuration Sources](/docs/sources/)
- [Secret Redaction](/docs/secret-redaction/)
- [Schema Generation](/docs/schema-generation/)
