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

func TestBuiltinValidatorsExtended(t *testing.T) {
	run := func(envKey, envVal string, load func() error, wantErr bool) func(*testing.T) {
		return func(t *testing.T) {
			t.Helper()
			if envVal == "" {
				_ = os.Unsetenv(envKey)
			} else {
				_ = os.Setenv(envKey, envVal)
				t.Cleanup(func() { _ = os.Unsetenv(envKey) })
			}
			err := load()
			if wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		}
	}

	type HTTPURLCfg struct {
		URL string `env:"V_HTTPURL" validate:"http_url"`
	}
	t.Run("http_url/valid", run("V_HTTPURL", "https://example.com/path", func() error {
		_, e := confkit.Load[HTTPURLCfg](confkit.FromEnv())
		return e
	}, false))
	t.Run("http_url/invalid", run("V_HTTPURL", "ftp://example.com", func() error {
		_, e := confkit.Load[HTTPURLCfg](confkit.FromEnv())
		return e
	}, true))

	type IPv4Cfg struct {
		IP string `env:"V_IPV4" validate:"ipv4"`
	}
	t.Run("ipv4/valid", run("V_IPV4", "192.168.1.1", func() error {
		_, e := confkit.Load[IPv4Cfg](confkit.FromEnv())
		return e
	}, false))
	t.Run("ipv4/invalid", run("V_IPV4", "::1", func() error {
		_, e := confkit.Load[IPv4Cfg](confkit.FromEnv())
		return e
	}, true))

	type IPv6Cfg struct {
		IP string `env:"V_IPV6" validate:"ipv6"`
	}
	t.Run("ipv6/valid", run("V_IPV6", "::1", func() error {
		_, e := confkit.Load[IPv6Cfg](confkit.FromEnv())
		return e
	}, false))
	t.Run("ipv6/invalid", run("V_IPV6", "192.168.1.1", func() error {
		_, e := confkit.Load[IPv6Cfg](confkit.FromEnv())
		return e
	}, true))

	type UUIDCfg struct {
		ID string `env:"V_UUID" validate:"uuid"`
	}
	t.Run("uuid/valid", run("V_UUID", "550e8400-e29b-41d4-a716-446655440000", func() error {
		_, e := confkit.Load[UUIDCfg](confkit.FromEnv())
		return e
	}, false))
	t.Run("uuid/invalid", run("V_UUID", "not-a-uuid", func() error {
		_, e := confkit.Load[UUIDCfg](confkit.FromEnv())
		return e
	}, true))

	type HostCfg struct {
		Host string `env:"V_HOST" validate:"hostname"`
	}
	t.Run("hostname/valid", run("V_HOST", "example.com", func() error {
		_, e := confkit.Load[HostCfg](confkit.FromEnv())
		return e
	}, false))
	t.Run("hostname/invalid", run("V_HOST", "-invalid-", func() error {
		_, e := confkit.Load[HostCfg](confkit.FromEnv())
		return e
	}, true))

	type PortStrCfg struct {
		Port string `env:"V_PORTSTR" validate:"port"`
	}
	t.Run("port_str/valid", run("V_PORTSTR", "8080", func() error {
		_, e := confkit.Load[PortStrCfg](confkit.FromEnv())
		return e
	}, false))
	t.Run("port_str/invalid", run("V_PORTSTR", "99999", func() error {
		_, e := confkit.Load[PortStrCfg](confkit.FromEnv())
		return e
	}, true))

	type RegexCfg struct {
		Code string `env:"V_REGEX" validate:"regex=^[A-Z]{3}$"`
	}
	t.Run("regex/valid", run("V_REGEX", "ABC", func() error {
		_, e := confkit.Load[RegexCfg](confkit.FromEnv())
		return e
	}, false))
	t.Run("regex/invalid", run("V_REGEX", "abc", func() error {
		_, e := confkit.Load[RegexCfg](confkit.FromEnv())
		return e
	}, true))

	type LenCfg struct {
		Code string `env:"V_LEN" validate:"len=3"`
	}
	t.Run("len/exact", run("V_LEN", "abc", func() error {
		_, e := confkit.Load[LenCfg](confkit.FromEnv())
		return e
	}, false))
	t.Run("len/wrong", run("V_LEN", "ab", func() error {
		_, e := confkit.Load[LenCfg](confkit.FromEnv())
		return e
	}, true))

	type ContainsCfg struct {
		S string `env:"V_CONTAINS" validate:"contains=foo"`
	}
	t.Run("contains/yes", run("V_CONTAINS", "foobar", func() error {
		_, e := confkit.Load[ContainsCfg](confkit.FromEnv())
		return e
	}, false))
	t.Run("contains/no", run("V_CONTAINS", "bar", func() error {
		_, e := confkit.Load[ContainsCfg](confkit.FromEnv())
		return e
	}, true))

	type SWCfg struct {
		S string `env:"V_SW" validate:"startswith=foo"`
	}
	t.Run("startswith/yes", run("V_SW", "foobar", func() error {
		_, e := confkit.Load[SWCfg](confkit.FromEnv())
		return e
	}, false))
	t.Run("startswith/no", run("V_SW", "barfoo", func() error {
		_, e := confkit.Load[SWCfg](confkit.FromEnv())
		return e
	}, true))

	type EWCfg struct {
		S string `env:"V_EW" validate:"endswith=bar"`
	}
	t.Run("endswith/yes", run("V_EW", "foobar", func() error {
		_, e := confkit.Load[EWCfg](confkit.FromEnv())
		return e
	}, false))
	t.Run("endswith/no", run("V_EW", "barfoo", func() error {
		_, e := confkit.Load[EWCfg](confkit.FromEnv())
		return e
	}, true))

	type AlphaCfg struct {
		S string `env:"V_ALPHA" validate:"alpha"`
	}
	t.Run("alpha/valid", run("V_ALPHA", "abc", func() error {
		_, e := confkit.Load[AlphaCfg](confkit.FromEnv())
		return e
	}, false))
	t.Run("alpha/invalid", run("V_ALPHA", "abc123", func() error {
		_, e := confkit.Load[AlphaCfg](confkit.FromEnv())
		return e
	}, true))

	type AlnumCfg struct {
		S string `env:"V_ALNUM" validate:"alphanum"`
	}
	t.Run("alphanum/valid", run("V_ALNUM", "abc123", func() error {
		_, e := confkit.Load[AlnumCfg](confkit.FromEnv())
		return e
	}, false))
	t.Run("alphanum/invalid", run("V_ALNUM", "abc!", func() error {
		_, e := confkit.Load[AlnumCfg](confkit.FromEnv())
		return e
	}, true))

	type NumericCfg struct {
		S string `env:"V_NUM" validate:"numeric"`
	}
	t.Run("numeric/valid", run("V_NUM", "12345", func() error {
		_, e := confkit.Load[NumericCfg](confkit.FromEnv())
		return e
	}, false))
	t.Run("numeric/invalid", run("V_NUM", "123abc", func() error {
		_, e := confkit.Load[NumericCfg](confkit.FromEnv())
		return e
	}, true))

	type LowerCfg struct {
		S string `env:"V_LOWER" validate:"lowercase"`
	}
	t.Run("lowercase/valid", run("V_LOWER", "abc", func() error {
		_, e := confkit.Load[LowerCfg](confkit.FromEnv())
		return e
	}, false))
	t.Run("lowercase/invalid", run("V_LOWER", "Abc", func() error {
		_, e := confkit.Load[LowerCfg](confkit.FromEnv())
		return e
	}, true))

	type UpperCfg struct {
		S string `env:"V_UPPER" validate:"uppercase"`
	}
	t.Run("uppercase/valid", run("V_UPPER", "ABC", func() error {
		_, e := confkit.Load[UpperCfg](confkit.FromEnv())
		return e
	}, false))
	t.Run("uppercase/invalid", run("V_UPPER", "Abc", func() error {
		_, e := confkit.Load[UpperCfg](confkit.FromEnv())
		return e
	}, true))

	type NECfg struct {
		S string `env:"V_NE" validate:"notempty"`
	}
	t.Run("notempty/valid", run("V_NE", "hello", func() error {
		_, e := confkit.Load[NECfg](confkit.FromEnv())
		return e
	}, false))
	t.Run("notempty/invalid", run("V_NE", "   ", func() error {
		_, e := confkit.Load[NECfg](confkit.FromEnv())
		return e
	}, true))

	// empty string skips format validators (strCheck early return)
	type OptEmailCfg struct {
		Email string `env:"V_OPT_EMAIL" validate:"email"`
	}
	t.Run("optional/empty_skips", run("V_OPT_EMAIL", "", func() error {
		_, e := confkit.Load[OptEmailCfg](confkit.FromEnv())
		return e
	}, false))
}

func TestPortIntField(t *testing.T) {
	type Cfg struct {
		Port int `env:"V_PORT_INT" validate:"port"`
	}
	t.Run("valid", func(t *testing.T) {
		t.Setenv("V_PORT_INT", "8080")
		_, err := confkit.Load[Cfg](confkit.FromEnv())
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	t.Run("invalid_high", func(t *testing.T) {
		t.Setenv("V_PORT_INT", "99999")
		_, err := confkit.Load[Cfg](confkit.FromEnv())
		if err == nil {
			t.Error("expected error for port > 65535")
		}
	})
	t.Run("invalid_zero", func(t *testing.T) {
		t.Setenv("V_PORT_INT", "0")
		_, err := confkit.Load[Cfg](confkit.FromEnv())
		if err == nil {
			t.Error("expected error for port 0")
		}
	})
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
