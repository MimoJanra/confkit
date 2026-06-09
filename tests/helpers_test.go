package confkit_test

import (
	"os"
	"strings"
	"testing"

	confkit "github.com/MimoJanra/confkit"
)

func writeTempYAML(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "test-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = f.Close() }()
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	return f.Name()
}

func writeTempJSON(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "test-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = f.Close() }()
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	return f.Name()
}

func writeTempTOML(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "test-*.toml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = f.Close() }()
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	return f.Name()
}

func assertReport(t *testing.T, err error) *confkit.ErrorReport {
	t.Helper()
	report, ok := err.(*confkit.ErrorReport)
	if !ok {
		t.Fatalf("expected *ErrorReport, got %T: %v", err, err)
	}
	if len(report.Errors) == 0 {
		t.Fatal("expected at least one error in report")
	}
	return report
}

func stringContains(s, substr string) bool {
	return strings.Contains(s, substr)
}
