package confkit

import (
	"reflect"
	"strings"
)

// FieldInfo represents metadata about a struct field extracted from tags.
type FieldInfo struct {
	Name        string            // Go struct field name
	Type        reflect.Type      // Field type
	Value       reflect.Value     // Current value (for walking structs)
	Path        string            // Dot-separated path (e.g. "server.port")
	Tags        map[string]string // Parsed tags (env, flag, yaml, json, default, validate, secret, desc)
	IsSecret    bool              // Whether field should be redacted
	HasDefault  bool              // Whether a default was specified
	IsNested    bool              // Whether this is a nested struct
}

// ScanFields walks a struct type and returns FieldInfo for each field.
func ScanFields(v any) []FieldInfo {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	typ := val.Type()
	return scanFieldsRecursive(val, typ, "")
}

func scanFieldsRecursive(val reflect.Value, typ reflect.Type, pathPrefix string) []FieldInfo {
	var fields []FieldInfo

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		name := field.Name
		path := name
		if pathPrefix != "" {
			path = pathPrefix + "." + name
		}

		// Parse struct tags
		tags := parseStructTags(field.Tag)

		info := FieldInfo{
			Name:       name,
			Type:       field.Type,
			Value:      fieldVal,
			Path:       path,
			Tags:       tags,
			IsSecret:   tags["secret"] == "true",
			HasDefault: tags["default"] != "",
		}

		// Check if it's a nested struct
		fieldType := field.Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		if fieldType.Kind() == reflect.Struct && !isSpecialType(fieldType) {
			info.IsNested = true
			// Recursively scan nested fields
			if fieldVal.Kind() == reflect.Ptr && fieldVal.IsNil() {
				fieldVal = reflect.New(fieldType)
			}
			if fieldVal.Kind() == reflect.Ptr {
				fieldVal = fieldVal.Elem()
			}
			nestedFields := scanFieldsRecursive(fieldVal, fieldType, path)
			fields = append(fields, nestedFields...)
			continue
		}

		fields = append(fields, info)
	}

	return fields
}

// parseStructTags parses common config tags from a struct field.
func parseStructTags(tag reflect.StructTag) map[string]string {
	result := make(map[string]string)

	// Common tag names
	tagNames := []string{"env", "flag", "yaml", "json", "default", "validate", "secret", "desc"}

	for _, tagName := range tagNames {
		if val, ok := tag.Lookup(tagName); ok {
			// For yaml and json, extract the field name (before comma)
			if tagName == "yaml" || tagName == "json" {
				val = strings.Split(val, ",")[0]
			}
			result[tagName] = val
		}
	}

	return result
}

// isSpecialType checks if a type is a special non-struct type (time.Time, time.Duration, etc).
func isSpecialType(typ reflect.Type) bool {
	return typ.PkgPath() == "time" && (typ.Name() == "Time" || typ.Name() == "Duration")
}
