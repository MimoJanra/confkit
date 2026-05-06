package structtags

import (
	"reflect"
	"strings"
)

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

func IsSpecialType(typ reflect.Type) bool {
	return typ.PkgPath() == "time" && (typ.Name() == "Time" || typ.Name() == "Duration")
}

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
