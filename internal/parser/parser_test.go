package parser

import (
	"reflect"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	p := New()

	t.Run("string", func(t *testing.T) {
		result, err := p.Parse("hello", reflect.TypeOf(""))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "hello" {
			t.Errorf("expected 'hello', got %q", result)
		}
	})

	t.Run("empty_returns_zero", func(t *testing.T) {
		types := []reflect.Type{
			reflect.TypeOf(""),
			reflect.TypeOf(0),
			reflect.TypeOf(false),
			reflect.TypeOf(float64(0)),
			reflect.TypeOf([]string{}),
		}
		for _, typ := range types {
			_, err := p.Parse("", typ)
			if err != nil {
				t.Errorf("Parse('', %v): unexpected error: %v", typ, err)
			}
		}
	})

	t.Run("bool", func(t *testing.T) {
		cases := []struct {
			input   string
			want    bool
			wantErr bool
		}{
			{"true", true, false},
			{"false", false, false},
			{"1", true, false},
			{"0", false, false},
			{"yes", true, false},
			{"no", false, false},
			{"invalid", false, true},
		}
		for _, tc := range cases {
			result, err := p.Parse(tc.input, reflect.TypeOf(false))
			if (err != nil) != tc.wantErr {
				t.Errorf("bool %q: got error %v, wantErr %v", tc.input, err, tc.wantErr)
				continue
			}
			if err == nil && result != tc.want {
				t.Errorf("bool %q: expected %v, got %v", tc.input, tc.want, result)
			}
		}
	})

	t.Run("int", func(t *testing.T) {
		cases := []struct {
			input   string
			typ     reflect.Type
			want    any
			wantErr bool
		}{
			{"42", reflect.TypeOf(int(0)), int(42), false},
			{"-128", reflect.TypeOf(int8(0)), int8(-128), false},
			{"200", reflect.TypeOf(int16(0)), int16(200), false},
			{"100", reflect.TypeOf(int32(0)), int32(100), false},
			{"400", reflect.TypeOf(int64(0)), int64(400), false},
			{"invalid", reflect.TypeOf(int(0)), nil, true},
			{"32768", reflect.TypeOf(int16(0)), nil, true},
		}
		for _, tc := range cases {
			result, err := p.Parse(tc.input, tc.typ)
			if (err != nil) != tc.wantErr {
				t.Errorf("int %q: got error %v, wantErr %v", tc.input, err, tc.wantErr)
				continue
			}
			if err == nil && result != tc.want {
				t.Errorf("int %q: expected %v, got %v", tc.input, tc.want, result)
			}
		}
	})

	t.Run("uint", func(t *testing.T) {
		cases := []struct {
			input   string
			typ     reflect.Type
			want    any
			wantErr bool
		}{
			{"255", reflect.TypeOf(uint8(0)), uint8(255), false},
			{"256", reflect.TypeOf(uint8(0)), nil, true},
			{"100", reflect.TypeOf(uint(0)), uint(100), false},
			{"200", reflect.TypeOf(uint16(0)), uint16(200), false},
			{"300", reflect.TypeOf(uint32(0)), uint32(300), false},
			{"400", reflect.TypeOf(uint64(0)), uint64(400), false},
			{"invalid", reflect.TypeOf(uint(0)), nil, true},
		}
		for _, tc := range cases {
			result, err := p.Parse(tc.input, tc.typ)
			if (err != nil) != tc.wantErr {
				t.Errorf("uint %q: got error %v, wantErr %v", tc.input, err, tc.wantErr)
				continue
			}
			if err == nil && result != tc.want {
				t.Errorf("uint %q: expected %v, got %v", tc.input, tc.want, result)
			}
		}
	})

	t.Run("float", func(t *testing.T) {
		result, err := p.Parse("3.14", reflect.TypeOf(float64(0)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		f := result.(float64)
		if f < 3.13 || f > 3.15 {
			t.Errorf("expected ~3.14, got %v", f)
		}
	})

	t.Run("duration", func(t *testing.T) {
		result, err := p.Parse("5s", reflect.TypeOf(time.Duration(0)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.(time.Duration) != 5*time.Second {
			t.Errorf("expected 5s, got %v", result)
		}
	})

	t.Run("time", func(t *testing.T) {
		result, err := p.Parse("2026-01-01T00:00:00Z", reflect.TypeOf(time.Time{}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		tm := result.(time.Time)
		if tm.Year() != 2026 || tm.Month() != time.January || tm.Day() != 1 {
			t.Errorf("unexpected date: %v", tm)
		}
	})

	t.Run("slice/string", func(t *testing.T) {
		result, err := p.Parse("a,b,c", reflect.TypeOf([]string{}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		s := result.([]string)
		if len(s) != 3 || s[0] != "a" || s[1] != "b" || s[2] != "c" {
			t.Errorf("expected [a b c], got %v", s)
		}
	})

	t.Run("slice/int", func(t *testing.T) {
		result, err := p.Parse("1,2,3", reflect.TypeOf([]int{}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		s := result.([]int)
		if len(s) != 3 || s[0] != 1 || s[1] != 2 || s[2] != 3 {
			t.Errorf("expected [1 2 3], got %v", s)
		}
	})

	t.Run("unsupported_type_error", func(t *testing.T) {
		_, err := p.Parse("test", reflect.TypeOf(struct{}{}))
		if err == nil {
			t.Error("expected error for unsupported type, got nil")
		}
	})
}

func TestParseMap(t *testing.T) {
	p := New()
	strStr := reflect.TypeOf(map[string]string{})
	strInt := reflect.TypeOf(map[string]int{})

	t.Run("string_string", func(t *testing.T) {
		val, err := p.Parse("KEY1=v1,KEY2=v2", strStr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		m := val.(map[string]string)
		if m["KEY1"] != "v1" || m["KEY2"] != "v2" {
			t.Errorf("unexpected map: %v", m)
		}
	})

	t.Run("int_values", func(t *testing.T) {
		val, err := p.Parse("a=1,b=2", strInt)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		m := val.(map[string]int)
		if m["a"] != 1 || m["b"] != 2 {
			t.Errorf("unexpected map: %v", m)
		}
	})

	t.Run("empty", func(t *testing.T) {
		val, err := p.Parse("", strStr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(val.(map[string]string)) != 0 {
			t.Error("expected empty map")
		}
	})

	t.Run("invalid_entry", func(t *testing.T) {
		_, err := p.Parse("noequals", strStr)
		if err == nil {
			t.Fatal("expected error for missing '='")
		}
	})

	t.Run("non_string_key_error", func(t *testing.T) {
		_, err := p.Parse("1=v", reflect.TypeOf(map[int]string{}))
		if err == nil {
			t.Fatal("expected error for non-string key")
		}
	})

	t.Run("quoted_values", func(t *testing.T) {
		cases := []struct {
			input string
			want  map[string]string
		}{
			{`KEY1="a,b",KEY2=c`, map[string]string{"KEY1": "a,b", "KEY2": "c"}},
			{`K1=simple,K2=also`, map[string]string{"K1": "simple", "K2": "also"}},
			{`K="x,y,z"`, map[string]string{"K": "x,y,z"}},
			{`A="1,2",B="3,4",C=5`, map[string]string{"A": "1,2", "B": "3,4", "C": "5"}},
		}
		for _, tc := range cases {
			result, err := p.Parse(tc.input, strStr)
			if err != nil {
				t.Errorf("input=%q: unexpected error: %v", tc.input, err)
				continue
			}
			got := result.(map[string]string)
			for k, wantVal := range tc.want {
				if got[k] != wantVal {
					t.Errorf("input=%q key=%q: got %q want %q", tc.input, k, got[k], wantVal)
				}
			}
		}
	})
}
