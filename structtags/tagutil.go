// Package structtags provides the struct-tag and field-name helpers shared by
// confkit's sources, validation and schema generation, so that every part of the
// library derives names and tag values the same way.
package structtags

import (
	"reflect"
	"strings"
)

// SnakeCase converts a Go field name to snake_case, used as the fallback key when a
// field carries no format tag.
//
// Runs of capitals are kept together, so "HTTPPort" becomes "http_port", "APIKey"
// becomes "api_key" and "ID" becomes "id".
func SnakeCase(s string) string {
	runes := []rune(s)
	var result strings.Builder
	for i, r := range runes {
		if i > 0 && r >= 'A' && r <= 'Z' {
			prev := runes[i-1]
			if prev >= 'a' && prev <= 'z' {
				result.WriteRune('_')
			} else if prev >= 'A' && prev <= 'Z' && i+1 < len(runes) {
				next := runes[i+1]
				if next >= 'a' && next <= 'z' {
					result.WriteRune('_')
				}
			}
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// IsSpecialType reports whether typ is time.Time or time.Duration.
//
// Both parse from a single string ("2006-01-02T15:04:05Z", "5s"), so they are treated
// as scalars rather than being walked as a struct or an integer.
func IsSpecialType(typ reflect.Type) bool {
	return typ.PkgPath() == "time" && (typ.Name() == "Time" || typ.Name() == "Duration")
}

// ParseStructTags extracts the tags confkit understands into a map, omitting any that
// are absent.
//
// For the yaml, json and toml tags only the name is kept, so `json:"port,omitempty"`
// yields "port"; a name of "-" is preserved so callers can treat it as a skip
// directive.
func ParseStructTags(tag reflect.StructTag) map[string]string {
	result := make(map[string]string)
	for _, name := range []string{"env", "flag", "yaml", "json", "toml", "default", "validate", "secret", "desc", "prefix", "short", "hidden"} {
		if val, ok := tag.Lookup(name); ok {
			if name == "yaml" || name == "json" || name == "toml" {
				val = strings.Split(val, ",")[0]
			}
			result[name] = val
		}
	}
	return result
}
