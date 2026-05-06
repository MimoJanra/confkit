package k8s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MimoJanra/confkit"
	"github.com/MimoJanra/confkit/structtags"
)

type KubernetesConfigMapSource struct {
	namespace string
	configMap string
	mountPath string
}

func NewKubernetesConfigMapSource(namespace, configMapName, mountPath string) *KubernetesConfigMapSource {
	return &KubernetesConfigMapSource{
		namespace: namespace,
		configMap: configMapName,
		mountPath: mountPath,
	}
}

func (k *KubernetesConfigMapSource) Name() string {
	return "kubernetes-configmap"
}

func (k *KubernetesConfigMapSource) Lookup(_ context.Context, field *confkit.FieldInfo) (any, bool, error) {
	var keys []string

	for _, tag := range []string{"env", "yaml", "json"} {
		if tagKey, ok := field.Tags[tag]; ok && tagKey != "" {
			keys = append(keys, tagKey)
		}
	}

	snakeKey := structtags.SnakeCase(field.Name)
	keys = append(keys, snakeKey)
	if field.Name != snakeKey {
		keys = append(keys, field.Name)
	}

	for _, key := range keys {
		if strings.Contains(key, "/") || strings.Contains(key, string(filepath.Separator)) || strings.Contains(key, "..") {
			continue
		}
		filePath := filepath.Join(k.mountPath, key)
		cleanPath := filepath.Clean(filePath)
		basePath := filepath.Clean(k.mountPath) + string(filepath.Separator)
		if !strings.HasPrefix(cleanPath, basePath) {
			continue
		}
		data, err := os.ReadFile(cleanPath)
		if err == nil {
			return string(data), true, nil
		}
		if !os.IsNotExist(err) {
			return "", false, fmt.Errorf("cannot read configmap file %s: %w", cleanPath, err)
		}
	}

	return "", false, nil
}

func FromKubernetesConfigMap(namespace, configMapName string) confkit.Source {
	return FromKubernetesConfigMapWithPath(namespace, configMapName, "")
}

func FromKubernetesConfigMapWithPath(namespace, configMapName, mountPath string) confkit.Source {
	if mountPath == "" {
		mountPath = filepath.Join("/var/run/secrets/config", namespace, configMapName)
	}
	return NewKubernetesConfigMapSource(namespace, configMapName, mountPath)
}
