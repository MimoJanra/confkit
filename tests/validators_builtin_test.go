package confkit_test

import (
	"fmt"
	"os"
	"testing"

	confkit "github.com/MimoJanra/confkit"
)

func TestBuiltinValidators(t *testing.T) {
	run := func(envKey, envVal string, load func() error, wantErr bool) func(*testing.T) {
		return func(t *testing.T) {
			t.Helper()
			_ = os.Setenv(envKey, envVal)
			t.Cleanup(func() { _ = os.Unsetenv(envKey) })
			err := load()
			if wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		}
	}

	type EmailCfg struct {
		Email string `env:"V_EMAIL" validate:"email"`
	}
	t.Run("email/valid", run("V_EMAIL", "user@example.com", func() error {
		_, e := confkit.Load[EmailCfg](confkit.FromEnv())
		return e
	}, false))
	t.Run("email/invalid", run("V_EMAIL", "notanemail", func() error {
		_, e := confkit.Load[EmailCfg](confkit.FromEnv())
		return e
	}, true))

	type URLCfg struct {
		URL string `env:"V_URL" validate:"url"`
	}
	t.Run("url/valid", run("V_URL", "https://example.com/path", func() error {
		_, e := confkit.Load[URLCfg](confkit.FromEnv())
		return e
	}, false))
	t.Run("url/invalid", run("V_URL", "not-a-url", func() error {
		_, e := confkit.Load[URLCfg](confkit.FromEnv())
		return e
	}, true))

	type IPCfg struct {
		IP string `env:"V_IP" validate:"ip"`
	}
	t.Run("ip/v4", run("V_IP", "192.168.1.1", func() error {
		_, e := confkit.Load[IPCfg](confkit.FromEnv())
		return e
	}, false))
	t.Run("ip/v6", run("V_IP", "::1", func() error {
		_, e := confkit.Load[IPCfg](confkit.FromEnv())
		return e
	}, false))
	t.Run("ip/invalid", run("V_IP", "999.999.999.999", func() error {
		_, e := confkit.Load[IPCfg](confkit.FromEnv())
		return e
	}, true))
}

func TestModelValidator(t *testing.T) {
	type Config struct {
		Username string `env:"MV_USER"`
		Password string `env:"MV_PASS"`
	}

	t.Run("passes", func(t *testing.T) {
		t.Setenv("MV_USER", "alice")
		t.Setenv("MV_PASS", "secret123")

		_, err := confkit.LoadWithOptions[Config](
			confkit.WithSource(confkit.FromEnv()),
			confkit.WithModelValidator(func(c *Config) error {
				if c.Username == "" {
					return fmt.Errorf("username required")
				}
				return nil
			}),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("fails", func(t *testing.T) {
		t.Setenv("MV_USER", "")
		t.Setenv("MV_PASS", "pass")

		_, err := confkit.LoadWithOptions[Config](
			confkit.WithSource(confkit.FromEnv()),
			confkit.WithModelValidator(func(c *Config) error {
				if c.Username == "" {
					return fmt.Errorf("username required")
				}
				return nil
			}),
		)
		if err == nil {
			t.Fatal("expected model validator error")
		}
	})
}

func TestAuditLoggerBasic(t *testing.T) {
	type Cfg struct {
		Host string `env:"AUDIT_HOST" default:"localhost"`
		Port int    `env:"AUDIT_PORT" default:"5432"`
	}
	var logged []confkit.AuditEntry
	_, err := confkit.LoadWithOptions[Cfg](
		confkit.WithSource(confkit.FromEnv()),
		confkit.WithAuditLogger(func(entries []confkit.AuditEntry) {
			logged = entries
		}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(logged) == 0 {
		t.Error("expected audit entries, got none")
	}
	for _, e := range logged {
		if e.Field == "" || e.Source == "" {
			t.Errorf("incomplete audit entry: %+v", e)
		}
	}
}

func TestMultiFileYAML(t *testing.T) {
	base := t.TempDir() + "/base.yaml"
	override := t.TempDir() + "/override.yaml"

	_ = os.WriteFile(base, []byte("host: localhost\nport: 5432\n"), 0644)
	_ = os.WriteFile(override, []byte("host: db.prod.internal\n"), 0644)

	type Cfg struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	}
	cfg, err := confkit.Load[Cfg](confkit.FromYAMLFiles(base, override))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Host != "db.prod.internal" {
		t.Errorf("expected db.prod.internal, got %s", cfg.Host)
	}
	if cfg.Port != 5432 {
		t.Errorf("expected 5432, got %d", cfg.Port)
	}
}
