package structtags

import (
	"reflect"
	"testing"
	"time"
)

func TestSnakeCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Port", "port"},
		{"HostName", "host_name"},
		{"DatabaseURL", "database_u_r_l"},
		{"already_lower", "already_lower"},
		{"A", "a"},
		{"", ""},
	}
	for _, tt := range tests {
		got := SnakeCase(tt.input)
		if got != tt.want {
			t.Errorf("SnakeCase(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestIsSpecialType(t *testing.T) {
	if !IsSpecialType(reflect.TypeOf(time.Time{})) {
		t.Error("Expected time.Time to be special")
	}
	if !IsSpecialType(reflect.TypeOf(time.Duration(0))) {
		t.Error("Expected time.Duration to be special")
	}
	if IsSpecialType(reflect.TypeOf("")) {
		t.Error("Expected string to not be special")
	}
	if IsSpecialType(reflect.TypeOf(0)) {
		t.Error("Expected int to not be special")
	}
}

func TestParseStructTags(t *testing.T) {
	type S struct {
		F string `env:"MY_VAR" yaml:"my_var,omitempty" default:"val" secret:"true"`
	}
	field, _ := reflect.TypeOf(S{}).FieldByName("F")
	tags := ParseStructTags(field.Tag)

	if tags["env"] != "MY_VAR" {
		t.Errorf("env tag: got %q", tags["env"])
	}
	if tags["yaml"] != "my_var" {
		t.Errorf("yaml tag after comma strip: got %q", tags["yaml"])
	}
	if tags["default"] != "val" {
		t.Errorf("default tag: got %q", tags["default"])
	}
	if tags["secret"] != "true" {
		t.Errorf("secret tag: got %q", tags["secret"])
	}
	if _, ok := tags["flag"]; ok {
		t.Error("flag tag should be absent")
	}
}

func TestParseStructTagsJSON(t *testing.T) {
	type S struct {
		F string `json:"field_name,omitempty"`
	}
	field, _ := reflect.TypeOf(S{}).FieldByName("F")
	tags := ParseStructTags(field.Tag)

	if tags["json"] != "field_name" {
		t.Errorf("json tag after comma strip: got %q", tags["json"])
	}
}
