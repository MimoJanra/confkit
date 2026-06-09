package confkit

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		type cfg struct {
			Port int    `env:"DEFAULT_PORT" default:"8080"`
			Mode string `env:"DEFAULT_MODE" default:"dev"`
		}
		c, err := Load[cfg](FromEnv())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Port != 8080 || c.Mode != "dev" {
			t.Errorf("unexpected defaults: port=%d mode=%s", c.Port, c.Mode)
		}
	})

	t.Run("from_env", func(t *testing.T) {
		type cfg struct {
			Port int    `env:"TEST_PORT"`
			Host string `env:"TEST_HOST"`
		}
		t.Setenv("TEST_PORT", "3000")
		t.Setenv("TEST_HOST", "localhost")

		c, err := Load[cfg](FromEnv())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Port != 3000 || c.Host != "localhost" {
			t.Errorf("unexpected values: port=%d host=%s", c.Port, c.Host)
		}
	})

	t.Run("map_field", func(t *testing.T) {
		type cfg struct {
			Labels map[string]string `env:"MAP_LABELS"`
		}
		_ = os.Setenv("MAP_LABELS", "app=web,env=prod")
		t.Cleanup(func() { _ = os.Unsetenv("MAP_LABELS") })

		c, err := Load[cfg](FromEnv())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Labels["app"] != "web" || c.Labels["env"] != "prod" {
			t.Errorf("unexpected labels: %v", c.Labels)
		}
	})

	t.Run("duration_field", func(t *testing.T) {
		type cfg struct {
			Timeout string `env:"DUR_TIMEOUT" default:"30s"`
		}
		c, err := Load[cfg](FromEnv())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Timeout != "30s" {
			t.Errorf("expected '30s', got %q", c.Timeout)
		}
	})

	t.Run("error_source", func(t *testing.T) {
		type cfg struct {
			Host string `env:"ERR_HOST"`
		}
		_, err := Load[cfg](NewErrorSource(fmt.Errorf("source failure")))
		if err == nil {
			t.Fatal("expected error from error source")
		}
	})
}

func TestLoadContext(t *testing.T) {
	type cfg struct {
		Port int `env:"PORT" default:"9090"`
	}
	c, err := LoadContext[cfg](context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if c.Port != 9090 {
		t.Errorf("expected 9090, got %d", c.Port)
	}
}

func TestLoadWithOptionsContext(t *testing.T) {
	type cfg struct {
		Port int `env:"PORT" default:"1234"`
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c, err := LoadWithOptionsContext[cfg](ctx)
	if err != nil {
		t.Fatal(err)
	}
	if c.Port != 1234 {
		t.Errorf("expected 1234, got %d", c.Port)
	}
}

func TestWithContext(t *testing.T) {
	type cfg struct {
		X string `default:"hello"`
	}
	c, err := LoadWithOptions[cfg](WithContext(context.Background()))
	if err != nil {
		t.Fatal(err)
	}
	if c.X != "hello" {
		t.Errorf("got %q", c.X)
	}
}

func TestMustLoad(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		type cfg struct{ X int `default:"42"` }
		c := MustLoad[cfg]()
		if c.X != 42 {
			t.Errorf("expected 42, got %d", c.X)
		}
	})

	t.Run("panics_on_error", func(t *testing.T) {
		type cfg struct {
			X int `env:"MUST_LOAD_PORT" validate:"required"`
		}
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustLoad should panic on error")
			}
		}()
		MustLoad[cfg]()
	})

	t.Run("with_context", func(t *testing.T) {
		type cfg struct{ X int `default:"7"` }
		c := MustLoadContext[cfg](context.Background())
		if c.X != 7 {
			t.Errorf("expected 7, got %d", c.X)
		}
	})
}

func TestValidateOnly(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		type cfg struct {
			Port int    `env:"PORT" default:"8080" validate:"min=1,max=65535"`
			Host string `default:"localhost"`
		}
		c, err := ValidateOnly[cfg](context.Background())
		if err != nil {
			t.Fatalf("ValidateOnly should succeed: %v", err)
		}
		if c.Port != 8080 {
			t.Errorf("expected port 8080, got %d", c.Port)
		}
	})

	t.Run("skips_hooks_and_audit", func(t *testing.T) {
		type cfg struct{ Port int `default:"8080"` }
		hookCalled, auditCalled := false, false

		_, err := ValidateOnly[cfg](context.Background(),
			WithLoadHook(func(bool, time.Duration, int) { hookCalled = true }),
			WithAuditLogger(func([]AuditEntry) { auditCalled = true }),
		)
		if err != nil {
			t.Fatal(err)
		}
		if hookCalled {
			t.Error("ValidateOnly should not call LoadHookFunc")
		}
		if auditCalled {
			t.Error("ValidateOnly should not call AuditLogger")
		}
	})

	t.Run("failure", func(t *testing.T) {
		type cfg struct {
			Port int `env:"V1_PORT" validate:"min=1,max=65535"`
		}
		t.Setenv("V1_PORT", "99999")
		_, err := ValidateOnly[cfg](context.Background(), WithSource(FromEnv()))
		if err == nil {
			t.Fatal("expected validation error")
		}
	})
}

func TestAuditLogger(t *testing.T) {
	t.Run("on_validation_failure", func(t *testing.T) {
		type cfg struct {
			Port int `env:"AUDIT_FAIL_PORT" validate:"min=1"`
		}
		_ = os.Setenv("AUDIT_FAIL_PORT", "0")
		t.Cleanup(func() { _ = os.Unsetenv("AUDIT_FAIL_PORT") })

		var called bool
		_, err := LoadWithOptions[cfg](
			WithSource(FromEnv()),
			WithAuditLogger(func([]AuditEntry) { called = true }),
		)
		if err == nil {
			t.Fatal("expected validation error")
		}
		if !called {
			t.Error("expected audit logger to be called on failure")
		}
	})

	t.Run("on_source_error", func(t *testing.T) {
		type cfg struct{ X string `env:"AUDITX"` }
		var called bool
		_, _ = LoadWithOptions[cfg](
			WithSource(NewErrorSource(fmt.Errorf("src err"))),
			WithAuditLogger(func([]AuditEntry) { called = true }),
		)
		if !called {
			t.Error("expected audit logger to be called on source error")
		}
	})
}
