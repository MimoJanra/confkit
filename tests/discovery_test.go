package confkit_test

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	confkit "github.com/MimoJanra/confkit"
)

func TestDefaultSearchDirs(t *testing.T) {
	dirs := confkit.DefaultSearchDirs("myapp")
	if len(dirs) < 3 {
		t.Errorf("expected at least 3 dirs, got %d", len(dirs))
	}
	if !slices.Contains(dirs, "./") {
		t.Error("expected './' in default search dirs")
	}
}

func TestFindFile(t *testing.T) {
	t.Run("finds_by_exact_path", func(t *testing.T) {
		p, ok := confkit.FindFile("config.yaml", "../testdata/")
		if !ok {
			t.Fatal("expected to find config.yaml in testdata/")
		}
		if !strings.HasSuffix(p, "config.yaml") {
			t.Errorf("unexpected path: %s", p)
		}
	})

	t.Run("finds_by_extension_probe", func(t *testing.T) {
		p, ok := confkit.FindFile("config", "../testdata/")
		if !ok {
			t.Fatal("expected to find config without extension in testdata/")
		}
		ext := filepath.Ext(p)
		if ext != ".yaml" && ext != ".yml" && ext != ".json" && ext != ".toml" {
			t.Errorf("expected known extension, got %q for %q", ext, p)
		}
	})

	t.Run("not_found", func(t *testing.T) {
		_, ok := confkit.FindFile("nonexistent", "/tmp/definitely-does-not-exist-9999/")
		if ok {
			t.Error("expected not found")
		}
	})

	t.Run("multiple_dirs", func(t *testing.T) {
		tmp := t.TempDir()
		f := filepath.Join(tmp, "app.yaml")
		_ = os.WriteFile(f, []byte("x: 1"), 0644)

		p, ok := confkit.FindFile("app.yaml", "/nonexistent/", tmp)
		if !ok {
			t.Fatal("expected to find file in second dir")
		}
		if p != filepath.Clean(f) {
			t.Errorf("unexpected path: %s", p)
		}
	})
}

func TestFindSource(t *testing.T) {
	t.Run("returns_yaml_source", func(t *testing.T) {
		src := confkit.FindSource("config.yaml", "../testdata/")
		if src == nil {
			t.Fatal("expected non-nil source")
		}
		if src.Name() != "yaml" {
			t.Errorf("expected yaml source, got %q", src.Name())
		}
	})

	t.Run("not_found_returns_error_source", func(t *testing.T) {
		type cfg struct {
			X string `yaml:"x"`
		}
		src := confkit.FindSource("nonexistent", "/tmp/definitely-not-there-9999/")
		_, err := confkit.Load[cfg](src)
		if err == nil {
			t.Fatal("expected error from not-found source")
		}
		if !errors.Is(err, confkit.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("returns_json_source", func(t *testing.T) {
		src := confkit.FindSource("config.json", "../testdata/")
		if src == nil {
			t.Fatal("expected non-nil source")
		}
		if src.Name() != "json" {
			t.Errorf("expected json source, got %q", src.Name())
		}
	})

	t.Run("returns_toml_source", func(t *testing.T) {
		src := confkit.FindSource("config.toml", "../testdata/")
		if src == nil {
			t.Fatal("expected non-nil source")
		}
		if src.Name() != "toml" {
			t.Errorf("expected toml source, got %q", src.Name())
		}
	})
}

func TestOverlayPath(t *testing.T) {
	cases := []struct {
		base string
		env  string
		want string
	}{
		{"config.yaml", "prod", "config.prod.yaml"},
		{"config.json", "staging", "config.staging.json"},
		{"config.toml", "dev", "config.dev.toml"},
		{"/etc/app/config.yaml", "prod", "/etc/app/config.prod.yaml"},
	}
	for _, tc := range cases {
		got := confkit.OverlayPath(tc.base, tc.env)
		if got != tc.want {
			t.Errorf("OverlayPath(%q, %q) = %q, want %q", tc.base, tc.env, got, tc.want)
		}
	}
}

func TestFindFileYmlExtension(t *testing.T) {
	dir := t.TempDir()

	t.Run("finds_yml", func(t *testing.T) {
		if err := os.WriteFile(filepath.Join(dir, "app.yml"), []byte("port: 1\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		got, ok := confkit.FindFile("app", dir)
		if !ok {
			t.Fatal("expected to find app.yml")
		}
		if filepath.Ext(got) != ".yml" {
			t.Fatalf("got %q, want a .yml file", got)
		}
	})

	t.Run("yaml_wins_over_yml", func(t *testing.T) {
		if err := os.WriteFile(filepath.Join(dir, "app.yaml"), []byte("port: 2\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		got, ok := confkit.FindFile("app", dir)
		if !ok {
			t.Fatal("expected to find a config file")
		}
		if filepath.Ext(got) != ".yaml" {
			t.Fatalf("got %q, want .yaml to take precedence", got)
		}
	})
}

func TestFindSourceLoadsYml(t *testing.T) {
	type Cfg struct {
		Port int `yaml:"port"`
	}

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "svc.yml"), []byte("port: 9090\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := confkit.Load[Cfg](confkit.FindSource("svc", dir))
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Port != 9090 {
		t.Fatalf("got %d, want 9090", cfg.Port)
	}
}
