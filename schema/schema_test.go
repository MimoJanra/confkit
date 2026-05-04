package schema

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestGenerateSchemaErrorOnNonStruct(t *testing.T) {
	_, err := GenerateSchema[int]()
	if err == nil {
		t.Error("expected error for non-struct type, got nil")
	}
}

func TestGenerateSchema(t *testing.T) {
	type Config struct {
		Port int    `json:"port" desc:"HTTP port" default:"8080" validate:"min=1,max=65535"`
		Host string `json:"host" desc:"HTTP host" default:"localhost" validate:"required"`
	}

	schema, err := GenerateSchema[Config]()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if schema.Type != "object" {
		t.Errorf("expected type 'object', got %s", schema.Type)
	}

	// Check properties
	if len(schema.Properties) != 2 {
		t.Errorf("expected 2 properties, got %d", len(schema.Properties))
	}

	// Check Port property
	portProp, ok := schema.Properties["port"]
	if !ok {
		t.Fatal("expected 'port' property")
	}
	if portProp.Type != "integer" {
		t.Errorf("expected port type 'integer', got %s", portProp.Type)
	}
	if portProp.Description != "HTTP port" {
		t.Errorf("expected description 'HTTP port', got %s", portProp.Description)
	}
	if portProp.Minimum == nil || *portProp.Minimum != 1 {
		t.Errorf("expected minimum 1, got %v", portProp.Minimum)
	}
	if portProp.Maximum == nil || *portProp.Maximum != 65535 {
		t.Errorf("expected maximum 65535, got %v", portProp.Maximum)
	}
	if portProp.Default != int64(8080) {
		t.Errorf("expected default 8080, got %v", portProp.Default)
	}

	// Check Host property
	hostProp, ok := schema.Properties["host"]
	if !ok {
		t.Fatal("expected 'host' property")
	}
	if hostProp.Type != "string" {
		t.Errorf("expected host type 'string', got %s", hostProp.Type)
	}

	// Check required fields
	if len(schema.Required) != 1 {
		t.Errorf("expected 1 required field, got %d", len(schema.Required))
	}
	if schema.Required[0] != "host" {
		t.Errorf("expected 'host' in required, got %v", schema.Required)
	}
}

func TestGenerateSchemaWithNested(t *testing.T) {
	type Database struct {
		Host string `json:"host" validate:"required"`
		Port int    `json:"port" default:"5432" validate:"min=1,max=65535"`
	}

	type Config struct {
		Port     int      `json:"port" default:"8080"`
		Database Database `json:"database"`
	}

	schema, err := GenerateSchema[Config]()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check nested database property
	dbProp, ok := schema.Properties["database"]
	if !ok {
		t.Fatal("expected 'database' property")
	}
	if dbProp.Type != "object" {
		t.Errorf("expected database type 'object', got %s", dbProp.Type)
	}

	// Check nested properties
	if len(dbProp.Properties) != 2 {
		t.Errorf("expected 2 nested properties, got %d", len(dbProp.Properties))
	}

	hostProp, ok := dbProp.Properties["host"]
	if !ok {
		t.Fatal("expected 'host' nested property")
	}
	if hostProp.Type != "string" {
		t.Errorf("expected host type 'string', got %s", hostProp.Type)
	}
}

func TestGenerateSchemaWithSecret(t *testing.T) {
	type Config struct {
		Password string `json:"password" secret:"true" validate:"required"`
	}

	schema, err := GenerateSchema[Config]()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	passProp, ok := schema.Properties["password"]
	if !ok {
		t.Fatal("expected 'password' property")
	}

	if !passProp.Secret {
		t.Errorf("expected secret flag to be true")
	}
}

func TestGenerateSchemaWithEnum(t *testing.T) {
	type Config struct {
		LogLevel string `json:"log_level" validate:"oneof=debug,info,warn,error"`
	}

	schema, err := GenerateSchema[Config]()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logProp, ok := schema.Properties["log_level"]
	if !ok {
		t.Fatal("expected 'log_level' property")
	}

	if len(logProp.Enum) != 4 {
		t.Errorf("expected 4 enum values, got %d", len(logProp.Enum))
	}
}

func TestGenerateSchemaWithDuration(t *testing.T) {
	type Config struct {
		Timeout time.Duration `json:"timeout" default:"30s"`
	}

	schema, err := GenerateSchema[Config]()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	timeoutProp, ok := schema.Properties["timeout"]
	if !ok {
		t.Fatal("expected 'timeout' property")
	}

	if timeoutProp.Type != "string" {
		t.Errorf("expected timeout type 'string', got %s", timeoutProp.Type)
	}
}

func TestGenerateSchemaJSON(t *testing.T) {
	type Config struct {
		Port int `json:"port" default:"8080"`
	}

	data, err := GenerateSchemaJSON[Config]()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var schema Schema
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatalf("failed to unmarshal schema: %v", err)
	}

	if schema.Type != "object" {
		t.Errorf("expected type 'object', got %s", schema.Type)
	}
}

func TestGenerateMarkdown(t *testing.T) {
	type Config struct {
		Port int    `json:"port" desc:"HTTP port" default:"8080" validate:"min=1,max=65535"`
		Host string `json:"host" desc:"HTTP host" default:"localhost" validate:"required"`
	}

	markdown, err := GenerateMarkdown[Config]()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that markdown contains expected sections
	if !strings.Contains(markdown, "## Configuration Fields") {
		t.Errorf("expected markdown to contain header")
	}

	if !strings.Contains(markdown, "port") {
		t.Errorf("expected markdown to contain 'port' field")
	}

	if !strings.Contains(markdown, "host") {
		t.Errorf("expected markdown to contain 'host' field")
	}

	if !strings.Contains(markdown, "HTTP port") {
		t.Errorf("expected markdown to contain 'HTTP port' description")
	}

	if !strings.Contains(markdown, "min=1") {
		t.Errorf("expected markdown to contain 'min=1' rule")
	}
}

func TestGenerateCLIHelp(t *testing.T) {
	type Config struct {
		Port int    `json:"port" desc:"HTTP port" default:"8080" validate:"min=1,max=65535"`
		Host string `json:"host" desc:"HTTP host" validate:"required"`
	}

	help, err := GenerateCLIHelp[Config]()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that help contains expected sections
	if !strings.Contains(help, "Usage:") {
		t.Errorf("expected help to contain 'Usage:'")
	}

	if !strings.Contains(help, "--port") {
		t.Errorf("expected help to contain '--port' flag")
	}

	if !strings.Contains(help, "--host") {
		t.Errorf("expected help to contain '--host' flag")
	}

	if !strings.Contains(help, "HTTP port") {
		t.Errorf("expected help to contain 'HTTP port' description")
	}

	if !strings.Contains(help, "default: 8080") {
		t.Errorf("expected help to contain 'default: 8080'")
	}

	if !strings.Contains(help, "required") {
		t.Errorf("expected help to contain 'required' for host field")
	}
}

func TestGenerateSchemaAllTypes(t *testing.T) {
	type Config struct {
		Count   int     `json:"count" default:"1"`
		Active  bool    `json:"active" default:"true"`
		Ratio   float64 `json:"ratio" validate:"min=0,max=1"`
		Tags    []string `json:"tags"`
		Name    string  `json:"name" validate:"min=3,max=50"`
	}

	schema, err := GenerateSchema[Config]()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// integer
	if p := schema.Properties["count"]; p == nil || p.Type != "integer" {
		t.Errorf("expected count to be integer, got %v", schema.Properties["count"])
	}

	// boolean
	if p := schema.Properties["active"]; p == nil || p.Type != "boolean" {
		t.Errorf("expected active to be boolean, got %v", schema.Properties["active"])
	}

	// number with min/max
	ratioProp := schema.Properties["ratio"]
	if ratioProp == nil || ratioProp.Type != "number" {
		t.Errorf("expected ratio to be number")
	}

	// array with items
	tagsProp := schema.Properties["tags"]
	if tagsProp == nil || tagsProp.Type != "array" {
		t.Errorf("expected tags to be array")
	}
	if tagsProp.Items == nil || tagsProp.Items.Type != "string" {
		t.Errorf("expected tags items to be string, got %v", tagsProp.Items)
	}

	// string with length constraints
	nameProp := schema.Properties["name"]
	if nameProp == nil || nameProp.Type != "string" {
		t.Errorf("expected name to be string")
	}
	if nameProp.MinLength == nil || *nameProp.MinLength != 3 {
		t.Errorf("expected name minLength 3, got %v", nameProp.MinLength)
	}
	if nameProp.MaxLength == nil || *nameProp.MaxLength != 50 {
		t.Errorf("expected name maxLength 50, got %v", nameProp.MaxLength)
	}
}

func TestMarkdownStringLengthConstraints(t *testing.T) {
	type Config struct {
		Name string `json:"name" validate:"min=3,max=50" desc:"User name"`
	}

	md, err := GenerateMarkdown[Config]()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(md, "3") {
		t.Errorf("expected markdown to contain length constraint 3, got:\n%s", md)
	}
}

func TestCLIHelpStringLengthConstraints(t *testing.T) {
	type Config struct {
		Name string `json:"name" validate:"min=3,max=50" desc:"User name"`
	}

	help, err := GenerateCLIHelp[Config]()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(help, "min length") {
		t.Errorf("expected CLI help to contain 'min length', got:\n%s", help)
	}
	if !strings.Contains(help, "max length") {
		t.Errorf("expected CLI help to contain 'max length', got:\n%s", help)
	}
}

func TestCLIHelpNestedRequired(t *testing.T) {
	type Database struct {
		Host string `json:"host" desc:"DB host" validate:"required"`
	}

	type Config struct {
		Database Database `json:"database"`
	}

	help, err := GenerateCLIHelp[Config]()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(help, "--database-host") {
		t.Errorf("expected '--database-host' flag in help, got:\n%s", help)
	}
	if !strings.Contains(help, "required") {
		t.Errorf("expected 'required' for nested host field, got:\n%s", help)
	}
}

func TestGenerateSchemaDefaults(t *testing.T) {
	type Config struct {
		Active  bool    `json:"active" default:"true"`
		Ratio   float64 `json:"ratio" default:"0.5"`
		Count   int     `json:"count" default:"42"`
		Name    string  `json:"name" default:"hello"`
	}

	schema, err := GenerateSchema[Config]()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p := schema.Properties["active"]; p == nil || p.Default != true {
		t.Errorf("expected active default true, got %v", schema.Properties["active"])
	}
	if p := schema.Properties["ratio"]; p == nil || p.Default != 0.5 {
		t.Errorf("expected ratio default 0.5, got %v", schema.Properties["ratio"])
	}
	if p := schema.Properties["name"]; p == nil || p.Default != "hello" {
		t.Errorf("expected name default 'hello', got %v", schema.Properties["name"])
	}
}

func TestCLIHelpConstraintsWithDefault(t *testing.T) {
	type Config struct {
		Port int    `json:"port" default:"8080" validate:"min=1,max=65535"`
		Name string `json:"name" default:"app" validate:"min=1,max=100"`
	}

	help, err := GenerateCLIHelp[Config]()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// port should show min/max constraints in the default line
	if !strings.Contains(help, "default: 8080") {
		t.Errorf("expected help to contain 'default: 8080', got:\n%s", help)
	}

	// name should show length constraints
	if !strings.Contains(help, "min length") {
		t.Errorf("expected help to contain 'min length', got:\n%s", help)
	}
}

func TestMarkdownAllTypes(t *testing.T) {
	type Config struct {
		Count   int      `json:"count"`
		Active  bool     `json:"active"`
		Ratio   float64  `json:"ratio"`
		Tags    []string `json:"tags"`
	}

	md, err := GenerateMarkdown[Config]()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(md, "int") {
		t.Errorf("expected markdown to contain 'int', got:\n%s", md)
	}
	if !strings.Contains(md, "bool") {
		t.Errorf("expected markdown to contain 'bool', got:\n%s", md)
	}
	if !strings.Contains(md, "float") {
		t.Errorf("expected markdown to contain 'float', got:\n%s", md)
	}
	if !strings.Contains(md, "array") {
		t.Errorf("expected markdown to contain 'array', got:\n%s", md)
	}
}

func TestSchemaFieldWithoutFormatTag(t *testing.T) {
	// When a field has no json/yaml/toml tag, getPropName falls back to snake_case
	type Config struct {
		ServerPort int  `default:"8080"` // No format tag: should become "server_port"
	}

	schema, err := GenerateSchema[Config]()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The field should appear as "server_port" (snake_case of ServerPort)
	if _, ok := schema.Properties["server_port"]; !ok {
		t.Errorf("expected 'server_port' property (snake_case fallback), got: %v", schema.Properties)
	}
}

func TestSchemaUintAndFloat32(t *testing.T) {
	type Config struct {
		Count  uint    `json:"count" validate:"min=0,max=1000"`
		Ratio  float32 `json:"ratio"`
	}

	schema, err := GenerateSchema[Config]()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p := schema.Properties["count"]; p == nil || p.Type != "integer" {
		t.Errorf("expected count to be integer")
	}
	if p := schema.Properties["ratio"]; p == nil || p.Type != "number" {
		t.Errorf("expected ratio to be number, got %v", schema.Properties["ratio"])
	}
}

func TestSchemaValidationPattern(t *testing.T) {
	type Config struct {
		Email string `json:"email" validate:"pattern=^[a-z]+$"`
	}

	schema, err := GenerateSchema[Config]()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p := schema.Properties["email"]; p == nil || p.Pattern == "" {
		t.Errorf("expected email to have pattern constraint")
	}
}

func TestMarkdownWithPattern(t *testing.T) {
	type Config struct {
		Code string `json:"code" validate:"pattern=^[A-Z]+$" desc:"Uppercase code"`
	}

	md, err := GenerateMarkdown[Config]()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(md, "pattern") {
		t.Errorf("expected markdown to show pattern rule, got:\n%s", md)
	}
}

func TestMarkdownWithNestedStruct(t *testing.T) {
	type Database struct {
		Host string `json:"host" desc:"DB host"`
		Port int    `json:"port" desc:"DB port"`
	}
	type Config struct {
		Port     int      `json:"port" desc:"HTTP port"`
		Database Database `json:"database" desc:"DB config"`
	}

	md, err := GenerateMarkdown[Config]()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Nested fields should appear with dot notation
	if !strings.Contains(md, "database.host") {
		t.Errorf("expected 'database.host' in markdown, got:\n%s", md)
	}
	// Parent struct (object type) should also appear
	if !strings.Contains(md, "| database |") {
		t.Errorf("expected 'database' row in markdown table, got:\n%s", md)
	}
}

func TestSchemaParseDefaultBool(t *testing.T) {
	type Config struct {
		Debug   bool    `json:"debug" default:"true"`
		Ratio   float32 `json:"ratio" default:"0.5"`
		Count   uint    `json:"count" default:"5"`
	}

	schema, err := GenerateSchema[Config]()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p := schema.Properties["debug"]; p == nil || p.Default != true {
		t.Errorf("expected debug default true, got %v", schema.Properties["debug"])
	}
}

func TestSchemaRequiredNestedSection(t *testing.T) {
	type Database struct {
		Host string `json:"host"`
	}
	type Config struct {
		Database Database `json:"database" validate:"required"`
	}

	schema, err := GenerateSchema[Config]()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// "database" should appear in the required list due to validate:"required" on the struct field
	found := false
	for _, req := range schema.Required {
		if req == "database" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'database' in required list, got: %v", schema.Required)
	}
}

func TestSchemaArrayOfInts(t *testing.T) {
	type Config struct {
		Ports []int    `json:"ports"`
		Flags []bool   `json:"flags"`
	}

	schema, err := GenerateSchema[Config]()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p := schema.Properties["ports"]; p == nil || p.Type != "array" || p.Items == nil || p.Items.Type != "integer" {
		t.Errorf("expected ports to be array of integer, got %+v", schema.Properties["ports"])
	}
	if p := schema.Properties["flags"]; p == nil || p.Type != "array" || p.Items == nil || p.Items.Type != "boolean" {
		t.Errorf("expected flags to be array of boolean, got %+v", schema.Properties["flags"])
	}
}

func TestSchemaMultipleValidationRules(t *testing.T) {
	type Config struct {
		Score int `json:"score" validate:"required,min=0,max=100"`
		Level string `json:"level" validate:"required,oneof=beginner,intermediate,advanced"`
	}

	schema, err := GenerateSchema[Config]()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// score should have min and max
	if p := schema.Properties["score"]; p == nil || p.Minimum == nil || p.Maximum == nil {
		t.Errorf("expected score to have min and max constraints")
	}

	// level should have enum values
	if p := schema.Properties["level"]; p == nil || len(p.Enum) != 3 {
		t.Errorf("expected level to have 3 enum values, got %v", schema.Properties["level"])
	}

	// both should be required
	if len(schema.Required) < 2 {
		t.Errorf("expected 2 required fields, got %d", len(schema.Required))
	}
}
