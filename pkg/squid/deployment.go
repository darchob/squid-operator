package squid

import (
	"fmt"
	"time"

	squidv1 "git.fr.clara.net/claranet/healthcare/buildops/projects/kubernetes/operators/squid-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func UpdatedDeployment(current *appsv1.Deployment) error {
	_, ok := current.Annotations["date"]
	if !ok {
		return fmt.Errorf("No way to perform RollingUpdate the date Label is not found")
	}
	current.Annotations["date"] = time.DateTime
	return nil
}

func NewDeployment(sr *squidv1.Configs) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: ObjectMeta(sr),
		Spec: appsv1.DeploymentSpec{
			Replicas: sr.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: Labels(sr),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: ObjectMeta(sr),
				Spec: corev1.PodSpec{
					ServiceAccountName: sr.Name,
					Volumes: []corev1.Volume{
						{
							Name: fmt.Sprintf("%s-config", sr.Name),
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: sr.Name,
									},
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:            sr.Name,
							Image:           fmt.Sprintf("%s:%s", sr.Spec.Image.Repository, sr.Spec.Image.Tags),
							ImagePullPolicy: corev1.PullAlways,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      fmt.Sprintf("%s-config", sr.Name),
									MountPath: "/etc/squid/squid.conf",
									SubPath:   "squid.conf",
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 3128,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									TCPSocket: &corev1.TCPSocketAction{
										Port: intstr.IntOrString{
											Type:   intstr.String,
											StrVal: "http",
										},
									},
								},
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									TCPSocket: &corev1.TCPSocketAction{
										Port: intstr.IntOrString{
											Type:   intstr.String,
											StrVal: "http",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
