---
layout: default
title: Schema Generation — confkit
---

# Schema Generation

confkit can automatically generate JSON Schema, Markdown documentation, and CLI help text from your struct definitions.

## JSON Schema Generation

Generate JSON Schema (draft-07 compatible) from a config struct:

```go
import (
    "encoding/json"
    "log"
    "github.com/MimoJanra/confkit/schema"
)

type Config struct {
    Port     int    `env:"PORT" default:"8080" validate:"min=1,max=65535" help:"Server port"`
    Host     string `env:"HOST" default:"localhost" help:"Server hostname"`
    Database string `env:"DATABASE_URL" validate:"required" secret:"true" help:"Database connection URL"`
}

func main() {
    s, err := schema.GenerateSchema[Config]()
    if err != nil {
        log.Fatal(err)
    }
    
    data, _ := json.MarshalIndent(s, "", "  ")
    fmt.Println(string(data))
}
```

**Output:**
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "title": "Config",
  "properties": {
    "Port": {
      "type": "integer",
      "default": 8080,
      "minimum": 1,
      "maximum": 65535,
      "description": "Server port"
    },
    "Host": {
      "type": "string",
      "default": "localhost",
      "description": "Server hostname"
    },
    "Database": {
      "type": "string",
      "description": "Database connection URL"
    }
  },
  "required": ["Database"]
}
```

## Markdown Documentation

Generate a Markdown table documenting all fields:

```go
import "github.com/MimoJanra/confkit/schema"

type Config struct {
    Port int    `env:"PORT" default:"8080" validate:"min=1,max=65535" help:"Server port"`
    Host string `env:"HOST" default:"localhost" help:"Server hostname"`
}

md := schema.GenerateMarkdown[Config]()
fmt.Println(md)
```

**Output:**
```markdown
# Config

| Field | Type | Required | Default | Validation | Description |
|-------|------|----------|---------|-----------|-------------|
| Host | string | no | localhost | | Server hostname |
| Port | int | no | 8080 | min=1,max=65535 | Server port |
```

Fields are output in **alphabetical order**, making the output deterministic across runs and Go versions.

## CLI Help Text

Generate `--help` style output for command-line tools:

```go
import "github.com/MimoJanra/confkit/schema"

type Config struct {
    Verbose bool   `flag:"verbose" short:"v" help:"Enable verbose output"`
    Output  string `flag:"output" short:"o" help:"Output file" default:"stdout"`
    Input   string `flag:"input" validate:"required" help:"Input file"`
}

help := schema.GenerateCLIHelp[Config]()
fmt.Println(help)
```

**Output:**
```
Options:
      --input        Input file (required)
  -o, --output       Output file (default: stdout)
  -v, --verbose      Enable verbose output
```

Fields are output in **alphabetical order**, making the output deterministic across runs and Go versions.

## Struct Tags for Schema

### `help`

Description for documentation and help text:

```go
type Config struct {
    Port int `env:"PORT" help:"Server port (1-65535)"`
}
```

Appears in:
- JSON Schema `description` field
- Markdown table
- CLI help text

### `hidden`

Hide from CLI help output:

```go
type Config struct {
    // Normal field
    Port int `env:"PORT" help:"Server port"`
    
    // Hidden from help
    InternalID string `env:"INTERNAL_ID" hidden:"true"`
}
```

Hidden fields still appear in JSON Schema and Markdown.

### `default`

Default values appear in schema:

```go
type Config struct {
    Port int `env:"PORT" default:"8080"`
}
```

JSON Schema:
```json
{"default": 8080}
```

Markdown:
```markdown
| Port | int | no | 8080 | | Server port |
```

### `validate`

Validation rules appear in schema:

```go
type Config struct {
    Port   int    `validate:"min=1,max=65535"`
    LogLevel string `validate:"oneof=debug info warn error"`
}
```

JSON Schema:
```json
{
  "Port": {"minimum": 1, "maximum": 65535},
  "LogLevel": {"enum": ["debug", "info", "warn", "error"]}
}
```

### `secret`

Secret fields are marked in schema:

```go
type Config struct {
    Password string `env:"DB_PASSWORD" secret:"true" help:"Database password"`
}
```

JSON Schema:
```json
{
  "Password": {
    "type": "string",
    "writeOnly": true,
    "description": "Database password"
  }
}
```

## Nested Struct Schemas

Nested structs generate nested JSON Schema:

```go
type Config struct {
    Server struct {
        Host string `env:"HOST" default:"localhost" help:"Server hostname"`
        Port int    `env:"PORT" default:"8080" help:"Server port"`
    } `help:"Server configuration"`
    
    Database struct {
        URL string `env:"URL" validate:"required" help:"Database URL"`
    } `help:"Database configuration"`
}
```

JSON Schema:
```json
{
  "properties": {
    "Server": {
      "type": "object",
      "description": "Server configuration",
      "properties": {
        "Host": {"type": "string", "default": "localhost"},
        "Port": {"type": "integer", "default": 8080}
      }
    },
    "Database": {
      "type": "object",
      "description": "Database configuration",
      "properties": {
        "URL": {"type": "string"}
      },
      "required": ["URL"]
    }
  }
}
```

## Markdown for Documentation

Use generated Markdown in your docs:

```go
// Generate and save to file
md := schema.GenerateMarkdown[Config]()
os.WriteFile("docs/config-reference.md", []byte(md), 0644)
```

Or embed in your README:

```bash
go run ./cmd/genconfig/main.go > docs/config-reference.md
```

## CLI Integration

Use generated help in your CLI application:

```go
package main

import (
    "flag"
    "fmt"
    "github.com/MimoJanra/confkit"
    "github.com/MimoJanra/confkit/schema"
)

type Config struct {
    Verbose bool   `flag:"verbose" short:"v" help:"Enable verbose output"`
    Config  string `flag:"config" help:"Config file" default:"config.yaml"`
}

func main() {
    flag.Parse()
    
    if flag.Bool("help", false, "") {
        help := schema.GenerateCLIHelp[Config]()
        fmt.Println(help)
        return
    }
    
    // Load and use config
    cfg, _ := confkit.Load[Config](confkit.FromFlags())
    fmt.Println(cfg)
}
```

## Real-World Example: Config Generator

Create a tool that generates config files:

```go
package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "github.com/MimoJanra/confkit"
    "github.com/MimoJanra/confkit/schema"
)

type AppConfig struct {
    Port     int    `env:"PORT" default:"8080"`
    Host     string `env:"HOST" default:"localhost"`
    Database string `env:"DATABASE_URL" validate:"required"`
}

func main() {
    format := flag.String("format", "json", "Output format: json, markdown, cli")
    flag.Parse()
    
    switch *format {
    case "json":
        s, _ := schema.GenerateSchema[AppConfig]()
        data, _ := json.MarshalIndent(s, "", "  ")
        fmt.Println(string(data))
        
    case "markdown":
        md := schema.GenerateMarkdown[AppConfig]()
        fmt.Println(md)
        
    case "cli":
        help := schema.GenerateCLIHelp[AppConfig]()
        fmt.Println(help)
    }
}
```

Run it:

```bash
go run main.go --format json > config.schema.json
go run main.go --format markdown > docs/config.md
go run main.go --format cli
```

## Schema Validation

Use generated JSON Schema to validate configs externally:

```go
import "github.com/xeipuuv/gojsonschema"

s, _ := schema.GenerateSchema[Config]()
schema := gojsonschema.NewGoLoader(s)
document := gojsonschema.NewStringLoader(`{"Port": 8080}`)

result, _ := gojsonschema.Validate(schema, document)
if result.Valid() {
    fmt.Println("Config is valid")
} else {
    for _, err := range result.Errors() {
        fmt.Println(err)
    }
}
```

## Best Practices

1. **Always add `help` descriptions**
   ```go
   Port int `env:"PORT" help:"Server port (1-65535)"`
   ```

2. **Use validation rules in tags**
   ```go
   Port int `env:"PORT" default:"8080" validate:"min=1,max=65535"`
   ```

3. **Mark secrets with `secret:"true"`**
   ```go
   Password string `env:"DB_PASSWORD" secret:"true"`
   ```

4. **Hide internal fields**
   ```go
   InternalID string `hidden:"true"`
   ```

5. **Generate and commit schema files**
   ```bash
   go run ./cmd/genconfig/ > docs/config.schema.json
   go run ./cmd/genconfig/ --format markdown > docs/config.md
   ```

6. **Use schema for documentation**
   - Share JSON Schema with external tools
   - Generate Markdown for your docs
   - Use CLI help for users

## Real-World Examples

See schema generation in action:

- **[Full Setup Example](https://github.com/MimoJanra/confkit/tree/main/examples)** — Demonstrates JSON Schema, Markdown docs, and CLI help generation
  - Generates all three formats automatically
  - Uses `help:` tags for documentation
  - Produces human-readable output

All examples in the repository include schema generation capabilities.

## Next Steps

- **[Sources](./sources.md)** — Configuration sources
- **[Examples](https://github.com/MimoJanra/confkit/tree/main/examples)** — Full examples including schema generation
