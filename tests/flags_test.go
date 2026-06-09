package confkit_test

import (
	"context"
	"testing"

	confkit "github.com/MimoJanra/confkit"
)

func TestFlagsSourceLongEq(t *testing.T) {
	src := confkit.FromFlagsWithArgs([]string{"--port=9000", "--host=localhost"})
	field := &confkit.FieldInfo{Name: "Port", Tags: map[string]string{"flag": "port"}}
	val, ok, err := src.Lookup(context.Background(), field)
	if err != nil || !ok || val != "9000" {
		t.Errorf("expected 9000, got %v ok=%v err=%v", val, ok, err)
	}
}

func TestFlagsSourceLongSpace(t *testing.T) {
	src := confkit.FromFlagsWithArgs([]string{"--host", "db.local", "--port", "5432"})
	field := &confkit.FieldInfo{Name: "Host", Tags: map[string]string{"flag": "host"}}
	val, ok, err := src.Lookup(context.Background(), field)
	if err != nil || !ok || val != "db.local" {
		t.Errorf("expected db.local, got %v ok=%v err=%v", val, ok, err)
	}
}

func TestFlagsSourceShortEq(t *testing.T) {
	src := confkit.FromFlagsWithArgs([]string{"-p=8080"})
	field := &confkit.FieldInfo{Name: "Port", Tags: map[string]string{"flag": "port", "short": "p"}}
	val, ok, err := src.Lookup(context.Background(), field)
	if err != nil || !ok || val != "8080" {
		t.Errorf("expected 8080, got %v ok=%v err=%v", val, ok, err)
	}
}

func TestFlagsSourceShortSpace(t *testing.T) {
	src := confkit.FromFlagsWithArgs([]string{"-p", "3000"})
	field := &confkit.FieldInfo{Name: "Port", Tags: map[string]string{"flag": "port", "short": "p"}}
	val, ok, err := src.Lookup(context.Background(), field)
	if err != nil || !ok || val != "3000" {
		t.Errorf("expected 3000, got %v ok=%v err=%v", val, ok, err)
	}
}

func TestFlagsSourceBoolFlag(t *testing.T) {
	src := confkit.FromFlagsWithArgs([]string{"--verbose"})
	field := &confkit.FieldInfo{Name: "Verbose", Tags: map[string]string{}}
	val, ok, err := src.Lookup(context.Background(), field)
	if err != nil || !ok || val != "true" {
		t.Errorf("expected true, got %v ok=%v err=%v", val, ok, err)
	}
}

func TestFlagsSourceMissing(t *testing.T) {
	src := confkit.FromFlagsWithArgs([]string{"--other=value"})
	field := &confkit.FieldInfo{Name: "Port", Tags: map[string]string{"flag": "port"}}
	_, ok, err := src.Lookup(context.Background(), field)
	if err != nil || ok {
		t.Errorf("expected not found, got ok=%v err=%v", ok, err)
	}
}

func TestFlagsSourceKebabCase(t *testing.T) {
	src := confkit.FromFlagsWithArgs([]string{"--database-url=postgres://localhost/db"})
	field := &confkit.FieldInfo{Name: "DatabaseURL", Tags: map[string]string{}}
	val, ok, err := src.Lookup(context.Background(), field)
	if err != nil || !ok || val != "postgres://localhost/db" {
		t.Errorf("expected postgres://localhost/db, got %v ok=%v err=%v", val, ok, err)
	}
}

func TestFlagsSourceSnakeCase(t *testing.T) {
	src := confkit.FromFlagsWithArgs([]string{"--database_url=postgres://localhost/db"})
	field := &confkit.FieldInfo{Name: "DatabaseURL", Tags: map[string]string{}}
	val, ok, err := src.Lookup(context.Background(), field)
	if err != nil || !ok || val != "postgres://localhost/db" {
		t.Errorf("expected postgres://localhost/db, got %v ok=%v err=%v", val, ok, err)
	}
}

func TestFlagsSourceNegativeNumber(t *testing.T) {
	src := confkit.FromFlagsWithArgs([]string{"--offset", "-1", "--delta", "-3.14"})
	field := &confkit.FieldInfo{Name: "Offset", Tags: map[string]string{"flag": "offset"}}
	val, ok, err := src.Lookup(context.Background(), field)
	if err != nil || !ok || val != "-1" {
		t.Errorf("expected -1, got %v ok=%v err=%v", val, ok, err)
	}
}

func TestFlagsSourceRepeatedFlag(t *testing.T) {
	src := confkit.FromFlagsWithArgs([]string{"--tag", "a", "--tag", "b", "--tag", "c"})
	field := &confkit.FieldInfo{Name: "Tag", Tags: map[string]string{"flag": "tag"}}
	val, ok, err := src.Lookup(context.Background(), field)
	if err != nil || !ok || val != "a,b,c" {
		t.Errorf("expected 'a,b,c', got %v ok=%v err=%v", val, ok, err)
	}
}

func TestFlagsSourceRepeatedSliceIntegration(t *testing.T) {
	type Cfg struct {
		Tags []string `flag:"tag"`
	}
	src := confkit.FromFlagsWithArgs([]string{"--tag", "foo", "--tag", "bar"})
	cfg, err := confkit.Load[Cfg](src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Tags) != 2 || cfg.Tags[0] != "foo" || cfg.Tags[1] != "bar" {
		t.Errorf("expected [foo bar], got %v", cfg.Tags)
	}
}

func TestFlagsIntegration(t *testing.T) {
	type Cfg struct {
		Host string `flag:"host"`
		Port int    `flag:"port"`
	}
	src := confkit.FromFlagsWithArgs([]string{"--host=myhost", "--port=1234"})
	cfg, err := confkit.Load[Cfg](src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Host != "myhost" {
		t.Errorf("expected myhost, got %s", cfg.Host)
	}
	if cfg.Port != 1234 {
		t.Errorf("expected 1234, got %d", cfg.Port)
	}
}

func TestFromFlags(t *testing.T) {
	src := confkit.FromFlags()
	if src == nil {
		t.Error("FromFlags returned nil")
	}
}
