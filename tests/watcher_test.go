package confkit_test

import (
	"os"
	"testing"
	"time"

	confkit "github.com/MimoJanra/confkit"
)

func TestNewConfigWatcher(t *testing.T) {
	tmpFile := writeTempYAML(t, "Port: 8080")
	defer func() { _ = os.Remove(tmpFile) }()

	watcher, err := confkit.NewConfigWatcher(tmpFile)
	if err != nil {
		t.Fatalf("NewConfigWatcher failed: %v", err)
	}
	if watcher == nil {
		t.Fatal("expected non-nil watcher")
	}
}

func TestConfigWatcherFileNotFound(t *testing.T) {
	_, err := confkit.NewConfigWatcher("/nonexistent/file.yaml")
	if err == nil {
		t.Fatal("Expected error for nonexistent file")
	}
}

func TestConfigWatcherDetectsChanges(t *testing.T) {
	tmpFile := writeTempYAML(t, "Port: 8080")
	defer func() { _ = os.Remove(tmpFile) }()

	watcher, err := confkit.NewConfigWatcher(tmpFile)
	if err != nil {
		t.Fatalf("NewConfigWatcher failed: %v", err)
	}

	changeDetected := false
	watcher.AddListener(func(oldCfg, newCfg any, err error) {
		changeDetected = true
	})

	watcher.SetPollInterval(100 * time.Millisecond)
	watcher.Start()
	time.Sleep(150 * time.Millisecond)

	if err := os.WriteFile(tmpFile, []byte("Port: 9000"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	time.Sleep(300 * time.Millisecond)
	watcher.Stop()

	if !changeDetected {
		t.Error("Expected file change to be detected")
	}
}

func TestConfigWatcherMultipleListeners(t *testing.T) {
	tmpFile := writeTempYAML(t, "Port: 8080")
	defer func() { _ = os.Remove(tmpFile) }()

	watcher, err := confkit.NewConfigWatcher(tmpFile)
	if err != nil {
		t.Fatalf("NewConfigWatcher failed: %v", err)
	}

	count := 0
	listener := func(oldCfg, newCfg any, err error) { count++ }

	watcher.AddListener(listener)
	watcher.AddListener(listener)
	watcher.AddListener(listener)

	watcher.SetPollInterval(100 * time.Millisecond)
	watcher.Start()
	time.Sleep(150 * time.Millisecond)

	if err := os.WriteFile(tmpFile, []byte("Port: 9000"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	time.Sleep(300 * time.Millisecond)
	watcher.Stop()

	if count != 3 {
		t.Errorf("Expected 3 listener notifications, got %d", count)
	}
}

func TestConfigWatcherStopsCleaning(t *testing.T) {
	tmpFile := writeTempYAML(t, "Port: 8080")
	defer func() { _ = os.Remove(tmpFile) }()

	watcher, err := confkit.NewConfigWatcher(tmpFile)
	if err != nil {
		t.Fatalf("NewConfigWatcher failed: %v", err)
	}

	watcher.SetPollInterval(100 * time.Millisecond)
	watcher.Start()
	time.Sleep(150 * time.Millisecond)

	stopped := make(chan struct{})
	go func() {
		watcher.Stop()
		close(stopped)
	}()

	select {
	case <-stopped:
	case <-time.After(2 * time.Second):
		t.Fatal("Watcher did not stop cleanly")
	}
}

func TestConfigWatcherStartIdempotent(t *testing.T) {
	tmpFile := writeTempYAML(t, "Port: 8080")
	defer func() { _ = os.Remove(tmpFile) }()

	watcher, err := confkit.NewConfigWatcher(tmpFile)
	if err != nil {
		t.Fatalf("NewConfigWatcher failed: %v", err)
	}

	watcher.SetPollInterval(100 * time.Millisecond)
	watcher.Start()
	watcher.Start()
	watcher.Start()
	time.Sleep(150 * time.Millisecond)

	stopped := make(chan struct{})
	go func() {
		watcher.Stop()
		close(stopped)
	}()

	select {
	case <-stopped:
	case <-time.After(2 * time.Second):
		t.Fatal("Watcher did not stop cleanly after multiple Start() calls")
	}
}

func TestLoadWithWatcher(t *testing.T) {
	type Cfg struct {
		Port int `yaml:"Port"`
	}
	tmpFile := writeTempYAML(t, "Port: 3000")
	defer func() { _ = os.Remove(tmpFile) }()

	cfg, watcher, err := confkit.LoadWithWatcher[Cfg](tmpFile, confkit.FromYAML(tmpFile))
	if err != nil {
		t.Fatalf("LoadWithWatcher failed: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if watcher == nil {
		t.Fatal("expected non-nil watcher")
	}
	if cfg.Port != 3000 {
		t.Errorf("expected Port 3000, got %d", cfg.Port)
	}
	// Start then Stop to exercise the full lifecycle without blocking
	watcher.SetPollInterval(100 * time.Millisecond)
	watcher.Start()
	stopped := make(chan struct{})
	go func() { watcher.Stop(); close(stopped) }()
	select {
	case <-stopped:
	case <-time.After(2 * time.Second):
		t.Fatal("watcher did not stop cleanly")
	}
}

func TestLoadWithWatcherError(t *testing.T) {
	type Cfg struct {
		Port int `yaml:"Port"`
	}
	// NewConfigWatcher fails for nonexistent file, Load[T] with no sources succeeds
	_, _, err := confkit.LoadWithWatcher[Cfg]("/nonexistent/path.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestAddDeltaListener(t *testing.T) {
	tmpFile := writeTempYAML(t, "port: 8080")
	defer func() { _ = os.Remove(tmpFile) }()

	watcher, err := confkit.NewConfigWatcher(tmpFile)
	if err != nil {
		t.Fatalf("NewConfigWatcher failed: %v", err)
	}

	var gotDelta confkit.ConfigDelta
	watcher.AddDeltaListener(func(delta confkit.ConfigDelta, _, _ map[string]any, err error) {
		gotDelta = delta
	})

	watcher.SetPollInterval(100 * time.Millisecond)
	watcher.Start()
	time.Sleep(150 * time.Millisecond)

	if err := os.WriteFile(tmpFile, []byte("port: 9000\nnew_key: added"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	time.Sleep(300 * time.Millisecond)
	watcher.Stop()

	if len(gotDelta.Changed)+len(gotDelta.Added) == 0 {
		t.Error("expected delta with changed or added keys")
	}
}

func TestConfigWatcherStopWithoutStart(t *testing.T) {
	tmpFile := writeTempYAML(t, "Port: 8080")
	defer func() { _ = os.Remove(tmpFile) }()

	watcher, err := confkit.NewConfigWatcher(tmpFile)
	if err != nil {
		t.Fatalf("NewConfigWatcher failed: %v", err)
	}

	done := make(chan struct{})
	go func() {
		watcher.Stop()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Stop() on a never-started watcher must not block")
	}
}

func TestConfigWatcherDoubleStop(t *testing.T) {
	tmpFile := writeTempYAML(t, "Port: 8080")
	defer func() { _ = os.Remove(tmpFile) }()

	watcher, err := confkit.NewConfigWatcher(tmpFile)
	if err != nil {
		t.Fatalf("NewConfigWatcher failed: %v", err)
	}
	watcher.Start()

	done := make(chan struct{})
	go func() {
		watcher.Stop()
		watcher.Stop()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("repeated Stop() must not block")
	}
}

func TestLoadWithWatcherStopWithoutStart(t *testing.T) {
	type Cfg struct {
		Port int `env:"PORT" default:"8080"`
	}

	tmpFile := writeTempYAML(t, "Port: 8080")
	defer func() { _ = os.Remove(tmpFile) }()

	_, watcher, err := confkit.LoadWithWatcher[Cfg](tmpFile, confkit.FromEnv())
	if err != nil {
		t.Fatalf("LoadWithWatcher failed: %v", err)
	}

	done := make(chan struct{})
	go func() {
		watcher.Stop()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("LoadWithWatcher returns an unstarted watcher, so Stop() must not block")
	}
}
