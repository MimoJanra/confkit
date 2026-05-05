package confkit

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(value string, targetType reflect.Type) (any, error) {
	if value == "" {
		return p.zeroValue(targetType), nil
	}

	if targetType.PkgPath() == "time" {
		if targetType.Name() == "Duration" {
			return parseDuration(value)
		} else if targetType.Name() == "Time" {
			return parseTime(value)
		}
	}

	switch targetType.Kind() {
	case reflect.String:
		return value, nil
	case reflect.Bool:
		return parseBool(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return parseInt(value, targetType)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return parseUint(value, targetType)
	case reflect.Float32, reflect.Float64:
		return parseFloat(value, targetType)
	case reflect.Slice:
		return p.parseSlice(value, targetType)
	default:
		return nil, fmt.Errorf("unsupported type: %v", targetType)
	}
}

func parseBool(value string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "1", "yes", "on":
		return true, nil
	case "false", "0", "no", "off":
		return false, nil
	default:
		return false, fmt.Errorf("invalid bool: %q", value)
	}
}

func parseInt(value string, targetType reflect.Type) (any, error) {
	i, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid %s: %q", targetType.Name(), value)
	}
	switch targetType.Kind() {
	case reflect.Int:
		if strconv.IntSize == 32 && (i < math.MinInt32 || i > math.MaxInt32) {
			return nil, fmt.Errorf("int out of range: %d", i)
		}
		return int(i), nil
	case reflect.Int8:
		if i < -128 || i > 127 {
			return nil, fmt.Errorf("int8 out of range: %d", i)
		}
		return int8(i), nil
	case reflect.Int16:
		if i < -32768 || i > 32767 {
			return nil, fmt.Errorf("int16 out of range: %d", i)
		}
		return int16(i), nil
	case reflect.Int32:
		if i < -2147483648 || i > 2147483647 {
			return nil, fmt.Errorf("int32 out of range: %d", i)
		}
		return int32(i), nil
	case reflect.Int64:
		return i, nil
	default:
		return nil, fmt.Errorf("unexpected int type: %v", targetType)
	}
}

func parseUint(value string, targetType reflect.Type) (any, error) {
	u, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid %s: %q", targetType.Name(), value)
	}
	switch targetType.Kind() {
	case reflect.Uint:
		if strconv.IntSize == 32 && u > math.MaxUint32 {
			return nil, fmt.Errorf("uint out of range: %d", u)
		}
		return uint(u), nil
	case reflect.Uint8:
		if u > 255 {
			return nil, fmt.Errorf("uint8 out of range: %d", u)
		}
		return uint8(u), nil
	case reflect.Uint16:
		if u > 65535 {
			return nil, fmt.Errorf("uint16 out of range: %d", u)
		}
		return uint16(u), nil
	case reflect.Uint32:
		if u > 4294967295 {
			return nil, fmt.Errorf("uint32 out of range: %d", u)
		}
		return uint32(u), nil
	case reflect.Uint64:
		return u, nil
	default:
		return nil, fmt.Errorf("unexpected uint type: %v", targetType)
	}
}

func parseFloat(value string, targetType reflect.Type) (any, error) {
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid %s: %q", targetType.Name(), value)
	}
	if targetType.Kind() == reflect.Float32 {
		return float32(f), nil
	}
	return f, nil
}

func parseDuration(value string) (any, error) {
	d, err := time.ParseDuration(value)
	if err != nil {
		return nil, fmt.Errorf("invalid duration: %q (use: 5s, 10m, 1h, etc.)", value)
	}
	return d, nil
}

func parseTime(value string) (any, error) {
	// RFC3339 only in v0.1
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, fmt.Errorf("invalid time: %q (use RFC3339 format: 2006-01-02T15:04:05Z07:00)", value)
	}
	return t, nil
}

func (p *Parser) parseSlice(value string, targetType reflect.Type) (any, error) {
	elemType := targetType.Elem()
	parts := strings.Split(strings.TrimSpace(value), ",")
	result := reflect.MakeSlice(targetType, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		parsed, err := p.Parse(part, elemType)
		if err != nil {
			return nil, fmt.Errorf("slice element parse error: %w", err)
		}
		result = reflect.Append(result, reflect.ValueOf(parsed))
	}

	return result.Interface(), nil
}

func (p *Parser) zeroValue(typ reflect.Type) any {
	return reflect.Zero(typ).Interface()
}
