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

// KubernetesConfigMapSource reads values from a ConfigMap that has been mounted as a
// directory of files, one file per key. It talks to the filesystem rather than the
// Kubernetes API, so it needs no cluster credentials.
type KubernetesConfigMapSource struct {
	namespace string
	configMap string
	mountPath string
}

// NewKubernetesConfigMapSource returns a source reading from mountPath. Prefer
// FromKubernetesConfigMap, which supplies the conventional mount path.
func NewKubernetesConfigMapSource(namespace, configMapName, mountPath string) *KubernetesConfigMapSource {
	return &KubernetesConfigMapSource{
		namespace: namespace,
		configMap: configMapName,
		mountPath: mountPath,
	}
}

// Name returns "kubernetes-configmap".
func (k *KubernetesConfigMapSource) Name() string {
	return "kubernetes-configmap"
}

// Lookup reads the file whose name matches the field, trying its `env`, `yaml` and
// `json` tags, then its snake_cased name, then its Go name.
//
// Keys that contain a path separator or "..", or that would resolve outside the mount
// path, are ignored so a crafted tag cannot escape the directory. A missing file means
// not found; any other read failure is an error.
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
		data, err := os.ReadFile(cleanPath) // #nosec G304
		if err == nil {
			return string(data), true, nil
		}
		if !os.IsNotExist(err) {
			return "", false, fmt.Errorf("cannot read configmap file %s: %w", cleanPath, err)
		}
	}

	return "", false, nil
}

// FromKubernetesConfigMap reads a ConfigMap mounted at the conventional location,
// /var/run/secrets/config/<namespace>/<configMapName>.
func FromKubernetesConfigMap(namespace, configMapName string) confkit.Source {
	return FromKubernetesConfigMapWithPath(namespace, configMapName, "")
}

// FromKubernetesConfigMapWithPath reads a ConfigMap mounted at mountPath. An empty
// mountPath falls back to the conventional location used by FromKubernetesConfigMap.
func FromKubernetesConfigMapWithPath(namespace, configMapName, mountPath string) confkit.Source {
	if mountPath == "" {
		mountPath = filepath.Join("/var/run/secrets/config", namespace, configMapName)
	}
	return NewKubernetesConfigMapSource(namespace, configMapName, mountPath)
}
