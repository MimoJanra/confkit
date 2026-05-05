package confkit

import (
	"github.com/MimoJanra/confkit/tagutil"
	"reflect"
)

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

func ScanFields(v any) []FieldInfo {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	return scanFieldsRecursive(val, val.Type(), "", nil)
}

func scanFieldsRecursive(val reflect.Value, typ reflect.Type, pathPrefix string, ancestorTags []map[string]string) []FieldInfo {
	var fields []FieldInfo

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		if !field.IsExported() {
			continue
		}

		name := field.Name
		path := name
		if pathPrefix != "" {
			path = pathPrefix + "." + name
		}

		tags := tagutil.ParseStructTags(field.Tag)

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

		fieldType := field.Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		if fieldType.Kind() == reflect.Struct && !tagutil.IsSpecialType(fieldType) {
			info.IsNested = true
			if fieldVal.Kind() == reflect.Ptr && fieldVal.IsNil() {
				fieldVal = reflect.New(fieldType)
			}
			if fieldVal.Kind() == reflect.Ptr {
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
