package confkit

import (
	"fmt"
	"reflect"

	"github.com/go-playground/validator/v10"
)

// Validator wraps go-playground/validator with confkit integration.
type Validator struct {
	v *validator.Validate
}

// NewValidator creates a new validator instance.
func NewValidator() *Validator {
	return &Validator{
		v: validator.New(),
	}
}

// ValidateConfig validates a populated config struct against its validate tags.
// Returns an ErrorReport with field-level validation errors, or nil if all valid.
func (vdr *Validator) ValidateConfig(cfg any, fields []FieldInfo) *ErrorReport {
	report := &ErrorReport{}

	// Use reflect to validate the struct
	val := reflect.ValueOf(cfg)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Validate each field that has a validate tag
	for _, field := range fields {
		validateTag := field.Tags["validate"]
		if validateTag == "" {
			continue // No validation rule for this field
		}

		// Get the field value by path
		fieldVal := getFieldByPath(val, field.Path)
		if !fieldVal.IsValid() {
			continue
		}

		// Run validation on this field using the validator
		// The Var method validates a single value against rules
		if err := vdr.v.Var(fieldVal.Interface(), validateTag); err != nil {
			// err is a FieldError or ValidationErrors
			if validationErr, ok := err.(validator.ValidationErrors); ok {
				for _, ve := range validationErr {
					fieldErr := FieldError{
						Path:    field.Path,
						Kind:    ErrorKindValidation,
						Rule:    ve.Tag(),
						Secret:  field.IsSecret,
						Value:   fieldValueToString(fieldVal, field.IsSecret),
						Source:  "validation",
						Message: formatValidationError(ve.Error(), ve.Tag()),
					}
					report.AddError(fieldErr)
				}
			} else {
				// Generic error
				fieldErr := FieldError{
					Path:    field.Path,
					Kind:    ErrorKindValidation,
					Rule:    validateTag,
					Secret:  field.IsSecret,
					Value:   fieldValueToString(fieldVal, field.IsSecret),
					Source:  "validation",
					Message: err.Error(),
				}
				report.AddError(fieldErr)
			}
		}
	}

	return report
}

// getFieldByPath retrieves a field value by dot-separated path (e.g., "Database.Host").
func getFieldByPath(val reflect.Value, path string) reflect.Value {
	parts := splitPath(path)

	current := val
	for _, part := range parts {
		if current.Kind() == reflect.Ptr {
			if current.IsNil() {
				return reflect.Value{}
			}
			current = current.Elem()
		}

		if current.Kind() != reflect.Struct {
			return reflect.Value{}
		}

		current = current.FieldByName(part)
		if !current.IsValid() {
			return reflect.Value{}
		}
	}

	return current
}

// splitPath splits a dot-separated path (e.g., "Database.Host" -> ["Database", "Host"]).
func splitPath(path string) []string {
	// Simple split; no escaping for now
	var parts []string
	var current string
	for _, ch := range path {
		if ch == '.' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

// fieldValueToString converts a field value to string for error messages.
func fieldValueToString(val reflect.Value, isSecret bool) string {
	if isSecret {
		return "[REDACTED]"
	}

	switch val.Kind() {
	case reflect.String:
		return val.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", val.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", val.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%v", val.Float())
	case reflect.Bool:
		return fmt.Sprintf("%v", val.Bool())
	case reflect.Slice:
		return fmt.Sprintf("%v", val.Interface())
	default:
		return fmt.Sprintf("%v", val.Interface())
	}
}

// formatValidationError converts a validator error message into a human-readable format.
func formatValidationError(errMsg, rule string) string {
	// errMsg comes from go-playground/validator
	// Try to extract meaningful info and format nicely

	// Some common rules and their messages
	switch rule {
	case "required":
		return "field is required"
	}

	// For min/max/pattern rules, try to extract the constraint
	// The error message from validator typically includes the rule
	// e.g., "Key: 'Port' Error: Field validation for 'Port' failed on the 'min' tag"

	// For now, return a generic message based on the rule
	if len(errMsg) > 0 {
		return errMsg
	}

	return fmt.Sprintf("validation failed: %s", rule)
}
