package squid

import (
	"fmt"

	squidv1 "git.fr.clara.net/claranet/healthcare/buildops/projects/kubernetes/operators/squid-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteMany,
			},
			StorageClassName: &sr.Spec.StorageClassName,
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceName(corev1.ResourceStorage): resource.MustParse("10Gi"),
				},
			},
		},
	}
}
