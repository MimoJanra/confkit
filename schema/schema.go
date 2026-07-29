// Package schema derives documentation and machine-readable schemas from a config
// struct, so that the struct definition stays the single source of truth.
//
// GenerateSchema and GenerateSchemaJSON emit JSON Schema for editor completion and
// CI validation, GenerateMarkdown emits a reference table, and GenerateCLIHelp emits
// an options listing. All four read the same `validate`, `default`, `desc`, `secret`,
// `short` and `hidden` tags that confkit uses when loading.
package schema

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/MimoJanra/confkit/structtags"
)

// Schema is a JSON Schema document, or one property within it.
//
// Secret, Short and Hidden are confkit extensions rather than standard JSON Schema
// keywords: they carry the `secret`, `short` and `hidden` tags through to the
// Markdown and CLI-help generators, and are omitted when unset.
type Schema struct {
	Title       string             `json:"title,omitempty"`
	Description string             `json:"description,omitempty"`
	Type        string             `json:"type"`
	Properties  map[string]*Schema `json:"properties,omitempty"`
	Required    []string           `json:"required,omitempty"`
	Items       *Schema            `json:"items,omitempty"`
	Minimum     *int               `json:"minimum,omitempty"`
	Maximum     *int               `json:"maximum,omitempty"`
	MinLength   *int               `json:"minLength,omitempty"`
	MaxLength   *int               `json:"maxLength,omitempty"`
	Pattern     string             `json:"pattern,omitempty"`
	Enum        []interface{}      `json:"enum,omitempty"`
	Default     interface{}        `json:"default,omitempty"`
	Secret      bool               `json:"secret,omitempty"`
	Short       string             `json:"short,omitempty"`
	Hidden      bool               `json:"hidden,omitempty"`
}

// GenerateSchema builds a JSON Schema for T, which must be a struct or a pointer to
// one; anything else, including an interface such as any, is an error.
//
// Property names follow the `json`, `yaml` or `toml` tag and otherwise the
// snake_cased field name. Embedded structs are flattened and fields tagged "-" are
// omitted, so the schema describes the same shape the loader accepts.
func GenerateSchema[T any]() (*Schema, error) {
	var cfg T
	cfgType := reflect.TypeOf(cfg)

	// T may be an interface (e.g. any), for which the zero value carries no type.
	if cfgType == nil {
		return nil, fmt.Errorf("GenerateSchema requires a concrete struct type")
	}
	if cfgType.Kind() == reflect.Pointer {
		cfgType = cfgType.Elem()
	}
	if cfgType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("GenerateSchema requires a struct type, got %v", cfgType.Kind())
	}

	s := &Schema{
		Type:       "object",
		Properties: make(map[string]*Schema),
		Required:   make([]string, 0),
	}

	if err := walkStruct(cfgType, s); err != nil {
		return nil, err
	}

	if len(s.Required) == 0 {
		s.Required = nil
	}

	return s, nil
}

// GenerateSchemaJSON is GenerateSchema marshalled to indented JSON, suitable for
// writing to a schema file that editors and CI can consume.
func GenerateSchemaJSON[T any]() ([]byte, error) {
	s, err := GenerateSchema[T]()
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(s, "", "  ")
}

func walkStruct(typ reflect.Type, parent *Schema) error {
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}

		fieldType := field.Type
		if fieldType.Kind() == reflect.Pointer {
			fieldType = fieldType.Elem()
		}

		// Embedded structs are flattened into the parent by the loader, so the
		// schema must flatten them too or it will not match the real config shape.
		if field.Anonymous && fieldType.Kind() == reflect.Struct && !structtags.IsSpecialType(fieldType) {
			if err := walkStruct(fieldType, parent); err != nil {
				return err
			}
			continue
		}

		tags := structtags.ParseStructTags(field.Tag)
		propName := getPropName(field.Name, tags)
		if propName == "" {
			continue
		}

		fieldSchema := &Schema{}

		if structtags.IsSpecialType(fieldType) {
			fieldSchema.Type = "string"
		} else if fieldType.Kind() == reflect.Struct {
			fieldSchema.Type = "object"
			fieldSchema.Properties = make(map[string]*Schema)
			if err := walkStruct(fieldType, fieldSchema); err != nil {
				return err
			}
		} else if fieldType.Kind() == reflect.Slice {
			fieldSchema.Type = "array"
			elemType := fieldType.Elem()
			if elemType.Kind() == reflect.Pointer {
				elemType = elemType.Elem()
			}
			fieldSchema.Items = &Schema{Type: getTypeString(elemType)}
		} else {
			fieldSchema.Type = getTypeString(fieldType)
		}

		if desc := tags["desc"]; desc != "" {
			fieldSchema.Description = desc
		}
		if def := tags["default"]; def != "" {
			fieldSchema.Default = parseDefaultValue(def, fieldType)
		}
		if tags["secret"] == "true" {
			fieldSchema.Secret = true
		}
		if short := tags["short"]; short != "" {
			fieldSchema.Short = short
		}
		if tags["hidden"] == "true" {
			fieldSchema.Hidden = true
		}
		if tags["validate"] != "" {
			applyValidationRules(tags["validate"], fieldType, fieldSchema, &parent.Required, propName)
		}

		parent.Properties[propName] = fieldSchema
	}

	return nil
}

// getPropName returns the schema property name, or "" for fields the loader
// skips because the winning tag is "-".
func getPropName(fieldName string, tags map[string]string) string {
	for _, tag := range []string{"json", "yaml", "toml"} {
		if name := tags[tag]; name != "" {
			if name == "-" {
				return ""
			}
			return name
		}
	}
	return structtags.SnakeCase(fieldName)
}

func getTypeString(typ reflect.Type) string {
	// time.Time and time.Duration are parsed from strings ("2006-01-02T15:04:05Z",
	// "5s"). Duration's kind is int64, so this must be checked before the switch.
	if structtags.IsSpecialType(typ) {
		return "string"
	}
	if typ.Kind() == reflect.Struct {
		return "object"
	}
	switch typ.Kind() {
	case reflect.String:
		return "string"
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Slice:
		return "array"
	default:
		return "string"
	}
}

func applyValidationRules(validate string, typ reflect.Type, s *Schema, required *[]string, propName string) {
	var rules []string
	var oneofValue string

	if idx := strings.Index(validate, "oneof="); idx != -1 {
		before := validate[:idx]
		rest := validate[idx+6:]

		endIdx := 0
		for i := 0; i < len(rest); i++ {
			if rest[i] == ',' && i+1 < len(rest) {
				afterComma := rest[i+1:]
				eqIdx := strings.Index(afterComma, "=")
				if eqIdx > 0 && eqIdx < 20 && !strings.Contains(afterComma[:eqIdx], ",") {
					endIdx = i
					break
				}
			}
		}

		if endIdx == 0 && !strings.Contains(rest, "=") {
			oneofValue = rest
		} else if endIdx > 0 {
			oneofValue = rest[:endIdx]
			rest = rest[endIdx+1:]
			if before != "" && !strings.HasSuffix(before, ",") {
				before = before + "," + rest
			} else {
				before = rest
			}
		} else {
			for i := 0; i < len(rest); i++ {
				if rest[i] == ',' {
					afterComma := strings.TrimSpace(rest[i+1:])
					if eqPos := strings.Index(afterComma, "="); eqPos > 0 && !strings.Contains(afterComma[:eqPos], ",") {
						endIdx = i
						break
					}
				}
			}
			if endIdx > 0 {
				oneofValue = strings.TrimSpace(rest[:endIdx])
				rest = strings.TrimSpace(rest[endIdx+1:])
				if before != "" {
					before = strings.TrimRight(before, ",") + "," + rest
				} else {
					before = rest
				}
			} else {
				oneofValue = rest
			}
		}

		if before != "" && before != "," {
			rules = strings.Split(strings.TrimRight(before, ","), ",")
		}
	} else {
		rules = strings.Split(validate, ",")
	}

	for _, rule := range rules {
		rule = strings.TrimSpace(rule)
		if rule == "" {
			continue
		}
		if rule == "required" {
			*required = append(*required, propName)
			continue
		}
		parts := strings.SplitN(rule, "=", 2)
		if len(parts) != 2 {
			continue
		}
		name, val := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		switch name {
		case "min":
			if n, err := strconv.Atoi(val); err == nil {
				if isNumericType(typ) {
					s.Minimum = &n
				} else if typ.Kind() == reflect.String {
					s.MinLength = &n
				}
			}
		case "max":
			if n, err := strconv.Atoi(val); err == nil {
				if isNumericType(typ) {
					s.Maximum = &n
				} else if typ.Kind() == reflect.String {
					s.MaxLength = &n
				}
			}
		case "pattern":
			s.Pattern = val
		}
	}

	if oneofValue != "" {
		for _, v := range strings.Split(oneofValue, ",") {
			if v = strings.TrimSpace(v); v != "" {
				s.Enum = append(s.Enum, v)
			}
		}
	}
}

func isNumericType(typ reflect.Type) bool {
	switch typ.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

func parseDefaultValue(defStr string, typ reflect.Type) interface{} {
	switch typ.Kind() {
	case reflect.String:
		return defStr
	case reflect.Bool:
		return defStr == "true" || defStr == "1" || defStr == "yes"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val, err := strconv.ParseInt(defStr, 10, 64); err == nil {
			return val
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if val, err := strconv.ParseUint(defStr, 10, 64); err == nil {
			return val
		}
	case reflect.Float32, reflect.Float64:
		if val, err := strconv.ParseFloat(defStr, 64); err == nil {
			return val
		}
	}
	return defStr
}

// GenerateMarkdown renders T as a Markdown table of fields, types, defaults, rules
// and descriptions, for pasting into a README. Rows are sorted by field name, nested
// structs follow their parent with dotted names, and cell values are escaped.
func GenerateMarkdown[T any]() (string, error) {
	s, err := GenerateSchema[T]()
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString("## Configuration Fields\n\n")
	sb.WriteString("| Field | Type | Default | Rules | Description |\n")
	sb.WriteString("|-------|------|---------|-------|-------------|\n")
	addMarkdownProperties(&sb, s.Properties, "")
	return sb.String(), nil
}

func addMarkdownProperties(sb *strings.Builder, props map[string]*Schema, prefix string) {
	keys := make([]string, 0, len(props))
	for k := range props {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, propName := range keys {
		prop := props[propName]
		fieldName := propName
		if prefix != "" {
			fieldName = prefix + "." + propName
		}
		fmt.Fprintf(sb, "| %s | %s | %s | %s | %s |\n",
			escapeMarkdownCell(fieldName),
			escapeMarkdownCell(getMarkdownType(prop.Type)),
			escapeMarkdownCell(getMarkdownDefault(prop.Default)),
			escapeMarkdownCell(getMarkdownRules(prop)),
			escapeMarkdownCell(prop.Description),
		)
		if prop.Type == "object" && prop.Properties != nil {
			addMarkdownProperties(sb, prop.Properties, fieldName)
		}
	}
}

// GenerateCLIHelp renders T as a --help style options listing.
//
// Flags are kebab-cased and nested structs contribute dash-joined names
// ("--db-host"). A `short` tag adds a single-letter alias, `hidden:"true"` omits the
// field, and defaults, bounds and required markers are appended.
func GenerateCLIHelp[T any]() (string, error) {
	s, err := GenerateSchema[T]()
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString("Usage: [OPTIONS]\n\nOptions:\n")
	addCLIHelpOptions(&sb, s.Properties, "", s.Required)
	return sb.String(), nil
}

func addCLIHelpOptions(sb *strings.Builder, props map[string]*Schema, prefix string, required []string) {
	keys := make([]string, 0, len(props))
	for k := range props {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, propName := range keys {
		prop := props[propName]
		if prop.Hidden {
			continue
		}

		kebabName := strings.ReplaceAll(propName, "_", "-")
		flagName := kebabName
		if prefix != "" {
			flagName = prefix + "-" + kebabName
		}

		if prop.Type == "object" && prop.Properties != nil {
			addCLIHelpOptions(sb, prop.Properties, flagName, prop.Required)
			continue
		}

		line := ""
		if prop.Short != "" {
			line = fmt.Sprintf("  -%s, --%s", prop.Short, flagName)
		} else {
			line = fmt.Sprintf("  --%s", flagName)
		}

		switch prop.Type {
		case "string":
			line += " VALUE"
		case "integer":
			line += " NUMBER"
		case "number":
			line += " FLOAT"
		case "boolean":
			line += " BOOL"
		}

		if prop.Description != "" {
			line += "  " + prop.Description
		}

		hasConstraints := prop.Minimum != nil || prop.Maximum != nil || prop.MinLength != nil || prop.MaxLength != nil
		if prop.Default != nil || hasConstraints {
			var parts []string
			if prop.Default != nil {
				parts = append(parts, fmt.Sprintf("default: %v", prop.Default))
			}
			if prop.Minimum != nil {
				parts = append(parts, fmt.Sprintf("min: %v", *prop.Minimum))
			}
			if prop.Maximum != nil {
				parts = append(parts, fmt.Sprintf("max: %v", *prop.Maximum))
			}
			if prop.MinLength != nil {
				parts = append(parts, fmt.Sprintf("min length: %v", *prop.MinLength))
			}
			if prop.MaxLength != nil {
				parts = append(parts, fmt.Sprintf("max length: %v", *prop.MaxLength))
			}
			line += " (" + strings.Join(parts, ", ") + ")"
		}

		for _, req := range required {
			if req == propName {
				line += " (required)"
				break
			}
		}

		sb.WriteString(line + "\n")
	}
}

// escapeMarkdownCell escapes pipes so that values such as a regex alternation
// ("^(dev|prod)$") do not split the surrounding table row into extra columns.
func escapeMarkdownCell(s string) string {
	return strings.ReplaceAll(s, "|", "\\|")
}

func getMarkdownType(schemaType string) string {
	switch schemaType {
	case "integer":
		return "int"
	case "number":
		return "float"
	default:
		return schemaType
	}
}

func getMarkdownDefault(def interface{}) string {
	if def == nil {
		return "—"
	}
	return fmt.Sprintf("%v", def)
}

func getMarkdownRules(s *Schema) string {
	var rules []string
	if s.Minimum != nil {
		rules = append(rules, fmt.Sprintf("min=%v", *s.Minimum))
	}
	if s.Maximum != nil {
		rules = append(rules, fmt.Sprintf("max=%v", *s.Maximum))
	}
	if s.MinLength != nil {
		rules = append(rules, fmt.Sprintf("min=%v chars", *s.MinLength))
	}
	if s.MaxLength != nil {
		rules = append(rules, fmt.Sprintf("max=%v chars", *s.MaxLength))
	}
	if len(s.Enum) > 0 {
		strs := make([]string, len(s.Enum))
		for i, e := range s.Enum {
			strs[i] = fmt.Sprintf("%v", e)
		}
		rules = append(rules, "oneof="+strings.Join(strs, ","))
	}
	if s.Pattern != "" {
		rules = append(rules, "pattern="+s.Pattern)
	}
	if s.Secret {
		rules = append(rules, "secret")
	}
	if len(rules) == 0 {
		return "—"
	}
	return strings.Join(rules, ", ")
}
