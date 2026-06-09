package confkit

import (
	"testing"
)

func TestEnvPrefix(t *testing.T) {
	t.Run("simple_prefix", func(t *testing.T) {
		type cfg struct {
			Host string `env:"HOST" prefix:"APP_"`
			Port int    `env:"PORT" prefix:"APP_"`
		}
		t.Setenv("APP_HOST", "myhost")
		t.Setenv("APP_PORT", "9090")

		c, err := Load[cfg](FromEnv())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Host != "myhost" {
			t.Errorf("expected 'myhost', got %q", c.Host)
		}
		if c.Port != 9090 {
			t.Errorf("expected 9090, got %d", c.Port)
		}
	})

	t.Run("hierarchical_prefix", func(t *testing.T) {
		type DB struct {
			Host string `env:"HOST" prefix:"DB_"`
		}
		type cfg struct {
			DB DB `prefix:"APP_"`
		}
		t.Setenv("APP_DB_HOST", "db.internal")

		c, err := Load[cfg](FromEnv())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.DB.Host != "db.internal" {
			t.Errorf("expected 'db.internal', got %q", c.DB.Host)
		}
	})

	t.Run("with_default", func(t *testing.T) {
		type cfg struct {
			Port int `env:"PORT" prefix:"SVC_" default:"8080"`
		}
		c, err := Load[cfg](FromEnv())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Port != 8080 {
			t.Errorf("expected default 8080, got %d", c.Port)
		}
	})

	t.Run("override_default", func(t *testing.T) {
		type cfg struct {
			Port int `env:"PORT" prefix:"OVERRIDE_" default:"8080"`
		}
		t.Setenv("OVERRIDE_PORT", "9000")

		c, err := Load[cfg](FromEnv())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Port != 9000 {
			t.Errorf("expected 9000, got %d", c.Port)
		}
	})

	t.Run("no_prefix", func(t *testing.T) {
		type cfg struct {
			Host string `env:"NOPREFIX_HOST"`
		}
		t.Setenv("NOPREFIX_HOST", "bare-host")

		c, err := Load[cfg](FromEnv())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Host != "bare-host" {
			t.Errorf("expected 'bare-host', got %q", c.Host)
		}
	})
}

func TestEnvSourceName(t *testing.T) {
	src := FromEnv()
	if src.Name() != "env" {
		t.Errorf("expected 'env', got %q", src.Name())
	}
}
