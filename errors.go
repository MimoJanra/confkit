package confkit

import "strings"

type ErrorKind string

const (
	ErrorKindParse      ErrorKind = "parse"
	ErrorKindValidation ErrorKind = "validation"
	ErrorKindIO         ErrorKind = "io"
)

type FieldError struct {
	Path    string
	Source  string
	Kind    ErrorKind
	Rule    string
	Message string
	Value   string
	Secret  bool
}

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

func Explain(err error) string {
	if err == nil {
		return ""
	}
	if report, ok := err.(*ErrorReport); ok {
		return report.Format()
	}
	return err.Error()
}

func (er *ErrorReport) AddError(fe FieldError) {
	er.Errors = append(er.Errors, fe)
}

func (er *ErrorReport) IsEmpty() bool {
	return len(er.Errors) == 0
}
