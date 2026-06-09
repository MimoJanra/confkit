package confkit

import (
	"reflect"
	"testing"
)

func TestScanFields(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		type cfg struct {
			Port int    `env:"PORT" default:"8080"`
			Mode string `env:"MODE" default:"dev"`
		}
		fields := ScanFields(cfg{})
		if len(fields) != 2 {
			t.Fatalf("expected 2 fields, got %d", len(fields))
		}
		if fields[0].Name != "Port" || fields[0].Tags["env"] != "PORT" || fields[0].Tags["default"] != "8080" {
			t.Errorf("unexpected Port field: %+v", fields[0])
		}
		if fields[1].Name != "Mode" {
			t.Errorf("expected Mode field, got %s", fields[1].Name)
		}
	})

	t.Run("pointer_struct", func(t *testing.T) {
		type cfg struct {
			Host string `env:"PSF_HOST"`
		}
		fields := ScanFields(&cfg{})
		if len(fields) == 0 {
			t.Error("expected fields from pointer struct")
		}
	})

	t.Run("embedded", func(t *testing.T) {
		type Base struct {
			Host string `env:"EMB_HOST"`
		}
		type cfg struct {
			Base
			Port int `env:"EMB_PORT"`
		}
		fields := ScanFields(cfg{})
		names := make(map[string]bool)
		for _, f := range fields {
			names[f.Name] = true
		}
		if !names["Host"] {
			t.Error("expected embedded Host field to be promoted")
		}
		if !names["Port"] {
			t.Error("expected Port field")
		}
	})
}

func TestFieldValueToString(t *testing.T) {
	cases := []struct {
		val  reflect.Value
		want string
	}{
		{reflect.ValueOf(true), "true"},
		{reflect.ValueOf(uint(5)), "5"},
		{reflect.ValueOf(float64(3.14)), "3.14"},
	}
	for _, tc := range cases {
		got := fieldValueToString(tc.val, false)
		if got != tc.want {
			t.Errorf("fieldValueToString(%v) = %q, want %q", tc.val, got, tc.want)
		}
	}
}

func TestIsZeroValue(t *testing.T) {
	t.Run("bool", func(t *testing.T) {
		if !isZeroValue(reflect.ValueOf(false)) {
			t.Error("false should be zero")
		}
		if isZeroValue(reflect.ValueOf(true)) {
			t.Error("true should not be zero")
		}
	})

	t.Run("slice", func(t *testing.T) {
		if !isZeroValue(reflect.ValueOf([]string{})) {
			t.Error("empty slice should be zero")
		}
		if isZeroValue(reflect.ValueOf([]string{"a"})) {
			t.Error("non-empty slice should not be zero")
		}
	})
}

func TestGetFieldByPath(t *testing.T) {
	t.Run("nil_ptr", func(t *testing.T) {
		type Inner struct{ Val string }
		type Outer struct{ Inner *Inner }
		result := getFieldByPath(reflect.ValueOf(Outer{}), "Inner.Val")
		if result.IsValid() {
			t.Error("expected invalid value for nil pointer path")
		}
	})

	t.Run("missing_field", func(t *testing.T) {
		result := getFieldByPath(reflect.ValueOf(struct{ X string }{}), "NoSuchField")
		if result.IsValid() {
			t.Error("expected invalid value for missing field")
		}
	})
}

func TestSetFieldValue(t *testing.T) {
	t.Run("nested_ptr_initialized", func(t *testing.T) {
		type Inner struct{ Val string }
		type Outer struct{ Inner *Inner }
		v := Outer{}
		setFieldValue(reflect.ValueOf(&v).Elem(), "Inner.Val", "hello")
		if v.Inner == nil || v.Inner.Val != "hello" {
			t.Errorf("expected Inner.Val='hello', got %+v", v)
		}
	})
}
