package schema

import (
	"encoding/json"
	"fmt"
	"github.com/MimoJanra/confkit/tagutil"
	"reflect"
	"strconv"
	"strings"
)

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

func GenerateSchema[T any]() (*Schema, error) {
	var cfg T
	cfgType := reflect.TypeOf(cfg)

	if cfgType.Kind() == reflect.Ptr {
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

		tags := tagutil.ParseStructTags(field.Tag)
		propName := getPropName(field.Name, tags)
		if propName == "" {
			continue
		}

		fieldSchema := &Schema{}
		fieldType := field.Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		if tagutil.IsSpecialType(fieldType) {
			fieldSchema.Type = "string"
		} else if fieldType.Kind() == reflect.Struct {
			fieldSchema.Type = "object"
			fieldSchema.Properties = make(map[string]*Schema)
			walkStruct(fieldType, fieldSchema)
		} else if fieldType.Kind() == reflect.Slice {
			fieldSchema.Type = "array"
			elemType := fieldType.Elem()
			if elemType.Kind() == reflect.Ptr {
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

func getPropName(fieldName string, tags map[string]string) string {
	for _, tag := range []string{"json", "yaml", "toml"} {
		if name := tags[tag]; name != "" {
			return name
		}
	}
	return tagutil.SnakeCase(fieldName)
}

func getTypeString(typ reflect.Type) string {
	if typ.Kind() == reflect.Struct {
		if typ.PkgPath() == "time" && (typ.Name() == "Duration" || typ.Name() == "Time") {
			return "string"
		}
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

// applyValidationRules applies validate tag constraints to the schema.
// oneof values contain commas (e.g. oneof=debug,info,warn) which conflict with the
// rule separator, so they are extracted before splitting.
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
	for propName, prop := range props {
		fieldName := propName
		if prefix != "" {
			fieldName = prefix + "." + propName
		}
		fmt.Fprintf(sb, "| %s | %s | %s | %s | %s |\n",
			fieldName,
			getMarkdownType(prop.Type),
			getMarkdownDefault(prop.Default),
			getMarkdownRules(prop),
			prop.Description,
		)
		if prop.Type == "object" && prop.Properties != nil {
			addMarkdownProperties(sb, prop.Properties, fieldName)
		}
	}
}

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
	for propName, prop := range props {
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
