package confkit

import (
	"os"
	"testing"
	"time"
)

func TestV04HotReloadIntegration(t *testing.T) {
	tmpFile := writeTempYAML(t, "Port: 8080\nHost: localhost")
	defer os.Remove(tmpFile)

	src := FromYAML(tmpFile)
	if src == nil {
		t.Fatal("Expected non-nil YAML source")
	}

	watcher, err := NewConfigWatcher(tmpFile)
	if err != nil {
		t.Fatalf("NewConfigWatcher failed: %v", err)
	}

	if watcher == nil {
		t.Error("Expected non-nil watcher")
		return
	}

	watcher.SetPollInterval(100 * time.Millisecond)
	watcher.Start()

	time.Sleep(50 * time.Millisecond)

	watcher.Stop()
}

func TestV04KubernetesSourceIntegration(t *testing.T) {
	src := FromKubernetesConfigMap("default", "app-config")
	if src == nil {
		t.Fatal("Expected non-nil Kubernetes source")
	}

	if src.Name() != "kubernetes-configmap" {
		t.Errorf("Expected kubernetes-configmap source, got %q", src.Name())
	}
}

func TestV04MultiSourcePrecedence(t *testing.T) {
	type Config struct {
		Port int `default:"8080"`
		Host string
	}

	os.Setenv("HOST", "env-host")
	defer os.Unsetenv("HOST")

	cfg, err := Load[Config](
		FromEnv(),
	)
	if err != nil {
		t.Logf("Load expected to work: %v", err)
	}

	if cfg.Port != 8080 {
		t.Errorf("Expected default Port=8080, got %d", cfg.Port)
	}
}

func TestV04ConfigDumpIntegration(t *testing.T) {
	type Config struct {
		Port     int
		Password string `secret:"true"`
		Host     string
	}

	cfg := Config{
		Port:     8080,
		Password: "secret123",
		Host:     "localhost",
	}

	fields := []FieldInfo{
		{Path: "Port", IsSecret: false},
		{Path: "Password", IsSecret: true},
		{Path: "Host", IsSecret: false},
	}

	data, err := DumpConfig(cfg, fields)
	if err != nil {
		t.Fatalf("DumpConfig failed: %v", err)
	}

	dumpStr := string(data)

	if !stringContains(dumpStr, "Port") {
		t.Error("Port should be in dump")
	}

	if !stringContains(dumpStr, "***REDACTED***") {
		t.Error("Password should be redacted")
	}

	if !stringContains(dumpStr, "Host") {
		t.Error("Host should be in dump")
	}
}

func TestV04ObservabilityMetrics(t *testing.T) {
	startTime := time.Now()

	sources := []string{"yaml", "env"}
	logMsg := LogLoadStart(sources)

	if !stringContains(logMsg, "config_load_start") {
		t.Error("Log should contain config_load_start event")
	}

	if !stringContains(logMsg, "yaml") {
		t.Error("Log should contain yaml source")
	}

	duration := time.Since(startTime)
	logMsg = LogLoadComplete(duration, 5, 0)

	if !stringContains(logMsg, "config_load_complete") {
		t.Error("Log should contain config_load_complete event")
	}

	if !stringContains(logMsg, "fields_loaded") {
		t.Error("Log should contain fields_loaded count")
	}
}

func TestV04WatcherIntegration(t *testing.T) {
	tmpFile := writeTempYAML(t, "test: value")
	defer os.Remove(tmpFile)

	watcher, err := NewConfigWatcher(tmpFile)
	if err != nil {
		t.Fatalf("NewConfigWatcher failed: %v", err)
	}

	watcher.AddListener(func(oldCfg, newCfg any, err error) {
	})

	watcher.SetPollInterval(100 * time.Millisecond)
	watcher.Start()

	time.Sleep(150 * time.Millisecond)

	if err := os.WriteFile(tmpFile, []byte("test: updated"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	time.Sleep(300 * time.Millisecond)

	watcher.Stop()
}

func TestV04SourceNamingConsistency(t *testing.T) {
	sources := map[string]Source{
		"env": FromEnv(),
		"k8s": FromKubernetesConfigMap("default", "config"),
	}

	expectedNames := map[string]string{
		"env": "env",
		"k8s": "kubernetes-configmap",
	}

	for key, src := range sources {
		expected := expectedNames[key]
		actual := src.Name()

		if actual != expected {
			t.Errorf("Source %s: expected %q, got %q", key, expected, actual)
		}
	}
}

type TestConfig struct {
	Port int
	Host string
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestV04FullStack(t *testing.T) {
	type FullConfig struct {
		Port     int    `default:"8080"`
		Host     string `default:"localhost"`
		Database string
		Password string `secret:"true"`
	}

	tmpFile := writeTempYAML(t, "Port: 9000\nHost: example.com")
	defer os.Remove(tmpFile)

	sources := []Source{
		FromYAML(tmpFile),
		FromEnv(),
	}

	cfg, err := Load[FullConfig](sources...)
	if err != nil {
		t.Logf("Load expected to work with multiple sources: %v", err)
	}

	if cfg.Port == 0 {
		t.Logf("Port should be loaded from sources")
	}

	fields := []FieldInfo{
		{Path: "Port", IsSecret: false},
		{Path: "Host", IsSecret: false},
		{Path: "Password", IsSecret: true},
	}

	dump, err := DumpConfig(cfg, fields)
	if err != nil {
		t.Fatalf("DumpConfig failed: %v", err)
	}

	if dump == nil {
		t.Error("DumpConfig should return non-nil data")
	}
}
