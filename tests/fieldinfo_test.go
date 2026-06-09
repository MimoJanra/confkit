package confkit_test

import (
	"testing"

	confkit "github.com/MimoJanra/confkit"
)

func TestScanFields(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		type cfg struct {
			Port int    `env:"PORT" default:"8080"`
			Mode string `env:"MODE" default:"dev"`
		}
		fields := confkit.ScanFields(cfg{})
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
		fields := confkit.ScanFields(&cfg{})
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
		fields := confkit.ScanFields(cfg{})
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
