package confkit

import (
	"os"
	"testing"
	"time"
)

func TestNewConfigWatcher(t *testing.T) {
	tmpFile := writeTempYAML(t, "Port: 8080")
	defer func() { _ = os.Remove(tmpFile) }()

	watcher, err := NewConfigWatcher(tmpFile)
	if err != nil {
		t.Fatalf("NewConfigWatcher failed: %v", err)
	}

	if watcher.filePath != tmpFile {
		t.Errorf("Expected filePath=%s, got %s", tmpFile, watcher.filePath)
	}
	if watcher.lastModTime.IsZero() {
		t.Error("Expected lastModTime to be set")
	}
}

func TestConfigWatcherFileNotFound(t *testing.T) {
	_, err := NewConfigWatcher("/nonexistent/file.yaml")
	if err == nil {
		t.Fatal("Expected error for nonexistent file")
	}
}

func TestConfigWatcherDetectsChanges(t *testing.T) {
	tmpFile := writeTempYAML(t, "Port: 8080")
	defer func() { _ = os.Remove(tmpFile) }()

	watcher, err := NewConfigWatcher(tmpFile)
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

	watcher, err := NewConfigWatcher(tmpFile)
	if err != nil {
		t.Fatalf("NewConfigWatcher failed: %v", err)
	}

	count := 0
	listener := func(oldCfg, newCfg any, err error) {
		count++
	}

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

	watcher, err := NewConfigWatcher(tmpFile)
	if err != nil {
		t.Fatalf("NewConfigWatcher failed: %v", err)
	}

	watcher.SetPollInterval(100 * time.Millisecond)
	watcher.Start()
	time.Sleep(150 * time.Millisecond)

	watcher.Stop()

	select {
	case <-watcher.done:
	case <-time.After(1 * time.Second):
		t.Fatal("Watcher did not stop cleanly")
	}
}
