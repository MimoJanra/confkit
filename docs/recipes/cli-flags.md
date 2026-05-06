---
layout: default
title: "Recipe: CLI Flags"
---

# Recipe: CLI Flags

Load configuration from command-line flags. Useful for CLI tools and scripts.

## Use Case

- CLI applications and tools
- One-time jobs and scripts
- User-facing applications

## Code

```go
package main

import (
    "flag"
    "log"
    "github.com/MimoJanra/confkit"
)

type Config struct {
    Input      string `flag:"input" short:"i" validate:"required" help:"Input file (required)"`
    Output     string `flag:"output" short:"o" default:"stdout" help:"Output file"`
    Format     string `flag:"format" short:"f" validate:"oneof=json yaml csv" default:"json" help:"Output format"`
    Verbose    bool   `flag:"verbose" short:"v" help:"Verbose output"`
    Threads    int    `flag:"threads" short:"t" default:"4" validate:"min=1,max=32" help:"Number of threads"`
    MaxSize    int    `flag:"max-size" default:"100" help:"Max size in MB"`
}

func main() {
    cfg, err := confkit.Load[Config](
        confkit.FromFlags(),
    )
    if err != nil {
        log.Fatal(confkit.Explain(err))
    }
    
    log.Printf("Input: %s", cfg.Input)
    log.Printf("Output: %s", cfg.Output)
    log.Printf("Format: %s", cfg.Format)
    log.Printf("Verbose: %v", cfg.Verbose)
    log.Printf("Threads: %d", cfg.Threads)
}
```

## CLI Usage

All three flag forms are supported:

```bash
# Show help
go run main.go -h

# Required argument
go run main.go --input data.txt

# --key=value form (equals sign)
go run main.go --input=data.txt --output=output.json

# --key value form (space-separated)
go run main.go --input data.txt --output output.json --format yaml

# -k value short form (space-separated)
go run main.go -i data.txt -o output.json -f yaml -v

# Mixed: long and short, equals and space
go run main.go --input=data.txt -o output.json --threads 8 --max-size=500
```

## Mixed Sources (YAML + Env + Flags)

Combine flags with files and environment. The first source to provide a value wins; later sources fill in only unset fields:

```go
cfg, err := confkit.Load[Config](
    confkit.FromFlags(),                // Highest priority — checked first
    confkit.FromEnv(),                  // Fills in what flags did not set
    confkit.FromYAML("config.yaml"),    // Fallback file defaults
)
```

Usage:

```bash
# Base config from file, override with env, then flags
export OUTPUT=env-output.json
go run main.go -i data.txt -o flag-output.json
# Result: Uses flag-output.json (flags win because they are checked first)
```

## Real-World Example: Image Processor

```go
package main

import (
    "fmt"
    "log"
    "github.com/MimoJanra/confkit"
)

type Config struct {
    Input       string `flag:"input" short:"i" validate:"required" help:"Input image file"`
    Output      string `flag:"output" short:"o" default:"output.jpg" help:"Output image file"`
    Quality     int    `flag:"quality" short:"q" default:"85" validate:"min=1,max=100" help:"JPEG quality (1-100)"`
    Width       int    `flag:"width" short:"w" help:"Resize width"`
    Height      int    `flag:"height" short:"h" help:"Resize height"`
    Grayscale   bool   `flag:"grayscale" short:"g" help:"Convert to grayscale"`
    Rotate      int    `flag:"rotate" short:"r" default:"0" validate:"oneof=0 90 180 270" help:"Rotation degrees"`
}

func main() {
    cfg, err := confkit.Load[Config](confkit.FromFlags())
    if err != nil {
        log.Fatal(confkit.Explain(err))
    }
    
    fmt.Printf("Processing %s → %s\n", cfg.Input, cfg.Output)
    fmt.Printf("Quality: %d%%\n", cfg.Quality)
    if cfg.Width > 0 || cfg.Height > 0 {
        fmt.Printf("Resize: %dx%d\n", cfg.Width, cfg.Height)
    }
    if cfg.Grayscale {
        fmt.Println("Apply: Grayscale")
    }
    if cfg.Rotate > 0 {
        fmt.Printf("Apply: Rotate %d°\n", cfg.Rotate)
    }
}
```

## Usage Examples

```bash
# Simple usage
go run main.go -i photo.jpg -o thumbnail.jpg -q 80

# With transformations
go run main.go -i photo.jpg -o output.jpg -w 400 -h 300 -g

# Full example
go run main.go \
  --input photo.jpg \
  --output final.jpg \
  --quality 90 \
  --width 1920 \
  --height 1080 \
  --rotate 90
```

## Built-in Help

confkit automatically generates help from struct tags:

```go
type Config struct {
    Input   string `flag:"input" short:"i" validate:"required" help:"Input file (required)"`
    Output  string `flag:"output" short:"o" default:"stdout" help:"Output file"`
    Verbose bool   `flag:"verbose" short:"v" help:"Verbose output"`
}
```

Help output:

```
Usage:
  -i, --input       Input file (required)
  -o, --output      Output file (default: stdout)
  -v, --verbose     Verbose output
```

## Flag Types

### String

```go
Input string `flag:"input" short:"i"`
```

### Integer

```go
Threads int `flag:"threads" short:"t" default:"4"`
```

### Boolean

```go
Verbose bool `flag:"verbose" short:"v"`
Debug   bool `flag:"debug" short:"d"`
```

### Slices

```go
Include []string `flag:"include" short:"I"` // -I a -I b
Exclude []string `flag:"exclude" short:"E"` // -E x -E y
```

## Validation with Flags

Validation rules are checked after parsing:

```go
type Config struct {
    Port      int    `flag:"port" short:"p" default:"8080" validate:"min=1,max=65535"`
    Format    string `flag:"format" short:"f" validate:"oneof=json yaml csv"`
    Threads   int    `flag:"threads" validate:"min=1,max=32"`
}
```

Invalid flag:

```bash
go run main.go --port 99999
# Error:
# Invalid configuration:
#
#   Port
#     error: must be between 1 and 65535, got 99999
#     source: flags (--port)
```

## Short vs Long Flags

```go
type Config struct {
    // Both --verbose and -v work
    Verbose bool `flag:"verbose" short:"v"`
    
    // Only --input and -i work
    Input   string `flag:"input" short:"i" validate:"required"`
    
    // No short form
    NoShort string `flag:"no-short"`
}
```

## Environment + Flags Precedence

The first source to provide a value wins. Put flags first so they take priority over env vars:

```go
cfg, err := confkit.Load[Config](
    confkit.FromFlags(),    // Higher priority — checked first
    confkit.FromEnv(),      // Fills in what flags did not set
)
```

Usage:

```bash
export OUTPUT=env-output.json
go run main.go --output flag-output.json
# Result: Uses flag-output.json (flag wins because it is checked first)
```

## Building a Complete CLI Tool

```go
package main

import (
    "fmt"
    "log"
    "github.com/MimoJanra/confkit"
    "github.com/MimoJanra/confkit/schema"
)

type Config struct {
    Command  string `flag:"command" short:"c" validate:"required" help:"Command to run"`
    Verbose  bool   `flag:"verbose" short:"v" help:"Verbose output"`
    Dry      bool   `flag:"dry-run" help:"Dry run mode"`
    Config   string `flag:"config" help:"Config file"`
    Workers  int    `flag:"workers" short:"w" default:"4" validate:"min=1,max=32"`
}

func main() {
    // Try to parse flags
    cfg, err := confkit.Load[Config](confkit.FromFlags())
    if err != nil {
        fmt.Println(confkit.Explain(err))
        fmt.Println("\nUsage:")
        fmt.Println(schema.GenerateCLIHelp[Config]())
        return
    }
    
    log.Printf("Running: %s (workers=%d, verbose=%v)", cfg.Command, cfg.Workers, cfg.Verbose)
}
```

## Best Practices

1. **Mark required flags**
   ```go
   Input string `flag:"input" validate:"required"`
   ```

2. **Provide defaults for optional flags**
   ```go
   Output string `flag:"output" default:"stdout"`
   ```

3. **Add helpful descriptions**
   ```go
   Port int `flag:"port" help:"Server port (1-65535)"`
   ```

4. **Use short flags for common options**
   ```go
   Verbose bool `flag:"verbose" short:"v"`
   ```

5. **Validate flag values**
   ```go
   Port int `validate:"min=1,max=65535"`
   ```

## See Also

- **[Getting Started](../docs/getting-started.md)** — Loading from flags
- **[Validation](../docs/validation.md)** — Validation rules
- **[Schema Generation](../docs/schema-generation.md)** — Auto-generating help
