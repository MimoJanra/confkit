package confkit_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	confkit "github.com/MimoJanra/confkit"
)

func TestLoad(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		type cfg struct {
			Port int    `env:"DEFAULT_PORT" default:"8080"`
			Mode string `env:"DEFAULT_MODE" default:"dev"`
		}
		c, err := confkit.Load[cfg](confkit.FromEnv())
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

		c, err := confkit.Load[cfg](confkit.FromEnv())
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

		c, err := confkit.Load[cfg](confkit.FromEnv())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Labels["app"] != "web" || c.Labels["env"] != "prod" {
			t.Errorf("unexpected labels: %v", c.Labels)
		}
	})

	t.Run("error_source", func(t *testing.T) {
		type cfg struct {
			Host string `env:"ERR_HOST"`
		}
		_, err := confkit.Load[cfg](confkit.NewErrorSource(fmt.Errorf("source failure")))
		if err == nil {
			t.Fatal("expected error from error source")
		}
	})
}

func TestLoadContext(t *testing.T) {
	type cfg struct {
		Port int `env:"PORT" default:"9090"`
	}
	c, err := confkit.LoadContext[cfg](context.Background())
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
	c, err := confkit.LoadWithOptionsContext[cfg](ctx)
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
	c, err := confkit.LoadWithOptions[cfg](confkit.WithContext(context.Background()))
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
		c := confkit.MustLoad[cfg]()
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
		confkit.MustLoad[cfg]()
	})

	t.Run("with_context", func(t *testing.T) {
		type cfg struct{ X int `default:"7"` }
		c := confkit.MustLoadContext[cfg](context.Background())
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
		c, err := confkit.ValidateOnly[cfg](context.Background())
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

		_, err := confkit.ValidateOnly[cfg](context.Background(),
			confkit.WithLoadHook(func(bool, time.Duration, int) { hookCalled = true }),
			confkit.WithAuditLogger(func([]confkit.AuditEntry) { auditCalled = true }),
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
		_, err := confkit.ValidateOnly[cfg](context.Background(), confkit.WithSource(confkit.FromEnv()))
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
		_, err := confkit.LoadWithOptions[cfg](
			confkit.WithSource(confkit.FromEnv()),
			confkit.WithAuditLogger(func([]confkit.AuditEntry) { called = true }),
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
		_, _ = confkit.LoadWithOptions[cfg](
			confkit.WithSource(confkit.NewErrorSource(fmt.Errorf("src err"))),
			confkit.WithAuditLogger(func([]confkit.AuditEntry) { called = true }),
		)
		if !called {
			t.Error("expected audit logger to be called on source error")
		}
	})
}
