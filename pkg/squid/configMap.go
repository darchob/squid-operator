package squid

import (
	"fmt"

	squidv1 "git.fr.clara.net/claranet/healthcare/buildops/projects/kubernetes/operators/squid-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Init secondary configMap
func ConfigMap(sr *squidv1.SquidInstance) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        sr.Name,
			Namespace:   sr.Namespace,
			Labels:      Labels(sr),
			Annotations: Annotations(),
		},
		Data: map[string]string{
			"squid-init.conf": fmt.Sprintf("%s\n", "# INITITAL CONFIGS"),
		},
	}
}
