package main

import (
	"fmt"
	"log"

	"github.com/MimoJanra/confkit"
)

// CLIToolConfig demonstrates configuration for a command-line tool
// using flags, environment variables, and config files.
type CLIToolConfig struct {
	Input struct {
		File      string   `flag:"input,i" env:"FILE" validate:"required" help:"Input file path"`
		Format    string   `flag:"format,f" env:"FORMAT" default:"json" validate:"oneof=json,yaml,csv,xml" help:"Input file format"`
		Encoding  string   `flag:"encoding,e" env:"ENCODING" default:"utf-8" help:"Input file encoding"`
	} `prefix:"INPUT_"`

	Output struct {
		File       string `flag:"output,o" env:"FILE" validate:"required" help:"Output file path"`
		Format     string `flag:"format,f" env:"FORMAT" default:"json" validate:"oneof=json,yaml,csv,xml" help:"Output file format"`
		Overwrite  bool   `flag:"overwrite,w" env:"OVERWRITE" default:"false" help:"Overwrite output file if exists"`
		Prettify   bool   `flag:"prettify,p" env:"PRETTIFY" default:"true" help:"Prettify output"`
	} `prefix:"OUTPUT_"`

	Processing struct {
		Threads       int    `flag:"threads,t" env:"THREADS" default:"4" validate:"min=1,max=32" help:"Number of processing threads"`
		ChunkSize     int    `flag:"chunk-size,c" env:"CHUNK_SIZE" default:"1000" validate:"min=1,max=1000000" help:"Chunk size for batch processing"`
		SkipErrors    bool   `flag:"skip-errors" env:"SKIP_ERRORS" default:"false" help:"Skip errors and continue processing"`
		ErrorLogFile  string `flag:"error-log,el" env:"ERROR_LOG_FILE" help:"Log file for errors"`
	} `prefix:"PROCESSING_"`

	Validation struct {
		Strict     bool   `flag:"strict" env:"STRICT" default:"true" help:"Strict validation mode"`
		AllowNull  bool   `flag:"allow-null" env:"ALLOW_NULL" default:"false" help:"Allow null values"`
		MaxErrors  int    `flag:"max-errors,me" env:"MAX_ERRORS" default:"100" validate:"min=1" help:"Max validation errors before stopping"`
		SchemaFile string `flag:"schema,s" env:"SCHEMA_FILE" help:"Validation schema file"`
	} `prefix:"VALIDATION_"`

	Logging struct {
		Verbose     bool   `flag:"verbose,v" env:"VERBOSE" default:"false" help:"Verbose output"`
		LogFile     string `flag:"log,l" env:"FILE" help:"Log file path (empty = stdout)"`
		LogLevel    string `flag:"log-level,ll" env:"LEVEL" default:"info" validate:"oneof=debug,info,warn,error" help:"Log level"`
		DebugMode   bool   `flag:"debug" env:"DEBUG" default:"false" help:"Enable debug mode"`
	} `prefix:"LOG_"`

	Performance struct {
		EnableCache  bool   `flag:"cache" env:"CACHE" default:"true" help:"Enable caching"`
		CacheSize    int    `flag:"cache-size,cs" env:"CACHE_SIZE" default:"100" validate:"min=1,max=10000" help:"Cache size in MB"`
		EnableCompression bool `flag:"compress" env:"COMPRESSION" default:"false" help:"Enable compression"`
		CompressionLevel int `flag:"compress-level,cl" env:"COMPRESSION_LEVEL" default:"6" validate:"min=1,max=9" help:"Compression level"`
	} `prefix:"PERFORMANCE_"`

	Advanced struct {
		CustomConfig string `flag:"config,C" env:"CONFIG_FILE" help:"Custom config file path"`
		DryRun       bool   `flag:"dry-run" env:"DRY_RUN" default:"false" help:"Dry run mode (no changes)"`
		Timeout      int    `flag:"timeout" env:"TIMEOUT" default:"300" validate:"min=1,max=3600" help:"Operation timeout in seconds"`
	} `prefix:"ADVANCED_"`
}

// ExampleCLITool shows how to use confkit for a CLI tool configuration.
func ExampleCLITool() error {
	cfg, err := confkit.Load[CLIToolConfig](
		confkit.FromEnv(),
		confkit.FromFlags(),
		confkit.FromYAMLOptional("tool-config.yaml"),
	)
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	log.Printf("📖 Input: %s (format=%s, encoding=%s)", cfg.Input.File, cfg.Input.Format, cfg.Input.Encoding)
	log.Printf("💾 Output: %s (format=%s, overwrite=%v)", cfg.Output.File, cfg.Output.Format, cfg.Output.Overwrite)
	log.Printf("⚙️  Processing: %d threads, chunk_size=%d, skip_errors=%v", cfg.Processing.Threads, cfg.Processing.ChunkSize, cfg.Processing.SkipErrors)
	log.Printf("✅ Validation: strict=%v, schema=%s", cfg.Validation.Strict, cfg.Validation.SchemaFile)
	log.Printf("📊 Performance: cache=%v (size=%dMB), compression=%v (level=%d)", cfg.Performance.EnableCache, cfg.Performance.CacheSize, cfg.Performance.EnableCompression, cfg.Performance.CompressionLevel)

	if cfg.Advanced.DryRun {
		log.Printf("🔍 DRY RUN MODE - no changes will be made")
	}

	return nil
}
