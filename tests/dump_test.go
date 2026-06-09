package confkit_test

import (
	"encoding/json"
	"strings"
	"testing"

	confkit "github.com/MimoJanra/confkit"
)

func TestDump(t *testing.T) {
	t.Run("json_default", func(t *testing.T) {
		type cfg struct {
			Port int    `json:"port"`
			Host string `json:"host"`
		}
		c := cfg{Port: 8080, Host: "localhost"}
		data, err := confkit.Dump(c)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var m map[string]any
		if err := json.Unmarshal(data, &m); err != nil {
			t.Fatalf("output is not valid JSON: %v", err)
		}
		if m["port"] != float64(8080) {
			t.Errorf("expected port 8080, got %v", m["port"])
		}
		if m["host"] != "localhost" {
			t.Errorf("expected host localhost, got %v", m["host"])
		}
	})

	t.Run("secret_redacted_by_default", func(t *testing.T) {
		type cfg struct {
			User     string `json:"user"`
			Password string `json:"password" secret:"true"`
		}
		c := cfg{User: "admin", Password: "s3cr3t"}
		data, err := confkit.Dump(c)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		s := string(data)
		if strings.Contains(s, "s3cr3t") {
			t.Errorf("secret should be redacted: %s", s)
		}
		if !strings.Contains(s, "***REDACTED***") {
			t.Errorf("expected REDACTED marker: %s", s)
		}
	})

	t.Run("secret_exposed_when_redact_false", func(t *testing.T) {
		type cfg struct {
			Password string `json:"password" secret:"true"`
		}
		c := cfg{Password: "s3cr3t"}
		data, err := confkit.Dump(c, confkit.WithDumpRedactSecrets(false))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(string(data), "s3cr3t") {
			t.Errorf("expected exposed secret when redact=false: %s", data)
		}
	})

	t.Run("nested_struct", func(t *testing.T) {
		type DB struct {
			Host string `json:"host"`
		}
		type cfg struct {
			DB DB `json:"db"`
		}
		c := cfg{DB: DB{Host: "db.internal"}}
		data, err := confkit.Dump(c)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(string(data), "db.internal") {
			t.Errorf("expected nested value in dump: %s", data)
		}
	})

	t.Run("slice_field", func(t *testing.T) {
		type cfg struct {
			Tags []string `json:"tags"`
		}
		c := cfg{Tags: []string{"a", "b", "c"}}
		data, err := confkit.Dump(c)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(string(data), `"a"`) {
			t.Errorf("expected slice values in dump: %s", data)
		}
	})

	t.Run("map_field", func(t *testing.T) {
		type cfg struct {
			Labels map[string]string `json:"labels"`
		}
		c := cfg{Labels: map[string]string{"env": "prod"}}
		data, err := confkit.Dump(c)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(string(data), "prod") {
			t.Errorf("expected map value in dump: %s", data)
		}
	})
}

func TestDumpYAML(t *testing.T) {
	t.Run("yaml_format", func(t *testing.T) {
		type cfg struct {
			Port int    `yaml:"port"`
			Host string `yaml:"host"`
		}
		c := cfg{Port: 9090, Host: "example.com"}
		data, err := confkit.DumpYAML(c)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		s := string(data)
		if !strings.Contains(s, "port:") {
			t.Errorf("expected yaml port key: %s", s)
		}
		if !strings.Contains(s, "example.com") {
			t.Errorf("expected host value: %s", s)
		}
	})

	t.Run("secret_redacted", func(t *testing.T) {
		type cfg struct {
			Token string `yaml:"token" secret:"true"`
		}
		data, err := confkit.DumpYAML(cfg{Token: "secret-token"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Contains(string(data), "secret-token") {
			t.Errorf("secret leaked in YAML dump: %s", data)
		}
	})
}

func TestDumpString(t *testing.T) {
	t.Run("returns_string", func(t *testing.T) {
		type cfg struct{ X int `json:"x"` }
		s := confkit.DumpString(cfg{X: 42})
		if !strings.Contains(s, "42") {
			t.Errorf("expected 42 in dump string: %s", s)
		}
	})
}
