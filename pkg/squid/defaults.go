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

func Labels(obj interface{}) map[string]string {
	switch sr := obj.(type) {
	case *squidv1.SquidInstance:
		return map[string]string{
			"app.kubernetes.io/name":       sr.Name,
			"app.kubernetes.io/version":    sr.Spec.Image.Tag,
			"app.kubernetes.io/part-of":    fmt.Sprintf("squid-%s", sr.Name),
			"app.kubernetes.io/managed-by": "squid-operator",
		}
	case *squidv1.SquidConfigs:
		return map[string]string{
			"app.kubernetes.io/name":       sr.Name,
			"app.kubernetes.io/part-of":    fmt.Sprintf("squid-%s", sr.Name),
			"app.kubernetes.io/managed-by": "squid-operator",
		}
	}
	return nil
}

func Annotations() map[string]string {
	return map[string]string{
		"date": time.DateTime,
	}
}

func ObjectMeta(obj interface{}) *metav1.ObjectMeta {
	switch sr := obj.(type) {
	case *squidv1.SquidInstance:
		return &metav1.ObjectMeta{
			Name:        sr.Name,
			Namespace:   sr.Namespace,
			Labels:      Labels(sr),
			Annotations: Annotations(),
			OwnerReferences: []metav1.OwnerReference{
				{
					Name:       sr.Name,
					Kind:       sr.Kind,
					APIVersion: sr.APIVersion,
					UID:        sr.UID,
				},
			},
		}
	case *squidv1.SquidConfigs:
		return &metav1.ObjectMeta{
			Name:        sr.Name,
			Namespace:   sr.Namespace,
			Labels:      Labels(sr),
			Annotations: Annotations(),
			OwnerReferences: []metav1.OwnerReference{
				{
					Name:       sr.Name,
					Kind:       sr.Kind,
					APIVersion: sr.APIVersion,
					UID:        sr.UID,
				},
			},
		}
	}

	return nil
}
