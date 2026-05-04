package schema

import (
	"strings"
	"testing"
)

func TestCLIHelpWithShortFlags(t *testing.T) {
	type Config struct {
		Port  int    `short:"p" desc:"HTTP port" default:"8080"`
		Host  string `short:"h" desc:"HTTP host" default:"localhost"`
		Debug bool   `short:"d" desc:"Debug mode"`
	}

	help, err := GenerateCLIHelp[Config]()
	if err != nil {
		t.Fatalf("GenerateCLIHelp failed: %v", err)
	}

	if !strings.Contains(help, "-p, --port") {
		t.Errorf("Expected short flag -p not found in help")
	}
	if !strings.Contains(help, "-h, --host") {
		t.Errorf("Expected short flag -h not found in help")
	}
	if !strings.Contains(help, "-d, --debug") {
		t.Errorf("Expected short flag -d not found in help")
	}
}

func TestCLIHelpWithHiddenFields(t *testing.T) {
	type Config struct {
		Port  int    `desc:"HTTP port" default:"8080"`
		Debug bool   `hidden:"true" desc:"Debug mode"`
		Token string `hidden:"true" desc:"API token"`
	}

	help, err := GenerateCLIHelp[Config]()
	if err != nil {
		t.Fatalf("GenerateCLIHelp failed: %v", err)
	}

	if !strings.Contains(help, "--port") {
		t.Errorf("Expected visible field --port not found in help")
	}
	if strings.Contains(help, "--debug") {
		t.Errorf("Expected hidden field --debug to be hidden from help")
	}
	if strings.Contains(help, "--token") {
		t.Errorf("Expected hidden field --token to be hidden from help")
	}
}

func TestCLIHelpMixed(t *testing.T) {
	type Config struct {
		Port     int    `short:"p" desc:"HTTP port" validate:"min=1,max=65535"`
		Host     string `desc:"HTTP host"`
		Debug    bool   `hidden:"true"`
		LogLevel string `short:"l" validate:"oneof=debug,info,warn,error"`
	}

	help, err := GenerateCLIHelp[Config]()
	if err != nil {
		t.Fatalf("GenerateCLIHelp failed: %v", err)
	}

	if !strings.Contains(help, "-p, --port") {
		t.Errorf("Expected short flag -p")
	}
	if !strings.Contains(help, "--host") {
		t.Errorf("Expected --host flag")
	}
	if strings.Contains(help, "--debug") {
		t.Errorf("Expected --debug to be hidden")
	}
	if !strings.Contains(help, "-l, --log-level") {
		t.Errorf("Expected short flag -l")
	}
}
