package confkit

import (
	"os"
	"testing"
)

func TestFromJSONFiles(t *testing.T) {
	t.Run("merges_files", func(t *testing.T) {
		base := t.TempDir() + "/base.json"
		override := t.TempDir() + "/override.json"

		_ = os.WriteFile(base, []byte(`{"host":"localhost","port":5432}`), 0644)
		_ = os.WriteFile(override, []byte(`{"host":"db.prod.internal"}`), 0644)

		type cfg struct {
			Host string `json:"host"`
			Port int    `json:"port"`
		}
		c, err := Load[cfg](FromJSONFiles(base, override))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Host != "db.prod.internal" {
			t.Errorf("expected db.prod.internal, got %s", c.Host)
		}
		if c.Port != 5432 {
			t.Errorf("expected 5432, got %d", c.Port)
		}
	})

	t.Run("single_file", func(t *testing.T) {
		f := t.TempDir() + "/cfg.json"
		_ = os.WriteFile(f, []byte(`{"host":"myhost"}`), 0644)

		type cfg struct {
			Host string `json:"host"`
		}
		c, err := Load[cfg](FromJSONFiles(f))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Host != "myhost" {
			t.Errorf("expected 'myhost', got %q", c.Host)
		}
	})

	t.Run("missing_file_error", func(t *testing.T) {
		type cfg struct{ X string `json:"x"` }
		src := FromJSONFiles("/nonexistent.json")
		_, err := Load[cfg](src)
		if err == nil {
			t.Fatal("expected error for missing file")
		}
	})
}

func TestFromTOMLFiles(t *testing.T) {
	t.Run("merges_files", func(t *testing.T) {
		base := t.TempDir() + "/base.toml"
		override := t.TempDir() + "/override.toml"

		_ = os.WriteFile(base, []byte("host = \"localhost\"\nport = 5432\n"), 0644)
		_ = os.WriteFile(override, []byte("host = \"db.prod.internal\"\n"), 0644)

		type cfg struct {
			Host string `toml:"host"`
			Port int    `toml:"port"`
		}
		c, err := Load[cfg](FromTOMLFiles(base, override))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Host != "db.prod.internal" {
			t.Errorf("expected db.prod.internal, got %s", c.Host)
		}
		if c.Port != 5432 {
			t.Errorf("expected 5432, got %d", c.Port)
		}
	})

	t.Run("single_file", func(t *testing.T) {
		f := t.TempDir() + "/cfg.toml"
		_ = os.WriteFile(f, []byte("host = \"toml-host\"\n"), 0644)

		type cfg struct {
			Host string `toml:"host"`
		}
		c, err := Load[cfg](FromTOMLFiles(f))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Host != "toml-host" {
			t.Errorf("expected 'toml-host', got %q", c.Host)
		}
	})

	t.Run("missing_file_error", func(t *testing.T) {
		type cfg struct{ X string `toml:"x"` }
		src := FromTOMLFiles("/nonexistent.toml")
		_, err := Load[cfg](src)
		if err == nil {
			t.Fatal("expected error for missing file")
		}
	})
}

func TestFromYAMLFilesDeepMerge(t *testing.T) {
	t.Run("deep_merge_nested", func(t *testing.T) {
		base := t.TempDir() + "/base.yaml"
		override := t.TempDir() + "/override.yaml"

		_ = os.WriteFile(base, []byte("database:\n  host: localhost\n  port: 5432\n"), 0644)
		_ = os.WriteFile(override, []byte("database:\n  host: db.prod.internal\n"), 0644)

		type DB struct {
			Host string `yaml:"host"`
			Port int    `yaml:"port"`
		}
		type cfg struct {
			DB DB `yaml:"database"`
		}
		c, err := Load[cfg](FromYAMLFiles(base, override))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.DB.Host != "db.prod.internal" {
			t.Errorf("expected db.prod.internal, got %s", c.DB.Host)
		}
		if c.DB.Port != 5432 {
			t.Errorf("expected port 5432 preserved from base, got %d", c.DB.Port)
		}
	})
}
