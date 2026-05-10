package confkit

import (
	"fmt"
	"math"
	"net"
	"net/mail"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

var (
	uuidRegex     = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	hostnameRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
)

func (v *Validator) validateBuiltin(fieldVal reflect.Value, field FieldInfo, rule ValidationRule) (FieldError, bool) {
	switch rule.Name {
	case "email":
		return strCheck(fieldVal, field, rule.Name, func(s string) (bool, string) {
			_, err := mail.ParseAddress(s)
			return err == nil, "must be a valid email address"
		}), true

	case "url":
		return strCheck(fieldVal, field, rule.Name, func(s string) (bool, string) {
			u, err := url.ParseRequestURI(s)
			ok := err == nil && u.Scheme != "" && u.Host != ""
			return ok, "must be a valid URL"
		}), true

	case "http_url":
		return strCheck(fieldVal, field, rule.Name, func(s string) (bool, string) {
			u, err := url.ParseRequestURI(s)
			ok := err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
			return ok, "must be a valid HTTP or HTTPS URL"
		}), true

	case "ip":
		return strCheck(fieldVal, field, rule.Name, func(s string) (bool, string) {
			return net.ParseIP(s) != nil, "must be a valid IP address"
		}), true

	case "ipv4":
		return strCheck(fieldVal, field, rule.Name, func(s string) (bool, string) {
			ip := net.ParseIP(s)
			return ip != nil && ip.To4() != nil, "must be a valid IPv4 address"
		}), true

	case "ipv6":
		return strCheck(fieldVal, field, rule.Name, func(s string) (bool, string) {
			ip := net.ParseIP(s)
			return ip != nil && ip.To4() == nil, "must be a valid IPv6 address"
		}), true

	case "uuid":
		return strCheck(fieldVal, field, rule.Name, func(s string) (bool, string) {
			return uuidRegex.MatchString(s), "must be a valid UUID"
		}), true

	case "hostname":
		return strCheck(fieldVal, field, rule.Name, func(s string) (bool, string) {
			return hostnameRegex.MatchString(s) && len(s) <= 253, "must be a valid hostname"
		}), true

	case "port":
		return numCheck(fieldVal, field, rule.Name, func(n int64) (bool, string) {
			return n >= 1 && n <= 65535, fmt.Sprintf("must be a valid port (1-65535), got %d", n)
		}), true

	case "regex":
		if rule.Value == "" {
			return FieldError{}, true
		}
		return strCheck(fieldVal, field, rule.Name, func(s string) (bool, string) {
			re, err := regexp.Compile(rule.Value)
			if err != nil {
				return false, fmt.Sprintf("invalid regex pattern %q", rule.Value)
			}
			return re.MatchString(s), fmt.Sprintf("must match pattern %q", rule.Value)
		}), true

	case "len":
		n, err := strconv.Atoi(rule.Value)
		if err != nil {
			return FieldError{}, true
		}
		return strCheck(fieldVal, field, rule.Name, func(s string) (bool, string) {
			return len([]rune(s)) == n, fmt.Sprintf("must be exactly %d characters, got %d", n, len([]rune(s)))
		}), true

	case "contains":
		return strCheck(fieldVal, field, rule.Name, func(s string) (bool, string) {
			return strings.Contains(s, rule.Value), fmt.Sprintf("must contain %q", rule.Value)
		}), true

	case "startswith":
		return strCheck(fieldVal, field, rule.Name, func(s string) (bool, string) {
			return strings.HasPrefix(s, rule.Value), fmt.Sprintf("must start with %q", rule.Value)
		}), true

	case "endswith":
		return strCheck(fieldVal, field, rule.Name, func(s string) (bool, string) {
			return strings.HasSuffix(s, rule.Value), fmt.Sprintf("must end with %q", rule.Value)
		}), true

	case "alpha":
		return strCheck(fieldVal, field, rule.Name, func(s string) (bool, string) {
			for _, r := range s {
				if !unicode.IsLetter(r) {
					return false, "must contain only letters"
				}
			}
			return true, ""
		}), true

	case "alphanum":
		return strCheck(fieldVal, field, rule.Name, func(s string) (bool, string) {
			for _, r := range s {
				if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
					return false, "must contain only letters and digits"
				}
			}
			return true, ""
		}), true

	case "numeric":
		return strCheck(fieldVal, field, rule.Name, func(s string) (bool, string) {
			for _, r := range s {
				if !unicode.IsDigit(r) {
					return false, "must contain only digits"
				}
			}
			return true, ""
		}), true

	case "lowercase":
		return strCheck(fieldVal, field, rule.Name, func(s string) (bool, string) {
			return s == strings.ToLower(s), "must be all lowercase"
		}), true

	case "uppercase":
		return strCheck(fieldVal, field, rule.Name, func(s string) (bool, string) {
			return s == strings.ToUpper(s), "must be all uppercase"
		}), true

	case "notempty":
		return strCheck(fieldVal, field, rule.Name, func(s string) (bool, string) {
			return strings.TrimSpace(s) != "", "must not be blank"
		}), true
	}

	return FieldError{}, false
}

func strCheck(fieldVal reflect.Value, field FieldInfo, rule string, check func(string) (bool, string)) FieldError {
	if fieldVal.Kind() != reflect.String {
		return FieldError{}
	}
	s := fieldVal.String()
	if s == "" {
		return FieldError{}
	}
	ok, msg := check(s)
	if !ok {
		return FieldError{
			Path:    field.Path,
			Kind:    ErrorKindValidation,
			Rule:    rule,
			Secret:  field.IsSecret,
			Value:   fieldValueToString(fieldVal, field.IsSecret),
			Source:  "validation",
			Message: msg,
		}
	}
	return FieldError{}
}

func numCheck(fieldVal reflect.Value, field FieldInfo, rule string, check func(int64) (bool, string)) FieldError {
	var n int64
	switch fieldVal.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n = fieldVal.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		un := fieldVal.Uint()
		if un > math.MaxInt64 {
			return FieldError{
				Path:    field.Path,
				Kind:    ErrorKindValidation,
				Rule:    rule,
				Secret:  field.IsSecret,
				Value:   fieldValueToString(fieldVal, field.IsSecret),
				Source:  "validation",
				Message: fmt.Sprintf("numeric value %d exceeds max supported value %d", un, int64(math.MaxInt64)),
			}
		}
		n = int64(un)
	case reflect.String:
		var err error
		n, err = strconv.ParseInt(fieldVal.String(), 10, 64)
		if err != nil {
			return FieldError{
				Path:    field.Path,
				Kind:    ErrorKindValidation,
				Rule:    rule,
				Secret:  field.IsSecret,
				Value:   fieldValueToString(fieldVal, field.IsSecret),
				Source:  "validation",
				Message: "must be a numeric value",
			}
		}
	default:
		return FieldError{}
	}
	ok, msg := check(n)
	if !ok {
		return FieldError{
			Path:    field.Path,
			Kind:    ErrorKindValidation,
			Rule:    rule,
			Secret:  field.IsSecret,
			Value:   fieldValueToString(fieldVal, field.IsSecret),
			Source:  "validation",
			Message: msg,
		}
	}
	return FieldError{}
}
