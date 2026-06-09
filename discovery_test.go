package confkit

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestDefaultSearchDirs(t *testing.T) {
	dirs := DefaultSearchDirs("myapp")
	if len(dirs) < 3 {
		t.Errorf("expected at least 3 dirs, got %d", len(dirs))
	}
	if !slices.Contains(dirs, "./") {
		t.Error("expected './' in default search dirs")
	}
}

func TestFindFile(t *testing.T) {
	t.Run("finds_by_exact_path", func(t *testing.T) {
		p, ok := FindFile("config.yaml", "testdata/")
		if !ok {
			t.Fatal("expected to find config.yaml in testdata/")
		}
		if !strings.HasSuffix(p, "config.yaml") {
			t.Errorf("unexpected path: %s", p)
		}
	})

	t.Run("finds_by_extension_probe", func(t *testing.T) {
		p, ok := FindFile("config", "testdata/")
		if !ok {
			t.Fatal("expected to find config without extension in testdata/")
		}
		ext := filepath.Ext(p)
		if ext != ".yaml" && ext != ".json" && ext != ".toml" {
			t.Errorf("expected known extension, got %q for %q", ext, p)
		}
	})

	t.Run("not_found", func(t *testing.T) {
		_, ok := FindFile("nonexistent", "/tmp/definitely-does-not-exist-9999/")
		if ok {
			t.Error("expected not found")
		}
	})

	t.Run("multiple_dirs", func(t *testing.T) {
		tmp := t.TempDir()
		f := filepath.Join(tmp, "app.yaml")
		_ = os.WriteFile(f, []byte("x: 1"), 0644)

		p, ok := FindFile("app.yaml", "/nonexistent/", tmp)
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
		src := FindSource("config.yaml", "testdata/")
		if src == nil {
			t.Fatal("expected non-nil source")
		}
		if src.Name() != "yaml" {
			t.Errorf("expected yaml source, got %q", src.Name())
		}
	})

	t.Run("not_found_returns_error_source", func(t *testing.T) {
		type cfg struct{ X string `yaml:"x"` }
		src := FindSource("nonexistent", "/tmp/definitely-not-there-9999/")
		_, err := Load[cfg](src)
		if err == nil {
			t.Fatal("expected error from not-found source")
		}
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("returns_json_source", func(t *testing.T) {
		src := FindSource("config.json", "testdata/")
		if src == nil {
			t.Fatal("expected non-nil source")
		}
		if src.Name() != "json" {
			t.Errorf("expected json source, got %q", src.Name())
		}
	})

	t.Run("returns_toml_source", func(t *testing.T) {
		src := FindSource("config.toml", "testdata/")
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
		got := OverlayPath(tc.base, tc.env)
		if got != tc.want {
			t.Errorf("OverlayPath(%q, %q) = %q, want %q", tc.base, tc.env, got, tc.want)
		}
	}
}
