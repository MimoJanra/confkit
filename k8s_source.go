package confkit

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/MimoJanra/confkit/tagutil"
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

func (k *KubernetesConfigMapSource) Lookup(field *FieldInfo) (any, bool, error) {
	var keys []string

	keys = append(keys, field.Name)

	for _, tag := range []string{"env", "yaml", "json"} {
		if tagKey, ok := field.Tags[tag]; ok {
			keys = append(keys, tagKey)
		}
	}

	keys = append(keys, tagutil.SnakeCase(field.Name))

	for _, key := range keys {
		filePath := filepath.Join(k.mountPath, key)
		data, err := os.ReadFile(filePath)
		if err == nil {
			return string(data), true, nil
		}
		if !os.IsNotExist(err) {
			return "", false, fmt.Errorf("cannot read configmap file %s: %w", filePath, err)
		}
	}

	return "", false, nil
}

func FromKubernetesConfigMap(namespace, configMapName string) Source {
	return FromKubernetesConfigMapWithPath(namespace, configMapName, "")
}

func FromKubernetesConfigMapWithPath(namespace, configMapName, mountPath string) Source {
	if mountPath == "" {
		mountPath = filepath.Join("/var/run/secrets/config", namespace, configMapName)
	}
	return NewKubernetesConfigMapSource(namespace, configMapName, mountPath)
}
