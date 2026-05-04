package confkit

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestScanFields(t *testing.T) {
	type testConfig struct {
		Port int    `env:"PORT" default:"8080"`
		Mode string `env:"MODE" default:"dev"`
	}

	config := testConfig{}
	fields := ScanFields(config)

	if len(fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(fields))
	}

	// Check Port field
	portField := fields[0]
	if portField.Name != "Port" {
		t.Errorf("expected field name Port, got %s", portField.Name)
	}
	if portField.Tags["env"] != "PORT" {
		t.Errorf("expected env tag PORT, got %s", portField.Tags["env"])
	}
	if portField.Tags["default"] != "8080" {
		t.Errorf("expected default tag 8080, got %s", portField.Tags["default"])
	}

	// Check Mode field
	modeField := fields[1]
	if modeField.Name != "Mode" {
		t.Errorf("expected field name Mode, got %s", modeField.Name)
	}
}

func TestParseString(t *testing.T) {
	p := NewParser()
	result, err := p.Parse("hello", reflect.TypeOf(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "hello" {
		t.Errorf("expected 'hello', got %q", result)
	}
}

func TestParseBool(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
		wantErr  bool
	}{
		{"true", true, false},
		{"false", false, false},
		{"1", true, false},
		{"0", false, false},
		{"yes", true, false},
		{"no", false, false},
		{"invalid", false, true},
	}

	p := NewParser()
	for _, tc := range tests {
		result, err := p.Parse(tc.input, reflect.TypeOf(false))
		if (err != nil) != tc.wantErr {
			t.Errorf("Parse(%q): got error %v, wantErr %v", tc.input, err, tc.wantErr)
			continue
		}
		if err == nil && result != tc.expected {
			t.Errorf("Parse(%q): expected %v, got %v", tc.input, tc.expected, result)
		}
	}
}

func TestParseInt(t *testing.T) {
	p := NewParser()

	tests := []struct {
		input    string
		typ      reflect.Type
		expected any
		wantErr  bool
	}{
		{"42", reflect.TypeOf(int(0)), int(42), false},
		{"255", reflect.TypeOf(uint8(0)), uint8(255), false},
		{"256", reflect.TypeOf(uint8(0)), nil, true}, // overflow
		{"-128", reflect.TypeOf(int8(0)), int8(-128), false},
		{"100", reflect.TypeOf(int32(0)), int32(100), false},
	}

	for _, tc := range tests {
		result, err := p.Parse(tc.input, tc.typ)
		if (err != nil) != tc.wantErr {
			t.Errorf("Parse(%q): got error %v, wantErr %v", tc.input, err, tc.wantErr)
			continue
		}
		if err == nil && result != tc.expected {
			t.Errorf("Parse(%q): expected %v, got %v", tc.input, tc.expected, result)
		}
	}
}

func TestParseFloat(t *testing.T) {
	p := NewParser()

	result, err := p.Parse("3.14", reflect.TypeOf(float64(0)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	f := result.(float64)
	if f < 3.13 || f > 3.15 {
		t.Errorf("expected ~3.14, got %v", f)
	}
}

func TestParseDuration(t *testing.T) {
	p := NewParser()

	result, err := p.Parse("5s", reflect.TypeOf(time.Duration(0)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	d := result.(time.Duration)
	if d != 5*time.Second {
		t.Errorf("expected 5s, got %v", d)
	}
}

func TestParseTime(t *testing.T) {
	p := NewParser()

	result, err := p.Parse("2026-01-01T00:00:00Z", reflect.TypeOf(time.Time{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tm := result.(time.Time)
	if tm.Year() != 2026 || tm.Month() != time.January || tm.Day() != 1 {
		t.Errorf("unexpected date: %v", tm)
	}
}

func TestParseSlice(t *testing.T) {
	p := NewParser()

	// []string
	result, err := p.Parse("a,b,c", reflect.TypeOf([]string{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	slice := result.([]string)
	if len(slice) != 3 || slice[0] != "a" || slice[1] != "b" || slice[2] != "c" {
		t.Errorf("expected [a b c], got %v", slice)
	}

	// []int
	result, err = p.Parse("1,2,3", reflect.TypeOf([]int{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intSlice := result.([]int)
	if len(intSlice) != 3 || intSlice[0] != 1 || intSlice[1] != 2 || intSlice[2] != 3 {
		t.Errorf("expected [1 2 3], got %v", intSlice)
	}
}

func TestLoadWithDefaults(t *testing.T) {
	type Config struct {
		Port int    `env:"DEFAULT_PORT" default:"8080"`
		Mode string `env:"DEFAULT_MODE" default:"dev"`
	}

	// Don't set env vars, so defaults apply

	cfg, err := Load[Config](FromEnv())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != 8080 {
		t.Errorf("expected Port 8080, got %d", cfg.Port)
	}
	if cfg.Mode != "dev" {
		t.Errorf("expected Mode 'dev', got %q", cfg.Mode)
	}
}

func TestLoadFromEnv(t *testing.T) {
	type Config struct {
		Port int    `env:"TEST_PORT"`
		Host string `env:"TEST_HOST"`
	}

	t.Setenv("TEST_PORT", "3000")
	t.Setenv("TEST_HOST", "localhost")

	cfg, err := Load[Config](FromEnv())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != 3000 {
		t.Errorf("expected Port 3000, got %d", cfg.Port)
	}
	if cfg.Host != "localhost" {
		t.Errorf("expected Host 'localhost', got %q", cfg.Host)
	}
}

func TestLoadErrorReportParsing(t *testing.T) {
	type Config struct {
		Port int `env:"BAD_PORT"`
	}

	t.Setenv("BAD_PORT", "not-a-number")

	_, err := Load[Config](FromEnv())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	report, ok := err.(*ErrorReport)
	if !ok {
		t.Fatalf("expected ErrorReport, got %T", err)
	}

	if len(report.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(report.Errors))
	}

	fieldErr := report.Errors[0]
	if fieldErr.Kind != ErrorKindParse {
		t.Errorf("expected ErrorKindParse, got %s", fieldErr.Kind)
	}
	if fieldErr.Path != "Port" {
		t.Errorf("expected Path 'Port', got %q", fieldErr.Path)
	}
}

func TestExplainError(t *testing.T) {
	type Config struct {
		API string `env:"API_KEY" secret:"true"`
	}

	t.Setenv("API_KEY", "super-secret-key")

	_, err := Load[Config](FromEnv())
	if err == nil {
		// Manually create an error to test
		report := &ErrorReport{}
		report.AddError(FieldError{
			Path:    "API",
			Source:  "env API_KEY",
			Kind:    ErrorKindParse,
			Message: "invalid value",
			Value:   "super-secret-key",
			Secret:  true,
		})
		err = report
	}

	explained := Explain(err)
	if !strings.Contains(explained, "<redacted>") {
		t.Errorf("expected secret to be redacted, got: %s", explained)
	}
}

func TestLoadFromYAML(t *testing.T) {
	type Config struct {
		Port int    `yaml:"port"`
		Host string `yaml:"host"`
		Mode string `yaml:"mode"`
	}

	cfg, err := Load[Config](FromYAML("testdata/config.yaml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != 3000 {
		t.Errorf("expected Port 3000, got %d", cfg.Port)
	}
	if cfg.Host != "localhost" {
		t.Errorf("expected Host 'localhost', got %q", cfg.Host)
	}
	if cfg.Mode != "production" {
		t.Errorf("expected Mode 'production', got %q", cfg.Mode)
	}
}

func TestLoadFromJSON(t *testing.T) {
	type Config struct {
		Port int    `json:"port"`
		Host string `json:"host"`
		Mode string `json:"mode"`
	}

	cfg, err := Load[Config](FromJSON("testdata/config.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != 2000 {
		t.Errorf("expected Port 2000, got %d", cfg.Port)
	}
	if cfg.Host != "0.0.0.0" {
		t.Errorf("expected Host '0.0.0.0', got %q", cfg.Host)
	}
	if cfg.Mode != "development" {
		t.Errorf("expected Mode 'development', got %q", cfg.Mode)
	}
}

func TestLoadMultipleSources(t *testing.T) {
	type Config struct {
		Port int    `yaml:"port" env:"PORT" default:"8080"`
		Host string `yaml:"host" env:"HOST" default:"localhost"`
		Mode string `yaml:"mode" env:"MODE" default:"dev"`
	}

	t.Setenv("MODE", "test")

	cfg, err := Load[Config](
		FromYAML("testdata/config.yaml"),
		FromEnv(),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// YAML provides port and host, env overrides mode
	if cfg.Port != 3000 {
		t.Errorf("expected Port 3000 from YAML, got %d", cfg.Port)
	}
	if cfg.Host != "localhost" {
		t.Errorf("expected Host 'localhost' from YAML, got %q", cfg.Host)
	}
	if cfg.Mode != "test" {
		t.Errorf("expected Mode 'test' from env override, got %q", cfg.Mode)
	}
}
