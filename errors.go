package confkit

import "strings"

// ErrorKind classifies why a field failed, so callers can react programmatically
// instead of matching on message text.
type ErrorKind string

// The kinds of failure a FieldError can report.
const (
	// ErrorKindParse means the raw value could not be converted to the field type.
	ErrorKindParse ErrorKind = "parse"
	// ErrorKindValidation means the value parsed but broke a `validate` rule.
	ErrorKindValidation ErrorKind = "validation"
	// ErrorKindIO means a Source failed to retrieve the value at all.
	ErrorKindIO ErrorKind = "io"
)

// FieldError describes one problem with one field.
//
// Value holds the offending input, already replaced with a redaction marker when
// the field is tagged `secret:"true"`; Secret records that the field was secret.
// Err, when set, is the underlying cause and makes errors.Is work through the
// enclosing ErrorReport.
type FieldError struct {
	Path    string
	Source  string
	Kind    ErrorKind
	Rule    string
	Message string
	Value   string
	Secret  bool
	Err     error // underlying error, enables errors.Is checks on ErrorReport
}

// ErrorReport collects every field problem found during a single load, so one
// error can report all of them instead of only the first.
//
// It implements error, and Unwrap() []error lets errors.Is and errors.As reach the
// individual causes.
type ErrorReport struct {
	Errors []FieldError
}

func (er *ErrorReport) Error() string {
	return er.Format()
}

func (er *ErrorReport) Unwrap() []error {
	errs := make([]error, len(er.Errors))
	for i, fe := range er.Errors {
		errs[i] = &singleFieldError{fe}
	}
	return errs
}

type singleFieldError struct {
	FieldError
}

func (e *singleFieldError) Error() string {
	return e.Message
}

func (e *singleFieldError) Unwrap() error {
	return e.Err
}

// Format renders the report as a multi-line, human-readable message listing each
// field, its source, the rule that failed and the offending value. Values of
// secret fields are shown redacted.
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
				sb.WriteString("***REDACTED***")
			} else {
				sb.WriteString(err.Value)
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// Explain formats err for display. An *ErrorReport is rendered by Format; any
// other error is returned as its own message. A nil error yields "".
func Explain(err error) string {
	if err == nil {
		return ""
	}
	if report, ok := err.(*ErrorReport); ok {
		return report.Format()
	}
	return err.Error()
}

// AddError appends fe to the report.
func (er *ErrorReport) AddError(fe FieldError) {
	er.Errors = append(er.Errors, fe)
}

// IsEmpty reports whether the report holds no errors.
func (er *ErrorReport) IsEmpty() bool {
	return len(er.Errors) == 0
}

// FirstError returns a pointer to the first error, or nil if the report is empty.
func (er *ErrorReport) FirstError() *FieldError {
	if len(er.Errors) == 0 {
		return nil
	}
	return &er.Errors[0]
}
