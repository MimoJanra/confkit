package interpolation

import (
	"os"
	"testing"
)

func TestResolve(t *testing.T) {
	t.Run("env_var", func(t *testing.T) {
		_ = os.Setenv("TEST_VAR", "value123")
		defer func() { _ = os.Unsetenv("TEST_VAR") }()

		r := NewResolver(10)
		result, err := r.Resolve("prefix-${TEST_VAR}-suffix", "test.field")
		if err != nil {
			t.Fatalf("Resolve failed: %v", err)
		}
		if result != "prefix-value123-suffix" {
			t.Errorf("expected 'prefix-value123-suffix', got '%s'", result)
		}
	})

	t.Run("default_fallback", func(t *testing.T) {
		_ = os.Unsetenv("NONEXISTENT_VAR")

		r := NewResolver(10)
		result, err := r.Resolve("${NONEXISTENT_VAR|fallback}", "test.field")
		if err != nil {
			t.Fatalf("Resolve failed: %v", err)
		}
		if result != "fallback" {
			t.Errorf("expected 'fallback', got '%s'", result)
		}
	})

	t.Run("multiple_vars", func(t *testing.T) {
		_ = os.Setenv("HOST", "localhost")
		_ = os.Setenv("PORT", "8080")
		defer func() {
			_ = os.Unsetenv("HOST")
			_ = os.Unsetenv("PORT")
		}()

		r := NewResolver(10)
		result, err := r.Resolve("http://${HOST}:${PORT}/api", "test.field")
		if err != nil {
			t.Fatalf("Resolve failed: %v", err)
		}
		if result != "http://localhost:8080/api" {
			t.Errorf("expected 'http://localhost:8080/api', got '%s'", result)
		}
	})

	t.Run("config_lookup", func(t *testing.T) {
		r := NewResolver(10)
		r.SetConfigValue("BASE_URL", "https://api.example.com")

		result, err := r.Resolve("${BASE_URL}/v1/users", "test.field")
		if err != nil {
			t.Fatalf("Resolve failed: %v", err)
		}
		if result != "https://api.example.com/v1/users" {
			t.Errorf("expected 'https://api.example.com/v1/users', got '%s'", result)
		}
	})

	t.Run("no_vars_passthrough", func(t *testing.T) {
		r := NewResolver(10)
		result, err := r.Resolve("plain text", "test.field")
		if err != nil {
			t.Fatalf("Resolve failed: %v", err)
		}
		if result != "plain text" {
			t.Errorf("expected 'plain text', got '%s'", result)
		}
	})

	t.Run("undefined_var_error", func(t *testing.T) {
		_ = os.Unsetenv("UNDEFINED_VAR")

		r := NewResolver(10)
		_, err := r.Resolve("${UNDEFINED_VAR}", "test.field")
		if err == nil {
			t.Fatal("expected error for undefined variable")
		}
	})

	t.Run("max_depth_error", func(t *testing.T) {
		r := NewResolver(2)
		r.SetConfigValue("A", "${B}")
		r.SetConfigValue("B", "${C}")
		r.SetConfigValue("C", "${D}")

		_, err := r.Resolve("${A}", "test.field")
		if err == nil {
			t.Fatal("expected error for max depth exceeded")
		}
	})

	t.Run("escaped_dollar", func(t *testing.T) {
		r := NewResolver(10)
		result, err := r.Resolve("$${ESCAPED}", "test.field")
		if err != nil {
			t.Fatalf("Resolve failed: %v", err)
		}
		if result != "${ESCAPED}" {
			t.Errorf("expected '${ESCAPED}', got '%s'", result)
		}
	})
}

func TestResolveEmptyDefault(t *testing.T) {
	r := NewResolver(10)
	got, err := r.Resolve("prefix-${UNDEFINED_VAR_FOR_TEST|}-suffix", "Field")
	if err != nil {
		t.Fatalf("explicit empty default must not error: %v", err)
	}
	if want := "prefix--suffix"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestResolveEscapedDollarWithoutPlaceholder(t *testing.T) {
	r := NewResolver(10)
	got, err := r.Resolve("a$$b", "Field")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := "a$b"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
