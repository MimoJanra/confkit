package confkit

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestErrorReport(t *testing.T) {
	t.Run("implements_error", func(t *testing.T) {
		report := &ErrorReport{}
		report.AddError(FieldError{
			Path:    "Port",
			Kind:    ErrorKindValidation,
			Rule:    "required",
			Message: "PORT is required",
		})
		msg := report.Error()
		if !strings.Contains(msg, "PORT") {
			t.Errorf("Error() should contain field name: %s", msg)
		}
	})

	t.Run("unwrap", func(t *testing.T) {
		sentinel := fmt.Errorf("underlying cause")
		report := &ErrorReport{}
		report.AddError(FieldError{
			Path:    "Host",
			Kind:    ErrorKindIO,
			Message: "could not read",
			Err:     sentinel,
		})
		errs := report.Unwrap()
		if len(errs) != 1 {
			t.Fatalf("expected 1 unwrapped error, got %d", len(errs))
		}
		if !errors.Is(errs[0], sentinel) {
			t.Error("Unwrap should expose the underlying Err")
		}
	})

	t.Run("unwrap_empty", func(t *testing.T) {
		report := &ErrorReport{}
		errs := report.Unwrap()
		if len(errs) != 0 {
			t.Errorf("empty report should unwrap to empty slice, got %d", len(errs))
		}
	})

	t.Run("join", func(t *testing.T) {
		report := &ErrorReport{}
		report.AddError(FieldError{Path: "A", Kind: ErrorKindValidation, Message: "a is bad"})
		report.AddError(FieldError{Path: "B", Kind: ErrorKindValidation, Message: "b is bad"})
		msg := report.Error()
		if !strings.Contains(msg, "A") || !strings.Contains(msg, "B") {
			t.Errorf("joined error should list all fields: %s", msg)
		}
	})

	t.Run("first_error", func(t *testing.T) {
		report := &ErrorReport{}
		report.AddError(FieldError{Path: "X", Kind: ErrorKindParse, Message: "bad parse"})
		report.AddError(FieldError{Path: "Y", Kind: ErrorKindValidation, Message: "bad val"})
		if len(report.Errors) != 2 {
			t.Fatalf("expected 2 errors, got %d", len(report.Errors))
		}
		if report.Errors[0].Path != "X" {
			t.Errorf("expected first error at X, got %s", report.Errors[0].Path)
		}
	})
}

func TestExplain(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		if got := Explain(nil); got != "" {
			t.Errorf("Explain(nil) should return empty string, got %q", got)
		}
	})

	t.Run("non_report", func(t *testing.T) {
		err := fmt.Errorf("plain error")
		got := Explain(err)
		if got != "plain error" {
			t.Errorf("Explain of plain error should return its message, got %q", got)
		}
	})

	t.Run("with_field", func(t *testing.T) {
		report := &ErrorReport{}
		report.AddError(FieldError{
			Path:    "Port",
			Kind:    ErrorKindValidation,
			Rule:    "min",
			Message: "PORT must be >= 1",
			Value:   "0",
		})
		got := Explain(report)
		if !strings.Contains(got, "Port") {
			t.Errorf("Explain should include field name: %s", got)
		}
		if !strings.Contains(got, "PORT must be >= 1") {
			t.Errorf("Explain should include message: %s", got)
		}
		if !strings.Contains(got, "0") {
			t.Errorf("Explain should include value: %s", got)
		}
	})

	t.Run("secret_redaction", func(t *testing.T) {
		report := &ErrorReport{}
		report.AddError(FieldError{
			Path:    "APIKey",
			Kind:    ErrorKindValidation,
			Message: "must be non-empty",
			Value:   "topsecret",
			Secret:  true,
		})
		got := Explain(report)
		if strings.Contains(got, "topsecret") {
			t.Errorf("Explain should redact secret value: %s", got)
		}
		if !strings.Contains(got, "***REDACTED***") {
			t.Errorf("Explain should show redaction marker: %s", got)
		}
	})
}

func TestNewErrorSource(t *testing.T) {
	t.Run("propagates_error", func(t *testing.T) {
		type cfg struct{ X string `env:"ERR_X"` }
		src := NewErrorSource(fmt.Errorf("intentional failure"))
		_, err := Load[cfg](src)
		if err == nil {
			t.Fatal("expected error from error source")
		}
	})

	t.Run("report_kind", func(t *testing.T) {
		type cfg struct{ X string `env:"ERR_X2"` }
		src := NewErrorSource(fmt.Errorf("io problem"))
		_, err := Load[cfg](src)
		report, ok := err.(*ErrorReport)
		if !ok {
			t.Fatalf("expected *ErrorReport, got %T", err)
		}
		if len(report.Errors) == 0 {
			t.Fatal("expected at least one error")
		}
		if report.Errors[0].Kind != ErrorKindIO {
			t.Errorf("expected ErrorKindIO, got %q", report.Errors[0].Kind)
		}
	})
}
