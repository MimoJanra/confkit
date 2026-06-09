package confkit_test

import (
	"os"
	"strings"
	"testing"

	confkit "github.com/MimoJanra/confkit"
)

func TestYAMLSource(t *testing.T) {
	t.Run("loads_fields", func(t *testing.T) {
		type cfg struct {
			Port int    `yaml:"port"`
			Host string `yaml:"host"`
		}
		c, err := confkit.Load[cfg](confkit.FromYAML("../testdata/config.yaml"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Port == 0 {
			t.Error("expected Port to be loaded from yaml")
		}
		if c.Host == "" {
			t.Error("expected Host to be loaded from yaml")
		}
	})

	t.Run("optional_missing", func(t *testing.T) {
		type cfg struct {
			Port int `yaml:"port" default:"8080"`
		}
		c, err := confkit.Load[cfg](confkit.FromYAMLOptional("/nonexistent/path.yaml"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Port != 8080 {
			t.Errorf("expected default 8080, got %d", c.Port)
		}
	})

	t.Run("missing_file_returns_error", func(t *testing.T) {
		type cfg struct {
			Host string `yaml:"host"`
		}
		_, err := confkit.Load[cfg](confkit.FromYAML("/nonexistent/path.yaml"))
		if err == nil {
			t.Fatal("expected error for missing required file")
		}
	})

	t.Run("nested_struct", func(t *testing.T) {
		type DB struct {
			Host string `yaml:"host"`
			Port int    `yaml:"port"`
		}
		type cfg struct {
			DB DB `yaml:"database"`
		}
		c, err := confkit.Load[cfg](confkit.FromYAML("../testdata/config.yaml"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.DB.Host == "" {
			t.Error("expected DB.Host to be loaded")
		}
	})
}

func TestJSONSource(t *testing.T) {
	t.Run("loads_fields", func(t *testing.T) {
		type cfg struct {
			Port int    `json:"port"`
			Host string `json:"host"`
		}
		c, err := confkit.Load[cfg](confkit.FromJSON("../testdata/config.json"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Port == 0 {
			t.Error("expected Port to be loaded from json")
		}
		if c.Host == "" {
			t.Error("expected Host to be loaded from json")
		}
	})

	t.Run("missing_file_error", func(t *testing.T) {
		type cfg struct {
			X string `json:"x"`
		}
		_, err := confkit.Load[cfg](confkit.FromJSON("/nonexistent.json"))
		if err == nil {
			t.Fatal("expected error for missing JSON file")
		}
	})

	t.Run("from_temp_file", func(t *testing.T) {
		type cfg struct {
			Name string `json:"name"`
			Val  int    `json:"val"`
		}
		tmpFile := writeTempJSON(t, `{"name":"test","val":42}`)
		defer func() { _ = os.Remove(tmpFile) }()

		c, err := confkit.Load[cfg](confkit.FromJSON(tmpFile))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Name != "test" || c.Val != 42 {
			t.Errorf("unexpected: %+v", c)
		}
	})
}

func TestTOMLSource(t *testing.T) {
	t.Run("loads_fields", func(t *testing.T) {
		type cfg struct {
			Port int    `toml:"port"`
			Host string `toml:"host"`
		}
		c, err := confkit.Load[cfg](confkit.FromTOML("../testdata/config.toml"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Port == 0 {
			t.Error("expected Port to be loaded from toml")
		}
		if c.Host == "" {
			t.Error("expected Host to be loaded from toml")
		}
	})

	t.Run("from_temp_file", func(t *testing.T) {
		type cfg struct {
			Name string `toml:"name"`
		}
		tmpFile := writeTempTOML(t, "name = \"myapp\"\n")
		defer func() { _ = os.Remove(tmpFile) }()

		c, err := confkit.Load[cfg](confkit.FromTOML(tmpFile))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Name != "myapp" {
			t.Errorf("expected 'myapp', got %q", c.Name)
		}
	})
}

func TestOverlay(t *testing.T) {
	t.Run("base_used_when_no_overlay", func(t *testing.T) {
		type cfg struct {
			Port int    `yaml:"port"`
			Host string `yaml:"host"`
		}
		src := confkit.FromOverlay(confkit.FromYAML("../testdata/config.yaml"), "nonexistent-env")
		c, err := confkit.Load[cfg](src)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Port == 0 {
			t.Error("expected Port from base file")
		}
	})

	t.Run("overlay_overrides_base", func(t *testing.T) {
		type cfg struct {
			Host string `yaml:"host"`
		}
		base := writeTempYAML(t, "host: base-host\n")
		defer func() { _ = os.Remove(base) }()

		overlayPath := confkit.OverlayPath(base, "prod")
		if err := os.WriteFile(overlayPath, []byte("host: prod-host\n"), 0644); err != nil {
			t.Fatalf("failed to write overlay: %v", err)
		}
		defer func() { _ = os.Remove(overlayPath) }()

		src := confkit.FromOverlay(confkit.FromYAML(base), "prod")
		c, err := confkit.Load[cfg](src)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Host != "prod-host" {
			t.Errorf("expected 'prod-host', got %q", c.Host)
		}
	})

	t.Run("overlay_path_format", func(t *testing.T) {
		p := confkit.OverlayPath("config.yaml", "staging")
		if p != "config.staging.yaml" {
			t.Errorf("expected 'config.staging.yaml', got %q", p)
		}
		p = confkit.OverlayPath("/etc/app/config.toml", "prod")
		if !strings.HasSuffix(p, "config.prod.toml") {
			t.Errorf("unexpected overlay path: %q", p)
		}
	})

	t.Run("non_file_source_passthrough", func(t *testing.T) {
		src := confkit.FromOverlay(confkit.FromEnv(), "prod")
		if src == nil {
			t.Fatal("expected non-nil source")
		}
	})

	t.Run("json_overlay", func(t *testing.T) {
		type cfg struct {
			Host string `json:"host"`
		}
		base := writeTempJSON(t, `{"host":"base-host"}`)
		defer func() { _ = os.Remove(base) }()

		overlayPath := confkit.OverlayPath(base, "prod")
		if err := os.WriteFile(overlayPath, []byte(`{"host":"prod-host"}`), 0644); err != nil {
			t.Fatalf("failed to write overlay: %v", err)
		}
		defer func() { _ = os.Remove(overlayPath) }()

		c, err := confkit.Load[cfg](confkit.FromOverlay(confkit.FromJSON(base), "prod"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Host != "prod-host" {
			t.Errorf("expected 'prod-host', got %q", c.Host)
		}
	})

	t.Run("toml_overlay", func(t *testing.T) {
		type cfg struct {
			Host string `toml:"host"`
		}
		base := writeTempTOML(t, "host = \"base-host\"\n")
		defer func() { _ = os.Remove(base) }()

		overlayPath := confkit.OverlayPath(base, "prod")
		if err := os.WriteFile(overlayPath, []byte("host = \"prod-host\"\n"), 0644); err != nil {
			t.Fatalf("failed to write overlay: %v", err)
		}
		defer func() { _ = os.Remove(overlayPath) }()

		c, err := confkit.Load[cfg](confkit.FromOverlay(confkit.FromTOML(base), "prod"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Host != "prod-host" {
			t.Errorf("expected 'prod-host', got %q", c.Host)
		}
	})
}

func TestInterpolation(t *testing.T) {
	t.Run("resolves_config_ref_in_yaml", func(t *testing.T) {
		type cfg struct {
			Host             string `yaml:"Host"`
			Port             int    `yaml:"Port"`
			ConnectionString string `yaml:"ConnectionString"`
		}
		c, err := confkit.Load[cfg](confkit.FromYAML("../testdata/interpolation.yaml"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(c.ConnectionString, "db.example.com") {
			t.Errorf("expected ConnectionString to contain resolved host, got: %s", c.ConnectionString)
		}
	})
}
