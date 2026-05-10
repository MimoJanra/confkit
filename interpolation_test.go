package confkit

import (
	"os"
	"testing"
)

func TestInterpolationBasic(t *testing.T) {
	_ = os.Setenv("TEST_VAR", "value123")
	defer func() { _ = os.Unsetenv("TEST_VAR") }()

	r := NewInterpolationResolver(10)
	result, err := r.Resolve("prefix-${TEST_VAR}-suffix", "test.field")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if result != "prefix-value123-suffix" {
		t.Errorf("Expected 'prefix-value123-suffix', got '%s'", result)
	}
}

func TestInterpolationWithDefault(t *testing.T) {
	_ = os.Unsetenv("NONEXISTENT_VAR")

	r := NewInterpolationResolver(10)
	result, err := r.Resolve("${NONEXISTENT_VAR|fallback}", "test.field")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if result != "fallback" {
		t.Errorf("Expected 'fallback', got '%s'", result)
	}
}

func TestInterpolationMultipleVars(t *testing.T) {
	_ = os.Setenv("HOST", "localhost")
	_ = os.Setenv("PORT", "8080")
	defer func() { _ = os.Unsetenv("HOST") }()
	defer func() { _ = os.Unsetenv("PORT") }()

	r := NewInterpolationResolver(10)
	result, err := r.Resolve("http://${HOST}:${PORT}/api", "test.field")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if result != "http://localhost:8080/api" {
		t.Errorf("Expected 'http://localhost:8080/api', got '%s'", result)
	}
}

func TestInterpolationConfigLookup(t *testing.T) {
	r := NewInterpolationResolver(10)
	r.SetConfigValue("BASE_URL", "https://api.example.com")

	result, err := r.Resolve("${BASE_URL}/v1/users", "test.field")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if result != "https://api.example.com/v1/users" {
		t.Errorf("Expected 'https://api.example.com/v1/users', got '%s'", result)
	}
}

func TestInterpolationNoVars(t *testing.T) {
	r := NewInterpolationResolver(10)
	result, err := r.Resolve("plain text", "test.field")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if result != "plain text" {
		t.Errorf("Expected 'plain text', got '%s'", result)
	}
}

func TestInterpolationUndefinedVar(t *testing.T) {
	_ = os.Unsetenv("UNDEFINED_VAR")

	r := NewInterpolationResolver(10)
	_, err := r.Resolve("${UNDEFINED_VAR}", "test.field")
	if err == nil {
		t.Fatal("Expected error for undefined variable, got nil")
	}
	if err.Error() != "interpolation in test.field: circular reference or undefined variable: ${UNDEFINED_VAR}" {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestInterpolationMaxDepth(t *testing.T) {
	r := NewInterpolationResolver(2)
	r.SetConfigValue("A", "${B}")
	r.SetConfigValue("B", "${C}")
	r.SetConfigValue("C", "${D}")

	_, err := r.Resolve("${A}", "test.field")
	if err == nil {
		t.Fatal("Expected error for max depth exceeded, got nil")
	}
}

func TestInterpolationEscaped(t *testing.T) {
	r := NewInterpolationResolver(10)
	result, err := r.Resolve("$${ESCAPED}", "test.field")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	if result != "${ESCAPED}" {
		t.Errorf("Expected '${ESCAPED}', got '%s'", result)
	}
}
