package confkit

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestDumpConfig(t *testing.T) {
	type config struct {
		Port int `json:"port"`
	}

	cfg := config{Port: 8080}
	fields := []FieldInfo{
		{Path: "Port", IsSecret: false},
	}

	data, err := DumpConfig(cfg, fields)
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
	fields := []FieldInfo{
		{Path: "Password", IsSecret: true},
	}

	data, err := DumpConfig(cfg, fields)
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
	type dbConfig struct {
		Host string
	}

	type config struct {
		Database dbConfig
	}

	cfg := config{Database: dbConfig{Host: "localhost"}}
	fields := []FieldInfo{
		{Path: "Database.Host", IsSecret: false},
	}

	data, err := DumpConfig(cfg, fields)
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
	log := LogLoadStart(sources)

	if !strings.Contains(log, `"event":"config_load_start"`) {
		t.Errorf("Expected event field, got %s", log)
	}
	if !strings.Contains(log, `"sources":["yaml","env"]`) {
		t.Errorf("Expected sources array, got %s", log)
	}
	if !strings.Contains(log, `"timestamp"`) {
		t.Errorf("Expected timestamp field, got %s", log)
	}
}

func TestLogLoadComplete(t *testing.T) {
	log := LogLoadComplete(100*time.Millisecond, 5, 0)

	if !strings.Contains(log, `"event":"config_load_complete"`) {
		t.Errorf("Expected event field, got %s", log)
	}
	if !strings.Contains(log, `"duration_ms":100`) {
		t.Errorf("Expected duration_ms=100, got %s", log)
	}
	if !strings.Contains(log, `"fields_loaded":5`) {
		t.Errorf("Expected fields_loaded=5, got %s", log)
	}
	if !strings.Contains(log, `"validation_errors":0`) {
		t.Errorf("Expected validation_errors=0, got %s", log)
	}
}

func TestParseFieldPath(t *testing.T) {
	tests := []struct {
		path     string
		expected []string
	}{
		{"Port", []string{"Port"}},
		{"Database.Host", []string{"Database", "Host"}},
		{"Database.Connection.Host", []string{"Database", "Connection", "Host"}},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := parseFieldPath(tt.path)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d parts, got %d", len(tt.expected), len(result))
				return
			}
			for i, part := range result {
				if part != tt.expected[i] {
					t.Errorf("Expected %q, got %q at index %d", tt.expected[i], part, i)
				}
			}
		})
	}
}

func TestGetFieldValue(t *testing.T) {
	type config struct {
		Port int
	}

	cfg := config{Port: 8080}
	val := reflect.ValueOf(cfg)

	result := getFieldValue(val, "Port")
	if result != 8080 {
		t.Errorf("Expected 8080, got %v", result)
	}
}

func TestLoadMetricsType(t *testing.T) {
	metrics := LoadMetrics{
		TotalTime: 100 * time.Millisecond,
		SourceTimes: map[string]time.Duration{
			"yaml": 50 * time.Millisecond,
			"env":  30 * time.Millisecond,
		},
		ValidationTime: 20 * time.Millisecond,
		ErrorCount:     0,
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
	fields := []FieldInfo{
		{Path: "Port", IsSecret: false},
		{Path: "Host", IsSecret: false},
		{Path: "Password", IsSecret: true},
	}

	data, err := DumpConfig(cfg, fields)
	if err != nil {
		t.Fatalf("DumpConfig failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if result["Port"] != float64(8080) {
		t.Errorf("Port mismatch")
	}
	if result["Host"] != "localhost" {
		t.Errorf("Host mismatch")
	}
	if result["Password"] != "***REDACTED***" {
		t.Errorf("Password not redacted")
	}
}
