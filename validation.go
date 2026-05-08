package confkit

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type CustomValidatorFunc func(reflect.Value) error

type Validator struct {
	LocalValidators map[string]CustomValidatorFunc
}

func NewValidator() *Validator {
	return &Validator{
		LocalValidators: make(map[string]CustomValidatorFunc),
	}
}

func (v *Validator) ValidateConfig(cfg any, fields []FieldInfo) *ErrorReport {
	report := &ErrorReport{}

	val := reflect.ValueOf(cfg)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	for _, field := range fields {
		validateTag := field.Tags["validate"]
		if validateTag == "" {
			continue
		}

		fieldVal := getFieldByPath(val, field.Path)
		if !fieldVal.IsValid() {
			continue
		}

		rules := parseValidationRules(validateTag)
		for _, rule := range rules {
			fieldErr := v.validateField(fieldVal, field, rule)
			if fieldErr.Message != "" {
				report.AddError(fieldErr)
				break
			}
		}
	}

	return report
}

type ValidationRule struct {
	Name  string
	Value string
}

func parseValidationRules(tag string) []ValidationRule {
	var rules []ValidationRule

	if idx := strings.Index(tag, "oneof="); idx != -1 {
		before := tag[:idx]
		rest := tag[idx+6:]

		knownRules := map[string]bool{"required": true}

		endIdx := -1
		for i, ch := range rest {
			if ch == ',' && i+1 < len(rest) {
				afterComma := rest[i+1:]
				for rule := range knownRules {
					if strings.HasPrefix(afterComma, rule) {
						n := len(rule)
						if len(afterComma) == n || (len(afterComma) > n && (afterComma[n] == ',' || afterComma[n] == '=')) {
							endIdx = i
							break
						}
					}
				}
				if endIdx != -1 {
					break
				}
				if eq := strings.Index(afterComma, "="); eq > 0 && eq < 20 && !strings.Contains(afterComma[:eq], ",") {
					endIdx = i
					break
				}
			}
		}

		var oneofValue, tagWithoutOneof string
		if endIdx == -1 {
			oneofValue = rest
		} else {
			oneofValue = rest[:endIdx]
			rest = rest[endIdx+1:]
			if before != "" && before != "," {
				tagWithoutOneof = strings.TrimRight(before, ",") + "," + rest
			} else {
				tagWithoutOneof = rest
			}
		}
		if tagWithoutOneof == "" {
			tagWithoutOneof = before
		}

		for _, part := range strings.Split(tagWithoutOneof, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			if eq := strings.Index(part, "="); eq != -1 {
				rules = append(rules, ValidationRule{Name: part[:eq], Value: part[eq+1:]})
			} else {
				rules = append(rules, ValidationRule{Name: part})
			}
		}
		if oneofValue != "" {
			rules = append(rules, ValidationRule{Name: "oneof", Value: oneofValue})
		}
	} else {
		for _, part := range strings.Split(tag, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			if eq := strings.Index(part, "="); eq != -1 {
				rules = append(rules, ValidationRule{Name: part[:eq], Value: part[eq+1:]})
			} else {
				rules = append(rules, ValidationRule{Name: part})
			}
		}
	}

	return rules
}

func (v *Validator) validateField(fieldVal reflect.Value, field FieldInfo, rule ValidationRule) FieldError {
	switch rule.Name {
	case "required":
		return v.validateRequired(fieldVal, field)
	case "min":
		return v.validateMin(fieldVal, field, rule)
	case "max":
		return v.validateMax(fieldVal, field, rule)
	case "oneof":
		return v.validateOneOf(fieldVal, field, rule)
	default:
		if fe, ok := v.validateBuiltin(fieldVal, field, rule); ok {
			return fe
		}
		if customFn, exists := v.LocalValidators[rule.Name]; exists {
			if err := customFn(fieldVal); err != nil {
				return FieldError{
					Path:    field.Path,
					Kind:    ErrorKindValidation,
					Rule:    rule.Name,
					Secret:  field.IsSecret,
					Value:   fieldValueToString(fieldVal, field.IsSecret),
					Source:  "validation",
					Message: err.Error(),
				}
			}
		}
		return FieldError{}
	}
}

func (v *Validator) validateRequired(fieldVal reflect.Value, field FieldInfo) FieldError {
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

func (v *Validator) validateMin(fieldVal reflect.Value, field FieldInfo, rule ValidationRule) FieldError {
	minVal, err := strconv.ParseFloat(rule.Value, 64)
	if err != nil {
		return FieldError{}
	}

	var failed bool
	var msg string
	switch fieldVal.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		failed = float64(fieldVal.Int()) < minVal
		msg = fmt.Sprintf("must be at least %v", minVal)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		failed = float64(fieldVal.Uint()) < minVal
		msg = fmt.Sprintf("must be at least %v", minVal)
	case reflect.Float32, reflect.Float64:
		failed = fieldVal.Float() < minVal
		msg = fmt.Sprintf("must be at least %v", minVal)
	case reflect.String:
		failed = len(fieldVal.String()) < int(minVal)
		msg = fmt.Sprintf("must be at least %v characters", int(minVal))
	}

	if failed {
		return FieldError{
			Path: field.Path, Kind: ErrorKindValidation, Rule: "min",
			Secret: field.IsSecret, Value: fieldValueToString(fieldVal, field.IsSecret),
			Source: "validation", Message: msg,
		}
	}
	return FieldError{}
}

func (v *Validator) validateMax(fieldVal reflect.Value, field FieldInfo, rule ValidationRule) FieldError {
	maxVal, err := strconv.ParseFloat(rule.Value, 64)
	if err != nil {
		return FieldError{}
	}

	var failed bool
	var msg string
	switch fieldVal.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		failed = float64(fieldVal.Int()) > maxVal
		msg = fmt.Sprintf("must be at most %v", maxVal)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		failed = float64(fieldVal.Uint()) > maxVal
		msg = fmt.Sprintf("must be at most %v", maxVal)
	case reflect.Float32, reflect.Float64:
		failed = fieldVal.Float() > maxVal
		msg = fmt.Sprintf("must be at most %v", maxVal)
	case reflect.String:
		failed = len(fieldVal.String()) > int(maxVal)
		msg = fmt.Sprintf("must be at most %v characters", int(maxVal))
	}

	if failed {
		return FieldError{
			Path: field.Path, Kind: ErrorKindValidation, Rule: "max",
			Secret: field.IsSecret, Value: fieldValueToString(fieldVal, field.IsSecret),
			Source: "validation", Message: msg,
		}
	}
	return FieldError{}
}

func (v *Validator) validateOneOf(fieldVal reflect.Value, field FieldInfo, rule ValidationRule) FieldError {
	fieldStr := fieldValueToString(fieldVal, false)
	for _, opt := range strings.Split(rule.Value, ",") {
		if strings.TrimSpace(opt) == fieldStr {
			return FieldError{}
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
		return !val.Bool()
	case reflect.Slice, reflect.Array:
		return val.Len() == 0
	case reflect.Ptr:
		return val.IsNil()
	default:
		return val.IsZero()
	}
}

func getFieldByPath(val reflect.Value, path string) reflect.Value {
	current := val
	for _, part := range splitPath(path) {
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
	default:
		return fmt.Sprintf("%v", val.Interface())
	}
}
