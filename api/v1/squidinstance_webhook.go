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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var squidinstancelog = logf.Log.WithName("squidinstance-resource")

// SetupWebhookWithManager will setup the manager to manage the webhooks
func (r *SquidInstance) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

const (
	DefaultImage        = "ubuntu/squid"
	DefaultTag          = "latest"
	DefaultStorageClass = "standard"
)

//+kubebuilder:webhook:path=/mutate-squid-cdk-clara-net-v1-squidinstance,mutating=true,failurePolicy=fail,sideEffects=None,groups=squid.cdk.clara.net,resources=squidinstances,verbs=create;update,versions=v1,name=msquidinstance.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &SquidInstance{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *SquidInstance) Default() {
	squidinstancelog.Info("default", "name", r.Name)
	if r.Spec.Image == nil {
		r.Spec.Image.Repository = DefaultImage
		r.Spec.Image.Tag = DefaultTag
	}

	if r.Spec.StorageClassName == "" {
		r.Spec.StorageClassName = DefaultStorageClass
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
//+kubebuilder:webhook:path=/validate-squid-cdk-clara-net-v1-squidinstance,mutating=false,failurePolicy=fail,sideEffects=None,groups=squid.cdk.clara.net,resources=squidinstances,verbs=create;update,versions=v1,name=vsquidinstance.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &SquidInstance{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *SquidInstance) ValidateCreate() (admission.Warnings, error) {
	squidinstancelog.Info("validate create", "name", r.Name)
	// TODO(user): fill in your validation logic upon object creation.
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *SquidInstance) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	squidinstancelog.Info("validate update", "name", r.Name)
	// TODO(user): fill in your validation logic upon object update.
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *SquidInstance) ValidateDelete() (admission.Warnings, error) {
	squidinstancelog.Info("validate delete", "name", r.Name)
	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
