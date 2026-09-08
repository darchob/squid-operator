/*
Copyright 2025.

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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	squidv1 "git.fr.clara.net/claranet/healthcare/buildops/projects/kubernetes/operators/squid-operator/api/v1"
)

// nolint:unused
// log is for logging in this package.
var squidinstancelog = logf.Log.WithName("squidinstance-resource")

// SetupSquidInstanceWebhookWithManager registers the webhook for SquidInstance in the manager.
func SetupSquidInstanceWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&squidv1.SquidInstance{}).
		WithValidator(&SquidInstanceCustomValidator{}).
		WithDefaulter(&SquidInstanceCustomDefaulter{}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-squid-cdk-clara-net-v1-squidinstance,mutating=true,failurePolicy=fail,sideEffects=None,groups=squid.cdk.clara.net,resources=squidinstances,verbs=create;update,versions=v1,name=msquidinstance-v1.kb.io,admissionReviewVersions=v1

// SquidInstanceCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind SquidInstance when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type SquidInstanceCustomDefaulter struct {
	// TODO(user): Add more fields as needed for defaulting
}

var _ webhook.CustomDefaulter = &SquidInstanceCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind SquidInstance.
func (d *SquidInstanceCustomDefaulter) Default(_ context.Context, obj runtime.Object) error {
	squidinstance, ok := obj.(*squidv1.SquidInstance)

	if !ok {
		return fmt.Errorf("expected an SquidInstance object but got %T", obj)
	}
	squidinstancelog.Info("Defaulting for SquidInstance", "name", squidinstance.GetName())

	// TODO(user): fill in your defaulting logic.

	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-squid-cdk-clara-net-v1-squidinstance,mutating=false,failurePolicy=fail,sideEffects=None,groups=squid.cdk.clara.net,resources=squidinstances,verbs=create;update,versions=v1,name=vsquidinstance-v1.kb.io,admissionReviewVersions=v1

// SquidInstanceCustomValidator struct is responsible for validating the SquidInstance resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type SquidInstanceCustomValidator struct {
	// TODO(user): Add more fields as needed for validation
}

var _ webhook.CustomValidator = &SquidInstanceCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type SquidInstance.
func (v *SquidInstanceCustomValidator) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	squidinstance, ok := obj.(*squidv1.SquidInstance)
	if !ok {
		return nil, fmt.Errorf("expected a SquidInstance object but got %T", obj)
	}
	squidinstancelog.Info("Validation for SquidInstance upon creation", "name", squidinstance.GetName())

	// TODO(user): fill in your validation logic upon object creation.

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type SquidInstance.
func (v *SquidInstanceCustomValidator) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	squidinstance, ok := newObj.(*squidv1.SquidInstance)
	if !ok {
		return nil, fmt.Errorf("expected a SquidInstance object for the newObj but got %T", newObj)
	}
	squidinstancelog.Info("Validation for SquidInstance upon update", "name", squidinstance.GetName())

	// TODO(user): fill in your validation logic upon object update.

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type SquidInstance.
func (v *SquidInstanceCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	squidinstance, ok := obj.(*squidv1.SquidInstance)
	if !ok {
		return nil, fmt.Errorf("expected a SquidInstance object but got %T", obj)
	}
	squidinstancelog.Info("Validation for SquidInstance upon deletion", "name", squidinstance.GetName())

	// TODO(user): fill in your validation logic upon object deletion.

	return nil, nil
}
