package confkit

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Validator handles basic validation of config fields.
// Supports: required, min, max, oneof
type Validator struct{}

// NewValidator creates a new validator instance.
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateConfig validates a populated config struct against its validate tags.
// Returns an ErrorReport with field-level validation errors, or nil if all valid.
func (v *Validator) ValidateConfig(cfg any, fields []FieldInfo) *ErrorReport {
	report := &ErrorReport{}

	val := reflect.ValueOf(cfg)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Validate each field that has a validate tag
	for _, field := range fields {
		validateTag := field.Tags["validate"]
		if validateTag == "" {
			continue
		}

		fieldVal := getFieldByPath(val, field.Path)
		if !fieldVal.IsValid() {
			continue
		}

		// Parse and run validation rules
		rules := parseValidationRules(validateTag)
		for _, rule := range rules {
			fieldErr := v.validateField(fieldVal, field, rule)
			if fieldErr.Message != "" { // Non-empty message means validation failed
				report.AddError(fieldErr)
				break // Stop at first error per field
			}
		}
	}

	return report
}

// ValidationRule represents a single validation constraint
type ValidationRule struct {
	Name  string // "required", "min", "max", "oneof", etc.
	Value string // the constraint value (e.g., "1" for min=1)
}

// parseValidationRules parses a validate tag into individual rules
// Example: "required,min=1,max=65535" -> [required, min=1, max=65535]
func parseValidationRules(tag string) []ValidationRule {
	var rules []ValidationRule

	for _, part := range strings.Split(tag, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if idx := strings.Index(part, "="); idx != -1 {
			rules = append(rules, ValidationRule{
				Name:  part[:idx],
				Value: part[idx+1:],
			})
		} else {
			rules = append(rules, ValidationRule{
				Name: part,
			})
		}
	}

	return rules
}

// validateField runs a single validation rule on a field value
func (v *Validator) validateField(fieldVal reflect.Value, field FieldInfo, rule ValidationRule) FieldError {
	switch rule.Name {
	case "required":
		return v.validateRequired(fieldVal, field, rule)
	case "min":
		return v.validateMin(fieldVal, field, rule)
	case "max":
		return v.validateMax(fieldVal, field, rule)
	case "oneof":
		return v.validateOneOf(fieldVal, field, rule)
	default:
		// Unknown rule - skip
		return FieldError{}
	}
}

// validateRequired checks if a field is provided (non-zero)
func (v *Validator) validateRequired(fieldVal reflect.Value, field FieldInfo, rule ValidationRule) FieldError {
	// Check if value is zero/empty
	if isZeroValue(fieldVal) {
		return FieldError{
			Path:    field.Path,
			Kind:    ErrorKindValidation,
			Rule:    "required",
			Secret:  field.IsSecret,
			Value:   fieldValueToString(fieldVal, field.IsSecret),
			Source:  "validation",
			Message: "field is required",
		}
	}
	return FieldError{}
}

// validateMin checks if a value meets minimum constraint
func (v *Validator) validateMin(fieldVal reflect.Value, field FieldInfo, rule ValidationRule) FieldError {
	minVal, err := strconv.ParseFloat(rule.Value, 64)
	if err != nil {
		return FieldError{} // Invalid constraint
	}

	switch fieldVal.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(fieldVal.Int()) < minVal {
			return FieldError{
				Path:    field.Path,
				Kind:    ErrorKindValidation,
				Rule:    "min",
				Secret:  field.IsSecret,
				Value:   fieldValueToString(fieldVal, field.IsSecret),
				Source:  "validation",
				Message: fmt.Sprintf("must be at least %v", minVal),
			}
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if float64(fieldVal.Uint()) < minVal {
			return FieldError{
				Path:    field.Path,
				Kind:    ErrorKindValidation,
				Rule:    "min",
				Secret:  field.IsSecret,
				Value:   fieldValueToString(fieldVal, field.IsSecret),
				Source:  "validation",
				Message: fmt.Sprintf("must be at least %v", minVal),
			}
		}

	case reflect.Float32, reflect.Float64:
		if fieldVal.Float() < minVal {
			return FieldError{
				Path:    field.Path,
				Kind:    ErrorKindValidation,
				Rule:    "min",
				Secret:  field.IsSecret,
				Value:   fieldValueToString(fieldVal, field.IsSecret),
				Source:  "validation",
				Message: fmt.Sprintf("must be at least %v", minVal),
			}
		}

	case reflect.String:
		// For strings, min is the minimum length
		if len(fieldVal.String()) < int(minVal) {
			return FieldError{
				Path:    field.Path,
				Kind:    ErrorKindValidation,
				Rule:    "min",
				Secret:  field.IsSecret,
				Value:   fieldValueToString(fieldVal, field.IsSecret),
				Source:  "validation",
				Message: fmt.Sprintf("must be at least %v characters", int(minVal)),
			}
		}
	}

	return FieldError{}
}

// validateMax checks if a value meets maximum constraint
func (v *Validator) validateMax(fieldVal reflect.Value, field FieldInfo, rule ValidationRule) FieldError {
	maxVal, err := strconv.ParseFloat(rule.Value, 64)
	if err != nil {
		return FieldError{} // Invalid constraint
	}

	switch fieldVal.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(fieldVal.Int()) > maxVal {
			return FieldError{
				Path:    field.Path,
				Kind:    ErrorKindValidation,
				Rule:    "max",
				Secret:  field.IsSecret,
				Value:   fieldValueToString(fieldVal, field.IsSecret),
				Source:  "validation",
				Message: fmt.Sprintf("must be at most %v", maxVal),
			}
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if float64(fieldVal.Uint()) > maxVal {
			return FieldError{
				Path:    field.Path,
				Kind:    ErrorKindValidation,
				Rule:    "max",
				Secret:  field.IsSecret,
				Value:   fieldValueToString(fieldVal, field.IsSecret),
				Source:  "validation",
				Message: fmt.Sprintf("must be at most %v", maxVal),
			}
		}

	case reflect.Float32, reflect.Float64:
		if fieldVal.Float() > maxVal {
			return FieldError{
				Path:    field.Path,
				Kind:    ErrorKindValidation,
				Rule:    "max",
				Secret:  field.IsSecret,
				Value:   fieldValueToString(fieldVal, field.IsSecret),
				Source:  "validation",
				Message: fmt.Sprintf("must be at most %v", maxVal),
			}
		}

	case reflect.String:
		// For strings, max is the maximum length
		if len(fieldVal.String()) > int(maxVal) {
			return FieldError{
				Path:    field.Path,
				Kind:    ErrorKindValidation,
				Rule:    "max",
				Secret:  field.IsSecret,
				Value:   fieldValueToString(fieldVal, field.IsSecret),
				Source:  "validation",
				Message: fmt.Sprintf("must be at most %v characters", int(maxVal)),
			}
		}
	}

	return FieldError{}
}

// validateOneOf checks if a value is one of allowed options
func (v *Validator) validateOneOf(fieldVal reflect.Value, field FieldInfo, rule ValidationRule) FieldError {
	allowed := strings.Split(rule.Value, ",")
	fieldStr := fieldValueToString(fieldVal, false) // Don't redact for comparison

	for _, opt := range allowed {
		if strings.TrimSpace(opt) == fieldStr {
			return FieldError{} // Valid
		}
	}

	return FieldError{
		Path:    field.Path,
		Kind:    ErrorKindValidation,
		Rule:    "oneof",
		Secret:  field.IsSecret,
		Value:   fieldValueToString(fieldVal, field.IsSecret),
		Source:  "validation",
		Message: fmt.Sprintf("must be one of: %s", rule.Value),
	}
}

// isZeroValue checks if a field has a zero/empty value
func isZeroValue(val reflect.Value) bool {
	switch val.Kind() {
	case reflect.String:
		return val.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return val.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return val.Float() == 0
	case reflect.Bool:
		return val.Bool() == false
	case reflect.Slice, reflect.Array:
		return val.Len() == 0
	case reflect.Ptr:
		return val.IsNil()
	default:
		return val.IsZero()
	}
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
