package confkit

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/MimoJanra/confkit/structtags"
	"gopkg.in/yaml.v3"
)

type DumpFormat int

const (
	FormatJSON DumpFormat = iota
	FormatYAML
)

type DumpOption func(*dumpOptions)

type dumpOptions struct {
	format       DumpFormat
	redactSecret bool
}

func WithDumpFormat(f DumpFormat) DumpOption {
	return func(o *dumpOptions) { o.format = f }
}

func WithDumpRedactSecrets(redact bool) DumpOption {
	return func(o *dumpOptions) { o.redactSecret = redact }
}

func Dump[T any](cfg T, opts ...DumpOption) ([]byte, error) {
	o := &dumpOptions{format: FormatJSON, redactSecret: true}
	for _, opt := range opts {
		opt(o)
	}
	m := buildDumpMap(reflect.ValueOf(cfg), o.redactSecret, o.format)
	if o.format == FormatYAML {
		return yaml.Marshal(m)
	}
	return json.MarshalIndent(m, "", "  ")
}

func DumpString[T any](cfg T, opts ...DumpOption) string {
	b, err := Dump(cfg, opts...)
	if err != nil {
		return fmt.Sprintf("<dump error: %v>", err)
	}
	return string(b)
}

func DumpYAML[T any](cfg T, opts ...DumpOption) ([]byte, error) {
	allOpts := make([]DumpOption, 0, len(opts)+1)
	allOpts = append(allOpts, WithDumpFormat(FormatYAML))
	allOpts = append(allOpts, opts...)
	return Dump(cfg, allOpts...)
}

func buildDumpMap(v reflect.Value, redact bool, format DumpFormat) map[string]any {
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}

	t := v.Type()
	result := make(map[string]any, t.NumField())

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		key, ok := dumpFieldKey(field, format)
		if !ok {
			continue
		}
		isSecret := field.Tag.Get("secret") == "true"
		result[key] = dumpFieldValue(v.Field(i), isSecret, redact, format)
	}
	return result
}

func dumpFieldValue(fv reflect.Value, isSecret, redact bool, format DumpFormat) any {
	for fv.Kind() == reflect.Pointer {
		if fv.IsNil() {
			return nil
		}
		fv = fv.Elem()
	}
	if isSecret && redact {
		return "***REDACTED***"
	}
	switch fv.Kind() {
	case reflect.Struct:
		if structtags.IsSpecialType(fv.Type()) {
			return fv.Interface()
		}
		return buildDumpMap(fv, redact, format)
	case reflect.Slice:
		return dumpSliceValue(fv, redact, format)
	case reflect.Map:
		return dumpMapValue(fv, redact, format)
	default:
		return fv.Interface()
	}
}

func dumpSliceValue(fv reflect.Value, redact bool, format DumpFormat) any {
	result := make([]any, fv.Len())
	for i := range fv.Len() {
		result[i] = dumpFieldValue(fv.Index(i), false, redact, format)
	}
	return result
}

func dumpMapValue(fv reflect.Value, redact bool, format DumpFormat) any {
	result := make(map[string]any, fv.Len())
	for _, k := range fv.MapKeys() {
		result[fmt.Sprintf("%v", k.Interface())] = dumpFieldValue(fv.MapIndex(k), false, redact, format)
	}
	return result
}

// dumpFieldKey returns (key, true) for fields to include, ("", false) for fields to skip.
// Handles json:"-" (skip), json:",omitempty" (empty name → fallback), etc.
// Tag priority: yaml→json→toml for YAML format; json→yaml→toml otherwise.
func dumpFieldKey(field reflect.StructField, format DumpFormat) (string, bool) {
	tagOrder := []string{"json", "yaml", "toml"}
	if format == FormatYAML {
		tagOrder = []string{"yaml", "json", "toml"}
	}
	for _, tag := range tagOrder {
		v := field.Tag.Get(tag)
		if v == "" {
			continue
		}
		name := strings.Split(v, ",")[0]
		if name == "-" {
			return "", false
		}
		if name != "" {
			return name, true
		}

	}
	return structtags.SnakeCase(field.Name), true
}
