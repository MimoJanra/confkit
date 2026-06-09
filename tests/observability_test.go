package confkit_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	confkit "github.com/MimoJanra/confkit"
)

func TestDumpConfig(t *testing.T) {
	type config struct {
		Port int `json:"port"`
	}
	cfg := config{Port: 8080}
	fields := []confkit.FieldInfo{{Path: "Port", IsSecret: false}}

	data, err := confkit.DumpConfig(cfg, fields)
	if err != nil {
		t.Fatalf("DumpConfig failed: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	if result["Port"] != float64(8080) {
		t.Errorf("Expected Port=8080, got %v", result["Port"])
	}
}

func TestDumpConfigWithSecret(t *testing.T) {
	type config struct {
		Password string `json:"password"`
	}
	cfg := config{Password: "secret123"}
	fields := []confkit.FieldInfo{{Path: "Password", IsSecret: true}}

	data, err := confkit.DumpConfig(cfg, fields)
	if err != nil {
		t.Fatalf("DumpConfig failed: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	if result["Password"] != "***REDACTED***" {
		t.Errorf("Expected password to be redacted, got %v", result["Password"])
	}
}

func TestDumpConfigWithNestedStruct(t *testing.T) {
	type dbConfig struct{ Host string }
	type config struct{ Database dbConfig }

	cfg := config{Database: dbConfig{Host: "localhost"}}
	fields := []confkit.FieldInfo{{Path: "Database.Host", IsSecret: false}}

	data, err := confkit.DumpConfig(cfg, fields)
	if err != nil {
		t.Fatalf("DumpConfig failed: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	if result["Database.Host"] != "localhost" {
		t.Errorf("Expected Database.Host=localhost, got %v", result["Database.Host"])
	}
}

func TestLogLoadStart(t *testing.T) {
	sources := []string{"yaml", "env"}
	logMsg := confkit.LogLoadStart(sources)
	if !stringContains(logMsg, "config_load_start") {
		t.Error("Log should contain config_load_start event")
	}
	if !stringContains(logMsg, "yaml") {
		t.Error("Log should contain yaml source")
	}
}

func TestLogLoadComplete(t *testing.T) {
	logMsg := confkit.LogLoadComplete(50*time.Millisecond, 5, 0)
	if !stringContains(logMsg, "config_load_complete") {
		t.Error("Log should contain config_load_complete event")
	}
	if !stringContains(logMsg, "fields_loaded") {
		t.Error("Log should contain fields_loaded count")
	}
}

func TestLoadMetricsType(t *testing.T) {
	metrics := confkit.LoadMetrics{
		TotalTime: 100 * time.Millisecond,
		SourceTimes: map[string]time.Duration{
			"yaml": 50 * time.Millisecond,
			"env":  30 * time.Millisecond,
		},
	}
	if metrics.TotalTime != 100*time.Millisecond {
		t.Errorf("Expected TotalTime=100ms, got %v", metrics.TotalTime)
	}
	if len(metrics.SourceTimes) != 2 {
		t.Errorf("Expected 2 sources, got %d", len(metrics.SourceTimes))
	}
}

func TestDumpConfigMultipleFields(t *testing.T) {
	type config struct {
		Port     int
		Host     string
		Password string
	}
	cfg := config{Port: 8080, Host: "localhost", Password: "secret"}
	fields := []confkit.FieldInfo{
		{Path: "Port", IsSecret: false},
		{Path: "Host", IsSecret: false},
		{Path: "Password", IsSecret: true},
	}
	data, err := confkit.DumpConfig(cfg, fields)
	if err != nil {
		t.Fatalf("DumpConfig failed: %v", err)
	}
	dumpStr := string(data)
	if !strings.Contains(dumpStr, "Port") {
		t.Error("Port should be in dump")
	}
	if !strings.Contains(dumpStr, "***REDACTED***") {
		t.Error("Password should be redacted")
	}
}
