package confkit

import (
	"fmt"
	"os"
	"testing"
)

func TestBuiltinValidators(t *testing.T) {
	run := func(envKey, envVal string, load func() error, wantErr bool) func(*testing.T) {
		return func(t *testing.T) {
			t.Helper()
			os.Setenv(envKey, envVal)
			t.Cleanup(func() { os.Unsetenv(envKey) })
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
	t.Run("email/valid", run("V_EMAIL", "user@example.com", func() error { _, e := Load[EmailCfg](FromEnv()); return e }, false))
	t.Run("email/invalid", run("V_EMAIL", "notanemail", func() error { _, e := Load[EmailCfg](FromEnv()); return e }, true))

	type URLCfg struct {
		URL string `env:"V_URL" validate:"url"`
	}
	t.Run("url/valid", run("V_URL", "https://example.com/path", func() error { _, e := Load[URLCfg](FromEnv()); return e }, false))
	t.Run("url/invalid", run("V_URL", "not-a-url", func() error { _, e := Load[URLCfg](FromEnv()); return e }, true))

	type HTTPURLCfg struct {
		URL string `env:"V_HTTPURL" validate:"http_url"`
	}
	t.Run("http_url/valid", run("V_HTTPURL", "http://example.com", func() error { _, e := Load[HTTPURLCfg](FromEnv()); return e }, false))
	t.Run("http_url/rejects_ftp", run("V_HTTPURL", "ftp://example.com", func() error { _, e := Load[HTTPURLCfg](FromEnv()); return e }, true))

	type IPCfg struct {
		IP string `env:"V_IP" validate:"ip"`
	}
	t.Run("ip/v4", run("V_IP", "192.168.1.1", func() error { _, e := Load[IPCfg](FromEnv()); return e }, false))
	t.Run("ip/v6", run("V_IP", "::1", func() error { _, e := Load[IPCfg](FromEnv()); return e }, false))
	t.Run("ip/invalid", run("V_IP", "999.999.999.999", func() error { _, e := Load[IPCfg](FromEnv()); return e }, true))

	type IPv4Cfg struct {
		IP string `env:"V_IPV4" validate:"ipv4"`
	}
	t.Run("ipv4/valid", run("V_IPV4", "10.0.0.1", func() error { _, e := Load[IPv4Cfg](FromEnv()); return e }, false))
	t.Run("ipv4/rejects_v6", run("V_IPV4", "::1", func() error { _, e := Load[IPv4Cfg](FromEnv()); return e }, true))

	type IPv6Cfg struct {
		IP string `env:"V_IPV6" validate:"ipv6"`
	}
	t.Run("ipv6/valid", run("V_IPV6", "2001:db8::1", func() error { _, e := Load[IPv6Cfg](FromEnv()); return e }, false))
	t.Run("ipv6/rejects_v4", run("V_IPV6", "192.168.1.1", func() error { _, e := Load[IPv6Cfg](FromEnv()); return e }, true))

	type UUIDCfg struct {
		ID string `env:"V_UUID" validate:"uuid"`
	}
	t.Run("uuid/valid", run("V_UUID", "550e8400-e29b-41d4-a716-446655440000", func() error { _, e := Load[UUIDCfg](FromEnv()); return e }, false))
	t.Run("uuid/invalid", run("V_UUID", "not-a-uuid", func() error { _, e := Load[UUIDCfg](FromEnv()); return e }, true))

	type PortCfg struct {
		Port int `env:"V_PORT" validate:"port"`
	}
	t.Run("port/valid", run("V_PORT", "8080", func() error { _, e := Load[PortCfg](FromEnv()); return e }, false))
	t.Run("port/zero", run("V_PORT", "0", func() error { _, e := Load[PortCfg](FromEnv()); return e }, true))
	t.Run("port/overflow", run("V_PORT", "65536", func() error { _, e := Load[PortCfg](FromEnv()); return e }, true))

	type RegexCfg struct {
		Code string `env:"V_CODE" validate:"regex=^[A-Z]{3}$"`
	}
	t.Run("regex/match", run("V_CODE", "ABC", func() error { _, e := Load[RegexCfg](FromEnv()); return e }, false))
	t.Run("regex/no_match", run("V_CODE", "abc", func() error { _, e := Load[RegexCfg](FromEnv()); return e }, true))

	type LenCfg struct {
		PIN string `env:"V_PIN" validate:"len=4"`
	}
	t.Run("len/exact", run("V_PIN", "1234", func() error { _, e := Load[LenCfg](FromEnv()); return e }, false))
	t.Run("len/short", run("V_PIN", "12", func() error { _, e := Load[LenCfg](FromEnv()); return e }, true))

	type ContainsCfg struct {
		Tag string `env:"V_TAG" validate:"contains=go"`
	}
	t.Run("contains/ok", run("V_TAG", "golang", func() error { _, e := Load[ContainsCfg](FromEnv()); return e }, false))
	t.Run("contains/fail", run("V_TAG", "rust", func() error { _, e := Load[ContainsCfg](FromEnv()); return e }, true))

	type StartsWithCfg struct {
		Path string `env:"V_PATH" validate:"startswith=/"`
	}
	t.Run("startswith/ok", run("V_PATH", "/usr/local", func() error { _, e := Load[StartsWithCfg](FromEnv()); return e }, false))
	t.Run("startswith/fail", run("V_PATH", "usr/local", func() error { _, e := Load[StartsWithCfg](FromEnv()); return e }, true))

	type EndsWithCfg struct {
		File string `env:"V_FILE" validate:"endswith=.go"`
	}
	t.Run("endswith/ok", run("V_FILE", "main.go", func() error { _, e := Load[EndsWithCfg](FromEnv()); return e }, false))
	t.Run("endswith/fail", run("V_FILE", "main.py", func() error { _, e := Load[EndsWithCfg](FromEnv()); return e }, true))

	type AlphaCfg struct {
		Name string `env:"V_ALPHA" validate:"alpha"`
	}
	t.Run("alpha/ok", run("V_ALPHA", "Alice", func() error { _, e := Load[AlphaCfg](FromEnv()); return e }, false))
	t.Run("alpha/fail", run("V_ALPHA", "Alice1", func() error { _, e := Load[AlphaCfg](FromEnv()); return e }, true))

	type AlphaNumCfg struct {
		Slug string `env:"V_SLUG" validate:"alphanum"`
	}
	t.Run("alphanum/ok", run("V_SLUG", "user123", func() error { _, e := Load[AlphaNumCfg](FromEnv()); return e }, false))
	t.Run("alphanum/fail", run("V_SLUG", "user-123", func() error { _, e := Load[AlphaNumCfg](FromEnv()); return e }, true))

	type NumericCfg struct {
		Digits string `env:"V_DIGITS" validate:"numeric"`
	}
	t.Run("numeric/ok", run("V_DIGITS", "12345", func() error { _, e := Load[NumericCfg](FromEnv()); return e }, false))
	t.Run("numeric/fail", run("V_DIGITS", "123a5", func() error { _, e := Load[NumericCfg](FromEnv()); return e }, true))

	type LowerCfg struct {
		Key string `env:"V_LOWER" validate:"lowercase"`
	}
	t.Run("lowercase/ok", run("V_LOWER", "snake_case", func() error { _, e := Load[LowerCfg](FromEnv()); return e }, false))
	t.Run("lowercase/fail", run("V_LOWER", "CamelCase", func() error { _, e := Load[LowerCfg](FromEnv()); return e }, true))

	type UpperCfg struct {
		Key string `env:"V_UPPER" validate:"uppercase"`
	}
	t.Run("uppercase/ok", run("V_UPPER", "CONSTANT", func() error { _, e := Load[UpperCfg](FromEnv()); return e }, false))
	t.Run("uppercase/fail", run("V_UPPER", "Constant", func() error { _, e := Load[UpperCfg](FromEnv()); return e }, true))

	type NotEmptyCfg struct {
		Name string `env:"V_NOTEMPTY" validate:"notempty"`
	}
	t.Run("notempty/ok", run("V_NOTEMPTY", "Alice", func() error { _, e := Load[NotEmptyCfg](FromEnv()); return e }, false))
	t.Run("notempty/blank", run("V_NOTEMPTY", "   ", func() error { _, e := Load[NotEmptyCfg](FromEnv()); return e }, true))

	type HostnameCfg struct {
		Host string `env:"V_HOST" validate:"hostname"`
	}
	t.Run("hostname/valid", run("V_HOST", "example.com", func() error { _, e := Load[HostnameCfg](FromEnv()); return e }, false))
	t.Run("hostname/invalid", run("V_HOST", "-invalid-", func() error { _, e := Load[HostnameCfg](FromEnv()); return e }, true))
}

func TestModelValidator(t *testing.T) {
	type TLSConfig struct {
		TLSEnabled bool   `env:"MV_TLS_ENABLED"`
		CertPath   string `env:"MV_CERT_PATH"`
	}

	tlsCheck := WithModelValidator(func(cfg *TLSConfig) error {
		if cfg.TLSEnabled && cfg.CertPath == "" {
			return fmt.Errorf("cert_path is required when tls_enabled is true")
		}
		return nil
	})

	t.Run("passes when tls disabled", func(t *testing.T) {
		os.Setenv("MV_TLS_ENABLED", "false")
		t.Cleanup(func() { os.Unsetenv("MV_TLS_ENABLED"); os.Unsetenv("MV_CERT_PATH") })
		_, err := LoadWithOptions[TLSConfig](WithSource(FromEnv()), tlsCheck)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("fails when tls enabled without cert", func(t *testing.T) {
		os.Setenv("MV_TLS_ENABLED", "true")
		os.Setenv("MV_CERT_PATH", "")
		t.Cleanup(func() { os.Unsetenv("MV_TLS_ENABLED"); os.Unsetenv("MV_CERT_PATH") })
		_, err := LoadWithOptions[TLSConfig](WithSource(FromEnv()), tlsCheck)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("passes when tls enabled with cert", func(t *testing.T) {
		os.Setenv("MV_TLS_ENABLED", "true")
		os.Setenv("MV_CERT_PATH", "/etc/certs/server.crt")
		t.Cleanup(func() { os.Unsetenv("MV_TLS_ENABLED"); os.Unsetenv("MV_CERT_PATH") })
		_, err := LoadWithOptions[TLSConfig](WithSource(FromEnv()), tlsCheck)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestAuditLogger(t *testing.T) {
	type Cfg struct {
		Host string `env:"AUDIT_HOST" default:"localhost"`
		Port int    `env:"AUDIT_PORT" default:"5432"`
	}

	var logged []AuditEntry
	_, err := LoadWithOptions[Cfg](
		WithSource(FromEnv()),
		WithAuditLogger(func(entries []AuditEntry) {
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

	os.WriteFile(base, []byte("host: localhost\nport: 5432\n"), 0644)
	os.WriteFile(override, []byte("host: db.prod.internal\n"), 0644)

	type Cfg struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	}

	cfg, err := Load[Cfg](FromYAMLFiles(base, override))
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
