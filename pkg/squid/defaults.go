package squid

import (
	"fmt"
	"time"

	squidv1 "git.fr.clara.net/claranet/healthcare/buildops/projects/kubernetes/operators/squid-operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	DuplicateError = "found duplicate rule : "

//	defaultData    = `
//
// include /etc/squid/squid.d/*.conf
// cache_dir ufs $prefix/var/cache 100 16 256
// `
)

func Labels(sr *squidv1.SquidInstance) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       sr.Name,
		"app.kubernetes.io/version":    sr.Spec.Image.Tags,
		"app.kubernetes.io/part-of":    fmt.Sprintf("squid-%s", sr.Name),
		"app.kubernetes.io/managed-by": "squid-operator",
	}
}

func Annotations() map[string]string {
	return map[string]string{
		"date": time.DateTime,
	}
}

func ObjectMeta(sr *squidv1.SquidInstance) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:        sr.Name,
		Namespace:   sr.Namespace,
		Labels:      Labels(sr),
		Annotations: Annotations(),
		OwnerReferences: []metav1.OwnerReference{
			{
				Name:       sr.Name,
				Kind:       sr.Kind,
				APIVersion: sr.APIVersion,
			},
		},
	}
}
