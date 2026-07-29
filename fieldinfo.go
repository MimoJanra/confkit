package confkit

import (
	"reflect"

	"github.com/MimoJanra/confkit/structtags"
)

// FieldInfo describes a single configuration field discovered by reflection.
//
// Struct tags are parsed once into FieldInfo and then reused by every Source, so
// sources need no knowledge of tag syntax. Path is the dotted Go field path
// ("DB.Host") and is the key used throughout loading, validation and errors.
// AncestorTags holds the tags of each enclosing struct, which is how env
// prefixes accumulate.
type FieldInfo struct {
	Name         string
	Type         reflect.Type
	Value        reflect.Value
	Path         string
	Tags         map[string]string
	AncestorTags []map[string]string
	IsSecret     bool
	HasDefault   bool
	IsNested     bool
}

// ScanFields walks v, which must be a struct or a pointer to one, and returns a
// FieldInfo for every exported leaf field.
//
// Nested structs contribute their fields with a dotted Path and are not returned
// themselves; embedded structs are flattened into the enclosing struct, matching
// the way encoding/json promotes them.
func ScanFields(v any) []FieldInfo {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	return scanFieldsRecursive(val, val.Type(), "", nil)
}

func initEmbeddedPointers(val reflect.Value, typ reflect.Type) {
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)
		if !field.IsExported() {
			continue
		}
		if field.Anonymous && field.Type.Kind() == reflect.Pointer {
			elemType := field.Type.Elem()
			if elemType.Kind() == reflect.Struct && !structtags.IsSpecialType(elemType) {
				if fieldVal.IsNil() {
					fieldVal.Set(reflect.New(elemType))
				}
				initEmbeddedPointers(fieldVal.Elem(), elemType)
			}
		} else if field.Type.Kind() == reflect.Struct && !structtags.IsSpecialType(field.Type) {
			initEmbeddedPointers(fieldVal, field.Type)
		}
	}
}

func scanFieldsRecursive(val reflect.Value, typ reflect.Type, pathPrefix string, ancestorTags []map[string]string) []FieldInfo {
	var fields []FieldInfo

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		if !field.IsExported() {
			continue
		}

		fieldType := field.Type
		if fieldType.Kind() == reflect.Pointer {
			fieldType = fieldType.Elem()
		}

		if field.Anonymous && fieldType.Kind() == reflect.Struct && !structtags.IsSpecialType(fieldType) {
			if fieldVal.Kind() == reflect.Pointer {
				if fieldVal.IsNil() {
					fieldVal = reflect.New(fieldType)
				}
				fieldVal = fieldVal.Elem()
			}
			fields = append(fields, scanFieldsRecursive(fieldVal, fieldType, pathPrefix, ancestorTags)...)
			continue
		}

		name := field.Name
		path := name
		if pathPrefix != "" {
			path = pathPrefix + "." + name
		}

		tags := structtags.ParseStructTags(field.Tag)

		info := FieldInfo{
			Name:         name,
			Type:         field.Type,
			Value:        fieldVal,
			Path:         path,
			Tags:         tags,
			AncestorTags: ancestorTags,
			IsSecret:     tags["secret"] == "true",
			HasDefault:   tags["default"] != "",
		}

		if fieldType.Kind() == reflect.Struct && !structtags.IsSpecialType(fieldType) {
			info.IsNested = true
			if fieldVal.Kind() == reflect.Pointer && fieldVal.IsNil() {
				fieldVal = reflect.New(fieldType)
			}
			if fieldVal.Kind() == reflect.Pointer {
				fieldVal = fieldVal.Elem()
			}
			newAncestors := make([]map[string]string, len(ancestorTags)+1)
			copy(newAncestors, ancestorTags)
			newAncestors[len(ancestorTags)] = tags
			fields = append(fields, scanFieldsRecursive(fieldVal, fieldType, path, newAncestors)...)
			continue
		}

		fields = append(fields, info)
	}

	return fields
}
