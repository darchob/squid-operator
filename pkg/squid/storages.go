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
		ObjectMeta: *ObjectMeta(sr),
		Data: map[string]string{
			"squid-init.conf": fmt.Sprintf("%s\n", "# INITITAL CONFIGS"),
		},
	}
}

func PersistentVolumeClaim(sr *squidv1.SquidInstance) *corev1.PersistentVolumeClaim {
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: *ObjectMeta(sr),
		Spec: corev1.PersistentVolumeClaimSpec{
			VolumeName: fmt.Sprintf("%s-volume", sr.Name),
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: Labels(sr),
			},
			StorageClassName: &sr.Spec.StorageClassName,
		},
	}
}
