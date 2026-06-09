package confkit

import (
	"os"
	"testing"
	"time"
)

func TestComputeDelta(t *testing.T) {
	t.Run("added", func(t *testing.T) {
		old := map[string]any{"a": "1"}
		new := map[string]any{"a": "1", "b": "2"}
		delta := computeDelta(old, new)
		if len(delta.Added) != 1 || delta.Added[0] != "b" {
			t.Errorf("expected added=[b], got %v", delta.Added)
		}
		if len(delta.Removed) != 0 || len(delta.Changed) != 0 {
			t.Errorf("unexpected removed/changed: %v / %v", delta.Removed, delta.Changed)
		}
	})

	t.Run("removed", func(t *testing.T) {
		old := map[string]any{"a": "1", "b": "2"}
		new := map[string]any{"a": "1"}
		delta := computeDelta(old, new)
		if len(delta.Removed) != 1 || delta.Removed[0] != "b" {
			t.Errorf("expected removed=[b], got %v", delta.Removed)
		}
	})

	t.Run("changed", func(t *testing.T) {
		old := map[string]any{"a": "1"}
		new := map[string]any{"a": "2"}
		delta := computeDelta(old, new)
		if len(delta.Changed) != 1 || delta.Changed[0] != "a" {
			t.Errorf("expected changed=[a], got %v", delta.Changed)
		}
	})

	t.Run("no_change", func(t *testing.T) {
		snap := map[string]any{"a": "1", "b": "2"}
		delta := computeDelta(snap, snap)
		if len(delta.Added)+len(delta.Removed)+len(delta.Changed) != 0 {
			t.Errorf("expected empty delta, got %+v", delta)
		}
	})
}

func TestComputeDeltaTypeChange(t *testing.T) {
	old := map[string]any{"x": 1}
	new := map[string]any{"x": "one"}
	delta := computeDelta(old, new)
	if len(delta.Changed) != 1 || delta.Changed[0] != "x" {
		t.Errorf("expected changed=[x] on type change, got %v", delta.Changed)
	}
}

func TestFlattenAnyMap(t *testing.T) {
	t.Run("flat", func(t *testing.T) {
		m := map[string]any{"a": "1", "b": "2"}
		got := flattenAnyMap("", m)
		if got["a"] != "1" || got["b"] != "2" {
			t.Errorf("unexpected flat map: %v", got)
		}
	})

	t.Run("nested", func(t *testing.T) {
		m := map[string]any{
			"db": map[string]any{"host": "localhost", "port": 5432},
		}
		got := flattenAnyMap("", m)
		if got["db.host"] != "localhost" {
			t.Errorf("expected db.host=localhost, got %v", got["db.host"])
		}
		if got["db.port"] != 5432 {
			t.Errorf("expected db.port=5432, got %v", got["db.port"])
		}
	})

	t.Run("with_prefix", func(t *testing.T) {
		m := map[string]any{"x": "1"}
		got := flattenAnyMap("root", m)
		if got["root.x"] != "1" {
			t.Errorf("expected root.x=1, got %v", got)
		}
	})
}

func TestAddDeltaListener(t *testing.T) {
	tmpFile := writeTempYAML(t, "port: 8080")
	defer func() { _ = os.Remove(tmpFile) }()

	watcher, err := NewConfigWatcher(tmpFile)
	if err != nil {
		t.Fatalf("NewConfigWatcher failed: %v", err)
	}

	var gotDelta ConfigDelta
	watcher.AddDeltaListener(func(delta ConfigDelta, _, _ map[string]any, err error) {
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

func TestParseFileToFlatMap(t *testing.T) {
	t.Run("yaml", func(t *testing.T) {
		f := writeTempYAML(t, "host: localhost\nport: 5432")
		defer func() { _ = os.Remove(f) }()

		m, err := parseFileToFlatMap(f)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if m["host"] != "localhost" {
			t.Errorf("expected host=localhost, got %v", m["host"])
		}
	})

	t.Run("json", func(t *testing.T) {
		f := writeTempJSON(t, `{"host":"jsonhost","port":3000}`)
		defer func() { _ = os.Remove(f) }()

		m, err := parseFileToFlatMap(f)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if m["host"] != "jsonhost" {
			t.Errorf("expected host=jsonhost, got %v", m["host"])
		}
	})

	t.Run("toml", func(t *testing.T) {
		f := writeTempTOML(t, "host = \"tomlhost\"\n")
		defer func() { _ = os.Remove(f) }()

		m, err := parseFileToFlatMap(f)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if m["host"] != "tomlhost" {
			t.Errorf("expected host=tomlhost, got %v", m["host"])
		}
	})

	t.Run("nested_flattened", func(t *testing.T) {
		f := writeTempYAML(t, "db:\n  host: nested-host\n  port: 5432")
		defer func() { _ = os.Remove(f) }()

		m, err := parseFileToFlatMap(f)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if m["db.host"] != "nested-host" {
			t.Errorf("expected db.host=nested-host, got %v", m["db.host"])
		}
	})

	t.Run("missing_file_error", func(t *testing.T) {
		_, err := parseFileToFlatMap("/nonexistent/path.yaml")
		if err == nil {
			t.Fatal("expected error for missing file")
		}
	})
}

func TestWatcherParseErrorSurfaced(t *testing.T) {
	tmpFile := writeTempYAML(t, "port: 8080")
	defer func() { _ = os.Remove(tmpFile) }()

	watcher, err := NewConfigWatcher(tmpFile)
	if err != nil {
		t.Fatalf("NewConfigWatcher failed: %v", err)
	}

	var parseErr error
	watcher.AddDeltaListener(func(_ ConfigDelta, _, _ map[string]any, err error) {
		if err != nil {
			parseErr = err
		}
	})

	watcher.SetPollInterval(100 * time.Millisecond)
	watcher.Start()
	time.Sleep(150 * time.Millisecond)

	if err := os.WriteFile(tmpFile, []byte(":\tinvalid yaml:::"), 0644); err != nil {
		t.Fatalf("failed to write invalid yaml: %v", err)
	}

	time.Sleep(300 * time.Millisecond)
	watcher.Stop()

	if parseErr == nil {
		t.Error("expected parse error to be surfaced to delta listener")
	}
}

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

func TestConfigWatcherStartIdempotent(t *testing.T) {
	tmpFile := writeTempYAML(t, "Port: 8080")
	defer func() { _ = os.Remove(tmpFile) }()

	watcher, err := NewConfigWatcher(tmpFile)
	if err != nil {
		t.Fatalf("NewConfigWatcher failed: %v", err)
	}

	watcher.SetPollInterval(100 * time.Millisecond)
	watcher.Start()
	watcher.Start()
	watcher.Start()

	time.Sleep(150 * time.Millisecond)

	watcher.Stop()

	select {
	case <-watcher.done:
	case <-time.After(1 * time.Second):
		t.Fatal("Watcher did not stop cleanly after multiple Start() calls")
	}
}
