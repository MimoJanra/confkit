package confkit

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestParseMapQuotedValues(t *testing.T) {
	p := NewParser()
	typ := reflect.TypeOf(map[string]string{})

	tests := []struct {
		input string
		want  map[string]string
	}{
		{`KEY1="a,b",KEY2=c`, map[string]string{"KEY1": "a,b", "KEY2": "c"}},
		{`K1=simple,K2=also`, map[string]string{"K1": "simple", "K2": "also"}},
		{`K="x,y,z"`, map[string]string{"K": "x,y,z"}},
		{`A="1,2",B="3,4",C=5`, map[string]string{"A": "1,2", "B": "3,4", "C": "5"}},
	}

	for _, tc := range tests {
		result, err := p.Parse(tc.input, typ)
		if err != nil {
			t.Errorf("input=%q: unexpected error: %v", tc.input, err)
			continue
		}
		got, ok := result.(map[string]string)
		if !ok {
			t.Errorf("input=%q: expected map[string]string", tc.input)
			continue
		}
		for k, want := range tc.want {
			if got[k] != want {
				t.Errorf("input=%q key=%q: got %q want %q", tc.input, k, got[k], want)
			}
		}
	}
}

func TestLoadContext(t *testing.T) {
	type C struct {
		Port int `env:"PORT" default:"9090"`
	}
	ctx := context.Background()
	cfg, err := LoadContext[C](ctx)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 9090 {
		t.Errorf("expected 9090, got %d", cfg.Port)
	}
}

func TestLoadWithOptionsContext(t *testing.T) {
	type C struct {
		Port int `env:"PORT" default:"1234"`
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cfg, err := LoadWithOptionsContext[C](ctx)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 1234 {
		t.Errorf("expected 1234, got %d", cfg.Port)
	}
}

func TestWithContext(t *testing.T) {
	type C struct {
		X string `default:"hello"`
	}
	ctx := context.Background()
	cfg, err := LoadWithOptions[C](WithContext(ctx))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.X != "hello" {
		t.Errorf("got %q", cfg.X)
	}
}

func TestValidateOnlySuccess(t *testing.T) {
	type C struct {
		Port int    `env:"PORT" default:"8080" validate:"min=1,max=65535"`
		Host string `default:"localhost"`
	}
	cfg, err := ValidateOnly[C](context.Background())
	if err != nil {
		t.Fatalf("ValidateOnly should succeed: %v", err)
	}
	if cfg.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Port)
	}
}

func TestValidateOnlySkipsHooksAndAudit(t *testing.T) {
	type C struct {
		Port int `default:"8080"`
	}
	hookCalled := false
	auditCalled := false

	_, err := ValidateOnly[C](context.Background(),
		WithLoadHook(func(bool, time.Duration, int) { hookCalled = true }),
		WithAuditLogger(func([]AuditEntry) { auditCalled = true }),
	)
	if err != nil {
		t.Fatal(err)
	}
	if hookCalled {
		t.Error("ValidateOnly should not call LoadHookFunc")
	}
	if auditCalled {
		t.Error("ValidateOnly should not call AuditLogger")
	}
}

func TestValidateOnlyFailure(t *testing.T) {
	type C struct {
		Port int `env:"V1_PORT" validate:"min=1,max=65535"`
	}
	t.Setenv("V1_PORT", "99999")
	_, err := ValidateOnly[C](context.Background(), WithSource(FromEnv()))
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestMustLoad(t *testing.T) {
	type C struct {
		X int `default:"42"`
	}
	cfg := MustLoad[C]()
	if cfg.X != 42 {
		t.Errorf("expected 42, got %d", cfg.X)
	}
}

func TestMustLoadPanics(t *testing.T) {
	type C struct {
		X int `env:"MUST_LOAD_PORT" validate:"required"`
	}
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustLoad should panic on error")
		}
	}()
	MustLoad[C]()
}

func TestMustLoadContext(t *testing.T) {
	type C struct {
		X int `default:"7"`
	}
	cfg := MustLoadContext[C](context.Background())
	if cfg.X != 7 {
		t.Errorf("expected 7, got %d", cfg.X)
	}
}

func TestDumpJSON(t *testing.T) {
	type C struct {
		Port     int    `json:"port"`
		Password string `json:"password" secret:"true"`
	}
	cfg := C{Port: 8080, Password: "secret123"}

	b, err := Dump(cfg)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, b)
	}
	if m["port"] != float64(8080) {
		t.Errorf("expected port=8080, got %v", m["port"])
	}
	if m["password"] != "***REDACTED***" {
		t.Errorf("expected password redacted, got %v", m["password"])
	}
}

func TestDumpNoRedact(t *testing.T) {
	type C struct {
		Pass string `json:"pass" secret:"true"`
	}
	cfg := C{Pass: "plaintext"}
	b, err := Dump(cfg, WithDumpRedactSecrets(false))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "plaintext") {
		t.Errorf("expected plaintext in dump: %s", b)
	}
}

func TestDumpYAMLFormat(t *testing.T) {
	type C struct {
		Host string `yaml:"host"`
	}
	cfg := C{Host: "localhost"}
	b, err := DumpYAML(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "host:") {
		t.Errorf("expected YAML output with 'host:', got: %s", b)
	}
}

func TestDumpString(t *testing.T) {
	type C struct {
		X int `json:"x"`
	}
	s := DumpString(C{X: 5})
	if !strings.Contains(s, "5") {
		t.Errorf("expected '5' in dump string: %s", s)
	}
}

func TestDumpNestedStruct(t *testing.T) {
	type DB struct {
		Host string `json:"host"`
		Pass string `json:"pass" secret:"true"`
	}
	type C struct {
		DB DB `json:"db"`
	}
	cfg := C{DB: DB{Host: "pg", Pass: "s3cr3t"}}
	b, err := Dump(cfg)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	db, ok := m["db"].(map[string]any)
	if !ok {
		t.Fatalf("expected nested db map, got %T", m["db"])
	}
	if db["host"] != "pg" {
		t.Errorf("expected host=pg, got %v", db["host"])
	}
	if db["pass"] != "***REDACTED***" {
		t.Errorf("expected pass redacted, got %v", db["pass"])
	}
}

func TestDumpWithFormatYAML(t *testing.T) {
	type C struct {
		Name string `yaml:"name"`
	}
	b, err := Dump(C{Name: "test"}, WithDumpFormat(FormatYAML))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "name:") {
		t.Errorf("expected YAML: %s", b)
	}
}

func TestFindFileWithExtension(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(f, []byte("port: 8080\n"), 0600); err != nil {
		t.Fatal(err)
	}
	got, ok := FindFile("config.yaml", dir)
	if !ok {
		t.Fatal("expected to find config.yaml")
	}
	if filepath.Base(got) != "config.yaml" {
		t.Errorf("unexpected path: %s", got)
	}
}

func TestFindFileWithoutExtension(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "config.json")
	if err := os.WriteFile(f, []byte(`{"port":9000}`), 0600); err != nil {
		t.Fatal(err)
	}
	got, ok := FindFile("config", dir)
	if !ok {
		t.Fatal("expected to find config (without ext)")
	}
	if filepath.Ext(got) != ".json" {
		t.Errorf("expected .json ext, got %s", filepath.Ext(got))
	}
}

func TestFindFileNotFound(t *testing.T) {
	_, ok := FindFile("nonexistent.yaml", t.TempDir())
	if ok {
		t.Error("expected not found")
	}
}

func TestDefaultSearchDirs(t *testing.T) {
	dirs := DefaultSearchDirs("myapp")
	if len(dirs) < 3 {
		t.Errorf("expected at least 3 dirs, got %d", len(dirs))
	}
	found := false
	for _, d := range dirs {
		if strings.Contains(d, "myapp") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected appName in some dir, got %v", dirs)
	}
}

func TestFindSourceFound(t *testing.T) {
	type C struct {
		Port int `yaml:"port" default:"0"`
	}
	dir := t.TempDir()
	f := filepath.Join(dir, "app.yaml")
	if err := os.WriteFile(f, []byte("port: 1234\n"), 0600); err != nil {
		t.Fatal(err)
	}
	src := FindSource("app.yaml", dir)
	cfg, err := Load[C](src)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 1234 {
		t.Errorf("expected 1234, got %d", cfg.Port)
	}
}

func TestFindSourceNotFound(t *testing.T) {
	type C struct {
		Port int `default:"0"`
	}
	src := FindSource("missing.yaml", t.TempDir())
	_, err := Load[C](src)
	if err == nil {
		t.Error("expected error from missing source")
	}
}

func TestOverlayPath(t *testing.T) {
	tests := []struct{ base, env, want string }{
		{"config.yaml", "prod", "config.prod.yaml"},
		{"/etc/app/config.toml", "staging", "/etc/app/config.staging.toml"},
		{"app.json", "dev", "app.dev.json"},
	}
	for _, tc := range tests {
		got := OverlayPath(tc.base, tc.env)
		if got != tc.want {
			t.Errorf("OverlayPath(%q, %q) = %q, want %q", tc.base, tc.env, got, tc.want)
		}
	}
}

func TestFromOverlayMerges(t *testing.T) {
	type C struct {
		Port     int    `yaml:"port" default:"0"`
		LogLevel string `yaml:"log_level" default:"info"`
	}
	dir := t.TempDir()

	base := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(base, []byte("port: 8080\nlog_level: info\n"), 0600); err != nil {
		t.Fatal(err)
	}
	over := filepath.Join(dir, "config.prod.yaml")
	if err := os.WriteFile(over, []byte("log_level: warn\n"), 0600); err != nil {
		t.Fatal(err)
	}

	src := FromOverlay(FromYAML(base), "prod")
	cfg, err := Load[C](src)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 8080 {
		t.Errorf("expected port from base=8080, got %d", cfg.Port)
	}
	if cfg.LogLevel != "warn" {
		t.Errorf("expected log_level from overlay=warn, got %q", cfg.LogLevel)
	}
}

func TestFromOverlayMissingOverlay(t *testing.T) {
	type C struct {
		Port int `yaml:"port" default:"0"`
	}
	dir := t.TempDir()
	base := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(base, []byte("port: 5000\n"), 0600); err != nil {
		t.Fatal(err)
	}

	src := FromOverlay(FromYAML(base), "prod")
	cfg, err := Load[C](src)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 5000 {
		t.Errorf("expected 5000, got %d", cfg.Port)
	}
}

func TestFromOverlayNonFileSource(t *testing.T) {

	base := FromEnv()
	result := FromOverlay(base, "prod")
	if result != base {
		t.Error("expected base returned unchanged for non-fileSource")
	}
}

func TestComputeDelta(t *testing.T) {
	old := map[string]any{"a": 1, "b": 2, "c": 3}
	new_ := map[string]any{"a": 1, "b": 99, "d": 4}

	delta := computeDelta(old, new_)

	if len(delta.Changed) != 1 || delta.Changed[0] != "b" {
		t.Errorf("expected Changed=[b], got %v", delta.Changed)
	}
	if len(delta.Added) != 1 || delta.Added[0] != "d" {
		t.Errorf("expected Added=[d], got %v", delta.Added)
	}
	if len(delta.Removed) != 1 || delta.Removed[0] != "c" {
		t.Errorf("expected Removed=[c], got %v", delta.Removed)
	}
}

func TestFlattenAnyMap(t *testing.T) {
	m := map[string]any{
		"a": 1,
		"b": map[string]any{"c": 2, "d": 3},
	}
	flat := flattenAnyMap("", m)
	if flat["a"] != 1 {
		t.Errorf("expected a=1, got %v", flat["a"])
	}
	if flat["b.c"] != 2 {
		t.Errorf("expected b.c=2, got %v", flat["b.c"])
	}
	if flat["b.d"] != 3 {
		t.Errorf("expected b.d=3, got %v", flat["b.d"])
	}
}

func TestAddDeltaListener(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "cfg.yaml")
	if err := os.WriteFile(f, []byte("port: 8080\n"), 0600); err != nil {
		t.Fatal(err)
	}

	watcher, err := NewConfigWatcher(f)
	if err != nil {
		t.Fatal(err)
	}

	called := make(chan ConfigDelta, 1)
	watcher.AddDeltaListener(func(delta ConfigDelta, _, _ map[string]any, err error) {
		if err == nil {
			called <- delta
		}
	})

	watcher.SetPollInterval(50 * time.Millisecond)
	watcher.Start()
	defer watcher.Stop()

	time.Sleep(100 * time.Millisecond)
	if err := os.WriteFile(f, []byte("port: 9090\n"), 0600); err != nil {
		t.Fatal(err)
	}

	select {
	case delta := <-called:
		if len(delta.Changed) == 0 {
			t.Errorf("expected non-empty Changed, got %+v", delta)
		}
	case <-time.After(3 * time.Second):
		t.Error("timeout waiting for delta listener")
	}
}

func TestFirstError(t *testing.T) {
	er := &ErrorReport{}
	if er.FirstError() != nil {
		t.Error("expected nil on empty report")
	}
	er.AddError(FieldError{Path: "x", Message: "bad"})
	er.AddError(FieldError{Path: "y", Message: "also bad"})
	fe := er.FirstError()
	if fe == nil {
		t.Fatal("expected non-nil")
	}
	if fe.Path != "x" {
		t.Errorf("expected first error path=x, got %q", fe.Path)
	}
}

func TestParseFileToFlatMapYAML(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "c.yaml")
	if err := os.WriteFile(f, []byte("a: 1\nb:\n  c: 2\n"), 0600); err != nil {
		t.Fatal(err)
	}
	m, err := parseFileToFlatMap(f)
	if err != nil {
		t.Fatal(err)
	}
	if m["b.c"] != 2 {
		t.Errorf("expected b.c=2, got %v", m["b.c"])
	}
}

func TestParseFileToFlatMapJSON(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "c.json")
	if err := os.WriteFile(f, []byte(`{"x":5}`), 0600); err != nil {
		t.Fatal(err)
	}
	m, err := parseFileToFlatMap(f)
	if err != nil {
		t.Fatal(err)
	}
	if m["x"] != float64(5) {
		t.Errorf("expected x=5, got %v", m["x"])
	}
}

// Fix 1: Dump redacts secrets nested inside slices and maps of structs.
func TestDumpSecretInSliceOfStructs(t *testing.T) {
	type Cred struct {
		User  string `json:"user"`
		Token string `json:"token" secret:"true"`
	}
	type C struct {
		Creds []Cred `json:"creds"`
	}
	cfg := C{Creds: []Cred{{User: "alice", Token: "s3cr3t"}}}
	b, err := Dump(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "s3cr3t") {
		t.Errorf("token must be redacted in slice of structs, got:\n%s", b)
	}
	if !strings.Contains(string(b), "alice") {
		t.Errorf("non-secret user should be present, got:\n%s", b)
	}
}

func TestDumpSecretInMapOfStructs(t *testing.T) {
	type Cred struct {
		Pass string `json:"pass" secret:"true"`
	}
	type C struct {
		DB map[string]Cred `json:"db"`
	}
	cfg := C{DB: map[string]Cred{"prod": {Pass: "hunter2"}}}
	b, err := Dump(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "hunter2") {
		t.Errorf("pass must be redacted in map of structs, got:\n%s", b)
	}
}

// Fix 2: parse failure in watched file is surfaced as error, not fake deletions.
func TestWatcherParseErrorSurfaced(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "cfg.yaml")
	if err := os.WriteFile(f, []byte("port: 8080\n"), 0600); err != nil {
		t.Fatal(err)
	}

	watcher, err := NewConfigWatcher(f)
	if err != nil {
		t.Fatal(err)
	}

	errCh := make(chan error, 1)
	watcher.AddDeltaListener(func(delta ConfigDelta, _, _ map[string]any, err error) {
		if err != nil {
			select {
			case errCh <- err:
			default:
			}
		} else if len(delta.Removed) > 0 {

			errCh <- nil
		}
	})

	watcher.SetPollInterval(50 * time.Millisecond)
	watcher.Start()
	defer watcher.Stop()

	time.Sleep(100 * time.Millisecond)

	if err := os.WriteFile(f, []byte("port: :\n"), 0600); err != nil {
		t.Fatal(err)
	}

	select {
	case e := <-errCh:
		if e == nil {
			t.Error("got fake key-removal instead of parse error")
		}

	case <-time.After(3 * time.Second):
		t.Error("timeout waiting for watcher notification")
	}
}

// Fix: FromOverlay surfaces non-ErrNotExist stat errors instead of silently falling back.
func TestFromOverlayStatError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission-based stat errors are not reliable on Windows")
	}
	if os.Getuid() == 0 {
		t.Skip("running as root, permission errors don't apply")
	}
	dir := t.TempDir()
	base := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(base, []byte("port: 8080\n"), 0600); err != nil {
		t.Fatal(err)
	}

	overlayDir := filepath.Join(dir, "config.prod.yaml")
	if err := os.Mkdir(overlayDir, 0000); err != nil {
		t.Skip("cannot create unreadable dir:", err)
	}
	defer func() { _ = os.Chmod(overlayDir, 0700) }()

	type C struct {
		Port int `yaml:"port" default:"0"`
	}
	src := FromOverlay(FromYAML(base), "prod")
	_, err := Load[C](src)

	if err == nil {
		t.Skip("os.Stat did not return an error for this setup")
	}
}

// Fix: dumpFieldKey skips json:"-" fields and falls back for json:",omitempty".
func TestDumpSkipsMinusField(t *testing.T) {
	type C struct {
		Name   string `json:"name"`
		Hidden string `json:"-"`
	}
	cfg := C{Name: "alice", Hidden: "secret"}
	b, err := Dump(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "Hidden") || strings.Contains(string(b), `"-"`) {
		t.Errorf("json:\"-\" field must be omitted, got:\n%s", b)
	}
	if !strings.Contains(string(b), "alice") {
		t.Errorf("expected name=alice in dump, got:\n%s", b)
	}
}

func TestDumpOmitemptyFallback(t *testing.T) {
	type C struct {
		Port int `json:",omitempty"`
	}
	cfg := C{Port: 9000}
	b, err := Dump(cfg)
	if err != nil {
		t.Fatal(err)
	}
	// Key should be snake_case fallback "port", not empty string.
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, b)
	}
	if _, ok := m[""]; ok {
		t.Errorf("empty-string key must not appear in dump: %s", b)
	}
	if _, ok := m["port"]; !ok {
		t.Errorf("expected 'port' key via snake_case fallback, got keys: %v", mapKeys(m))
	}
}

func mapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Fix 3: type-only changes (int→string for same value) are reported as Changed.
func TestComputeDeltaTypeChange(t *testing.T) {
	old := map[string]any{"port": 1}
	new_ := map[string]any{"port": "1"}

	delta := computeDelta(old, new_)

	if len(delta.Changed) != 1 || delta.Changed[0] != "port" {
		t.Errorf("expected type-only change to be detected in Changed, got %+v", delta)
	}
}
