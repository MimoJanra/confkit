package confkit

import (
	"strings"
)

// ErrorKind describes the type of configuration error.
type ErrorKind string

const (
	ErrorKindParse      ErrorKind = "parse"      // failed to parse value to target type
	ErrorKindValidation ErrorKind = "validation" // validation rule failed
	ErrorKindMissing    ErrorKind = "missing"    // required field not provided
	ErrorKindIO         ErrorKind = "io"         // file not found, read error, etc.
)

// FieldError represents a single field configuration error.
type FieldError struct {
	Path    string    // dot-separated path (e.g., "server.port")
	Source  string    // source identifier (e.g., "env PORT", "yaml", "default")
	Kind    ErrorKind // error category
	Rule    string    // validation rule that failed (e.g., "min", "required", "url")
	Message string    // human-readable error message
	Value   string    // actual value provided
	Secret  bool      // if true, redact value in output
}

// ErrorReport is a collection of field errors during configuration loading.
type ErrorReport struct {
	Errors []FieldError
}

// Error implements the error interface.
func (er *ErrorReport) Error() string {
	return er.Format()
}

// Format returns a human-readable formatted error report.
func (er *ErrorReport) Format() string {
	if len(er.Errors) == 0 {
		return "no errors"
	}

	var sb strings.Builder
	sb.WriteString("Invalid configuration:\n")

	for _, err := range er.Errors {
		sb.WriteString("\n  ")
		sb.WriteString(err.Path)
		sb.WriteString("\n")

		if err.Source != "" {
			sb.WriteString("    source: ")
			sb.WriteString(err.Source)
			sb.WriteString("\n")
		}

		sb.WriteString("    error: ")
		sb.WriteString(err.Message)
		sb.WriteString("\n")

		if err.Value != "" {
			sb.WriteString("    got: ")
			if err.Secret {
				sb.WriteString("<redacted>")
			} else {
				sb.WriteString(err.Value)
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// Explain formats an error for user display. Returns the original error if it's not an ErrorReport.
func Explain(err error) string {
	if err == nil {
		return ""
	}

	var report *ErrorReport
	if ok := asErrorReport(err, &report); ok {
		return report.Format()
	}

	return err.Error()
}

// asErrorReport extracts an ErrorReport from an error using type assertion.
func asErrorReport(err error, target **ErrorReport) bool {
	if err == nil {
		return false
	}
	report, ok := err.(*ErrorReport)
	if ok {
		*target = report
		return true
	}
	return false
}

// AddError appends a FieldError to the report.
func (er *ErrorReport) AddError(fe FieldError) {
	er.Errors = append(er.Errors, fe)
}

// IsEmpty checks if the report has any errors.
func (er *ErrorReport) IsEmpty() bool {
	return len(er.Errors) == 0
}
