package squid

import (
	squidv1 "git.fr.clara.net/claranet/healthcare/buildops/projects/kubernetes/operators/squid-operator/api/v1"
	corev1 "k8s.io/api/core/v1"

	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func Service(sr *squidv1.SquidInstance) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: ObjectMeta(sr),
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Protocol:   corev1.ProtocolTCP,
					Port:       3128,
					TargetPort: intstr.FromInt32(3128),
				},
			},
			Selector: map[string]string{
				"app.kubernetes.io/name": sr.Name,
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}

func Ingress(sr *squidv1.SquidInstance) *networkingv1.Ingress {
	return &networkingv1.Ingress{
		ObjectMeta: ObjectMeta(sr),
		Spec:       sr.Spec.IngressSpec,
	}
}
