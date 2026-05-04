package confkit

import (
	"fmt"
	"os"
	"path/filepath"
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
	filePath := filepath.Join(k.mountPath, field.Name)

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("cannot read configmap file %s: %w", filePath, err)
	}

	return string(data), true, nil
}

func FromKubernetesConfigMap(namespace, configMapName string) Source {
	defaultMountPath := filepath.Join("/var/run/secrets/config", namespace, configMapName)
	return NewKubernetesConfigMapSource(namespace, configMapName, defaultMountPath)
}

func FromKubernetesConfigMapWithPath(namespace, configMapName, mountPath string) Source {
	return NewKubernetesConfigMapSource(namespace, configMapName, mountPath)
}
