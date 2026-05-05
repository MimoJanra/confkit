package main

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"time"

	"confkit"
	"confkit/schema"
)

type AppConfig struct {
	Port     int    `toml:"port" desc:"HTTP server port" default:"8080" validate:"min=1,max=65535"`
	Host     string `toml:"host" desc:"HTTP server host" default:"localhost" validate:"required"`
	Database struct {
		URL      string        `toml:"url" desc:"Database connection string" validate:"required"`
		MaxConn  int           `toml:"max_conn" default:"10" validate:"min=1,max=100"`
		Timeout  time.Duration `toml:"timeout" default:"30s" validate:"min=1s,max=5m"`
		Password string        `toml:"password" secret:"true" validate:"required"`
	} `toml:"database"`
	LogLevel string `toml:"log_level" desc:"Log level" default:"info" validate:"oneof=debug,info,warn,error"`
}

func main() {
	cfg, err := confkit.LoadWithOptions[AppConfig](
		confkit.WithSource(confkit.FromTOML("examples/config.toml")),
		confkit.WithSource(confkit.FromEnv()),
		confkit.WithValidator("dburl", func(v reflect.Value) error {
			if v.Kind() == reflect.String {
				s := v.String()
				if s == "" {
					return fmt.Errorf("database URL cannot be empty")
				}
				if !isValidDBURL(s) {
					return fmt.Errorf("invalid database URL format")
				}
			}
			return nil
		}),
	)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Println("✅ Configuration loaded successfully!")
	fmt.Printf("Config: %+v\n\n", cfg)

	// Generate JSON Schema
	s, err := schema.GenerateSchema[AppConfig]()
	if err != nil {
		log.Fatalf("Failed to generate schema: %v", err)
	}

	schemaJSON, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal schema: %v", err)
	}

	fmt.Println("📋 Generated JSON Schema:")
	fmt.Println(string(schemaJSON))
	fmt.Println()

	// Generate Markdown documentation
	docs, err := schema.GenerateMarkdown[AppConfig]()
	if err != nil {
		log.Fatalf("Failed to generate docs: %v", err)
	}

	fmt.Println("📖 Generated Documentation:")
	fmt.Println(docs)
	fmt.Println()

	// Generate CLI help text
	help, err := schema.GenerateCLIHelp[AppConfig]()
	if err != nil {
		log.Fatalf("Failed to generate help: %v", err)
	}

	fmt.Println("❓ Generated CLI Help:")
	fmt.Println(help)
}

func isValidDBURL(url string) bool {
	// Simple validation: check for common database URL prefixes
	if len(url) < 5 {
		return false
	}
	if len(url) >= 7 && url[:7] == "postgres" {
		return true
	}
	if len(url) >= 5 && url[:5] == "mysql" {
		return true
	}
	if len(url) >= 7 && url[:7] == "mongodb" {
		return true
	}
	if len(url) >= 6 && url[:6] == "sqlite" {
		return true
	}
	return false
}
