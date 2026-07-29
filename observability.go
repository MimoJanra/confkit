package confkit

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// LoadMetrics holds timing and error counts for a single load, for callers that
// record their own metrics.
type LoadMetrics struct {
	TotalTime      time.Duration
	SourceTimes    map[string]time.Duration
	ValidationTime time.Duration
	ErrorCount     int
}

// DumpConfig serializes cfg as flat JSON keyed by Go field path ("DB.Host"),
// redacting fields tagged `secret:"true"`. It requires the FieldInfo slice from
// ScanFields.
//
// Prefer Dump, DumpString or DumpYAML, which need no FieldInfo and emit nested
// keys named after the struct tags.
func DumpConfig(cfg any, fields []FieldInfo) ([]byte, error) {
	dump := make(map[string]any)
	val := reflect.ValueOf(cfg)

	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}

	for _, field := range fields {
		value := getFieldValue(val, field.Path)
		if field.IsSecret {
			dump[field.Path] = "***REDACTED***"
		} else {
			dump[field.Path] = value
		}
	}

	return json.MarshalIndent(dump, "", "  ")
}

func getFieldValue(val reflect.Value, path string) any {
	parts := splitPath(path)
	current := val

	for _, part := range parts {
		field := current.FieldByName(part)
		if !field.IsValid() {
			return nil
		}

		if field.Kind() == reflect.Pointer {
			if field.IsNil() {
				return nil
			}
			current = field.Elem()
		} else {
			current = field
		}
	}

	if current.IsValid() {
		return current.Interface()
	}
	return nil
}

// LogLoadStart returns a one-line JSON event naming the sources about to be read,
// ready to write to a structured log.
func LogLoadStart(sources []string) string {
	payload := map[string]any{
		"event":     "config_load_start",
		"sources":   sources,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	return toJSON(payload)
}

// LogLoadComplete returns a one-line JSON event summarising a finished load. It
// records counts and timing only, never field values.
func LogLoadComplete(duration time.Duration, fieldCount int, errorCount int) string {
	return fmt.Sprintf(`{"event":"config_load_complete","duration_ms":%d,"fields_loaded":%d,"validation_errors":%d}`,
		duration.Milliseconds(), fieldCount, errorCount)
}

func toJSON(data any) string {
	b, _ := json.Marshal(data)
	return string(b)
}

func splitPath(path string) []string {
	var parts []string
	var cur strings.Builder
	for _, ch := range path {
		if ch == '.' {
			if cur.Len() > 0 {
				parts = append(parts, cur.String())
				cur.Reset()
			}
		} else {
			cur.WriteRune(ch)
		}
	}
	if cur.Len() > 0 {
		parts = append(parts, cur.String())
	}
	return parts
}
