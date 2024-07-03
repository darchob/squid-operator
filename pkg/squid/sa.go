package squid

import (
	squidv1 "git.fr.clara.net/claranet/healthcare/buildops/projects/kubernetes/operators/squid-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
)

func NewServiceAccount(sr *squidv1.SquidInstance) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: ObjectMeta(sr),
		// ImagePullSecrets: ,
	}
}
