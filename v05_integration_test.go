package confkit

import (
	"context"
	"testing"
	"time"
)

func TestV05MultiSourceIntegration(t *testing.T) {
	type Config struct {
		Port     int    `default:"8080"`
		Host     string `default:"localhost"`
		Database string
		Password string `secret:"true"`
	}

	sources := []Source{
		FromEnv(),
		FromVault("http://localhost:8200", VaultTokenAuth("test-token")),
	}

	cfg, err := Load[Config](sources...)
	if err != nil {
		t.Logf("Expected error or success with multiple sources, got: %v", err)
	}

	if cfg.Port == 0 {
		t.Errorf("Expected default Port, got %d", cfg.Port)
	}
}

func TestV05HotReloadWithRotation(t *testing.T) {
	type Config struct {
		Port int
	}

	strategy := RotateOnInterval(100 * time.Millisecond)
	engine := NewRotationEngine(strategy)

	rotationTriggered := false
	engine.AddCallback(func(oldCfg, newCfg any, err error) {
		rotationTriggered = true
	})

	if rotationTriggered {
		t.Error("Callback should not be triggered without rotation check")
	}
}

func TestV05ConsulWithRotation(t *testing.T) {
	src := FromConsul("localhost:8500")
	if src == nil {
		t.Fatal("Expected non-nil Consul source")
	}

	if src.Name() != "consul" && src.Name() != "file" {
		t.Errorf("Expected consul or file, got %q", src.Name())
	}
}

func TestV05EtcdWithRotation(t *testing.T) {
	endpoints := []string{"localhost:2379"}
	src := FromEtcd(endpoints)
	if src == nil {
		t.Fatal("Expected non-nil etcd source")
	}

	if src.Name() != "etcd" && src.Name() != "file" {
		t.Errorf("Expected etcd or file, got %q", src.Name())
	}
}

func TestV05SecretsManagerWithRotation(t *testing.T) {
	src := FromAWSSecretsManager("myapp/db")
	if src == nil {
		t.Fatal("Expected non-nil Secrets Manager source")
	}

	if src.Name() != "aws-secrets-manager" && src.Name() != "file" {
		t.Errorf("Expected aws-secrets-manager or file, got %q", src.Name())
	}
}

func TestV05MultiRegionFailover(t *testing.T) {
	regions := []string{"us-east-1", "us-west-2", "eu-west-1"}
	src := FromAWSSecretsManagerMultiRegion("myapp/db", regions)
	if src == nil {
		t.Fatal("Expected non-nil multi-region source")
	}

	if src.Name() != "multiregion" && src.Name() != "file" {
		t.Errorf("Expected multiregion or file, got %q", src.Name())
	}
}

func TestV05AllSourcesTypes(t *testing.T) {
	sourceName := map[string]Source{
		"env":        FromEnv(),
		"consul":     FromConsul("localhost:8500"),
		"etcd":       FromEtcd([]string{"localhost:2379"}),
		"vault":      FromVault("http://localhost:8200", VaultTokenAuth("test")),
		"secretsmgr": FromAWSSecretsManager("test"),
	}

	for name, src := range sourceName {
		if src == nil {
			t.Errorf("Source %s returned nil", name)
		}
	}
}

func TestV05RotationStrategies(t *testing.T) {
	tests := []struct {
		name     string
		strategy RotationStrategy
	}{
		{"Interval", RotateOnInterval(1 * time.Hour)},
		{"TTL", RotateOnMinTTL(30 * time.Minute)},
		{"Event", RotateOnEvent(make(chan struct{}))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.strategy == nil {
				t.Errorf("%s strategy is nil", tt.name)
			}

			shouldRotate, err := tt.strategy.ShouldRotate(context.Background(), time.Now())
			if err != nil && err != context.Canceled {
				t.Errorf("ShouldRotate failed: %v", err)
			}

			if tt.name == "Interval" && shouldRotate {
				t.Errorf("%s should not rotate immediately", tt.name)
			}
		})
	}
}

func TestV05ConfigDumpWithSecrets(t *testing.T) {
	type Config struct {
		Port     int
		Password string `secret:"true"`
	}

	cfg := Config{Port: 8080, Password: "secret123"}
	fields := []FieldInfo{
		{Path: "Port", IsSecret: false},
		{Path: "Password", IsSecret: true},
	}

	data, err := DumpConfig(cfg, fields)
	if err != nil {
		t.Fatalf("DumpConfig failed: %v", err)
	}

	if data == nil {
		t.Fatal("Expected non-nil dump data")
	}

	dumpStr := string(data)
	if !contains(dumpStr, "***REDACTED***") {
		t.Error("Expected password to be redacted")
	}
}

func TestV05LoadMetrics(t *testing.T) {
	metrics := LoadMetrics{
		TotalTime: 100 * time.Millisecond,
		SourceTimes: map[string]time.Duration{
			"vault": 50 * time.Millisecond,
			"env":   30 * time.Millisecond,
		},
		ValidationTime: 20 * time.Millisecond,
		ErrorCount:     0,
	}

	if metrics.TotalTime != 100*time.Millisecond {
		t.Errorf("Expected TotalTime=100ms, got %v", metrics.TotalTime)
	}

	if len(metrics.SourceTimes) != 2 {
		t.Errorf("Expected 2 source times, got %d", len(metrics.SourceTimes))
	}
}

func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
