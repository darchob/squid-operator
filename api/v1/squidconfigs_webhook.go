/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"context"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var squidconfigslog = logf.Log.WithName("squidconfigs-resource")

// SetupWebhookWithManager will setup the manager to manage the webhooks
func (r *SquidConfigs) SetupWebhookWithManager(mgr ctrl.Manager) error {
	r.client = mgr.GetClient()
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-squid-cdk-clara-net-v1-squidconfigs,mutating=true,failurePolicy=fail,sideEffects=None,groups=squid.cdk.clara.net,resources=squidconfigs,verbs=create;update,versions=v1,name=msquidconfigs.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &SquidConfigs{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *SquidConfigs) Default() {
	squidconfigslog.Info("default", "name", r.Name)
	// Nothing todo here
}

//+kubebuilder:webhook:path=/validate-squid-cdk-clara-net-v1-squidconfigs,mutating=false,failurePolicy=fail,sideEffects=None,groups=squid.cdk.clara.net,resources=squidconfigs,verbs=create;update,versions=v1,name=vsquidconfigs.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &SquidConfigs{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *SquidConfigs) ValidateCreate() (admission.Warnings, error) {
	squidconfigslog.Info("validate create", "name", r.Name)
	ctx := context.Background()

	ctxTimeout, cancel := context.WithTimeout(ctx, time.Minute*2)
	defer cancel()

	if err := r.validate(ctxTimeout); err != nil {
		return admission.Warnings{"cloud not validate squid config"}, err
	}

	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *SquidConfigs) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	squidconfigslog.Info("validate update", "name", r.Name)
	ctx := context.Background()
	ctxTimeout, cancel := context.WithTimeout(ctx, time.Minute*2)
	defer cancel()

	if err := r.validate(ctxTimeout); err != nil {
		return admission.Warnings{"cloud not validate squid config"}, err
	}

	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *SquidConfigs) ValidateDelete() (admission.Warnings, error) {
	squidconfigslog.Info("validate delete", "name", r.Name)

	configmap := &corev1.ConfigMap{}
	ctx := context.Background()
	dataKey := fmt.Sprintf("%s.conf", r.Name)
	configMapName := r.GetAnnotations()["squid.ckd.clara.net/instance"]

	if err := r.client.Get(ctx, client.ObjectKey{Namespace: r.Namespace, Name: configMapName}, configmap); err != nil {
		return admission.Warnings{"could not delete configs"}, err
	}

	if _, ok := configmap.Data[dataKey]; ok {
		return admission.Warnings{"could not delete configs"}, fmt.Errorf("configmap %s still contains the following rules %s ", configMapName, dataKey)
	}

	return nil, nil
}

func (r *SquidConfigs) validate(ctx context.Context) error {
	squidInstance := &SquidInstance{}
	instanceName, ok := r.GetAnnotations()["squid.ckd.clara.net/instance"]
	if !ok {
		return fmt.Errorf("missing annotation squid.ckd.clara.net/instance")
	}

	if err := r.client.Get(ctx, client.ObjectKey{Namespace: r.Namespace, Name: instanceName}, squidInstance); err != nil {
		return err
	}
	job := r.webhookJob(squidInstance.Spec.Image.Repository, squidInstance.Spec.Image.Tag)
	if err := r.client.Create(ctx, job); err != nil {
		return err
	}

	if err := r.waitForJobCompletion(ctx, job.Name); err != nil {
		return err
	}

	return nil
}

func (r *SquidConfigs) webhookJob(imageName, imageTag string) *batchv1.Job {
	spec := r.Spec.DeepCopy()
	copyCommand := fmt.Sprintf("echo '%s' > /etc/squid/conf.d/00-squid.conf && squid -k parse -f /etc/squid/conf.d/00-squid.conf", spec.Rules)

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("validate-%s-", r.Name),
			Namespace:    r.Namespace,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:            "validator",
							Image:           fmt.Sprintf("%s:%s", imageName, imageTag),
							ImagePullPolicy: corev1.PullAlways,
							Command:         []string{"/bin/sh", "-c", copyCommand},
						},
					},
				},
			},
		},
	}
}

func (r *SquidConfigs) waitForJobCompletion(ctx context.Context, jobName string) error {
	var job batchv1.Job

	// Utilisation de PollUntilContextTimeout au lieu de PollImmediate
	err := wait.PollUntilContextCancel(
		ctx,
		time.Second*5,
		true,
		func(ctx context.Context) (bool, error) {
			if err := r.client.Get(ctx, types.NamespacedName{
				Namespace: r.Namespace,
				Name:      jobName,
			}, &job); err != nil {
				if client.IgnoreNotFound(err) == nil {
					return false, err
				}
			}

			if job.Status.Succeeded > 0 {
				return true, nil
			}
			if job.Status.Failed > 0 {
				return false, fmt.Errorf("validation job failed")
			}
			return false, nil
		})

	return err
}
