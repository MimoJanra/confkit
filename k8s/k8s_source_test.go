package k8s

import (
	"testing"

	"github.com/MimoJanra/confkit"
)

func TestKubernetesSourceIntegration(t *testing.T) {
	src := FromKubernetesConfigMap("default", "app-config")
	if src == nil {
		t.Fatal("Expected non-nil Kubernetes source")
	}

	if src.Name() != "kubernetes-configmap" {
		t.Errorf("Expected kubernetes-configmap source, got %q", src.Name())
	}
}

func TestSourceNamingConsistency(t *testing.T) {
	sources := map[string]confkit.Source{
		"env": confkit.FromEnv(),
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
