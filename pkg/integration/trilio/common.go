package trilio

import (
	"context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func findTrilioInstallation(ctx context.Context, client kubernetes.Interface, namespace string) (*v1.ConfigMap, error) {
	trilioCofigmapName := "k8s-triliovault-config"
	cm, err := client.CoreV1().ConfigMaps(namespace).Get(ctx, trilioCofigmapName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return cm, nil
}
