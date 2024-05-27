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

package controller

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	squidv1 "git.fr.clara.net/claranet/healthcare/buildops/projects/kubernetes/operators/squid-operator/api/v1"
)

// RulesReconciler reconciles a Rules object
type RulesReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=squid.cdk.clara.net,resources=rules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=squid.cdk.clara.net,resources=rules/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=squid.cdk.clara.net,resources=rules/finalizers,verbs=update

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *RulesReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	rules := &squidv1.Rules{}
	rules.Status.Merged = false

	if err := r.Get(ctx, req.NamespacedName, rules); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Ensure Finalizer
	if !controllerutil.ContainsFinalizer(rules, finalizerName) {
		rules.ObjectMeta.Finalizers = append(rules.ObjectMeta.Finalizers, finalizerName)
		if err := r.Update(ctx, rules); err != nil {
			return ctrl.Result{}, err
		}
		return r.handlingStatusRules(ctx, rules)
	}

	// Delete deployment and ConfigMap
	if !rules.ObjectMeta.DeletionTimestamp.IsZero() {
		rules.ObjectMeta.Finalizers = []string{}

		if err := r.CleanRules(ctx, rules); err != nil {
			r.Recorder.Event(rules, "Warning", "Not Merged", fmt.Sprintf("Rule %s has been not cleaned with error %s", rules.Name, err))
			return r.handlingStatusRules(ctx, rules)
		}

		if err := r.Update(ctx, rules); err != nil {
			return r.handlingStatusRules(ctx, rules)
		}

		return ctrl.Result{}, nil
	}

	if err := r.EnsureRules(ctx, rules); err != nil {
		r.Recorder.Event(rules, "Warning", "Not Merged", fmt.Sprintf("Rule %s has been not merged with error %s", rules.Name, err))
		return r.handlingStatusRules(ctx, rules)
	}

	rules.Status.Merged = true
	r.Recorder.Event(rules, "Normal", "Merged", fmt.Sprintf("Rule %s has been merged for SquidConfigs %s", rules.Name, rules.Spec.SquidConfig.Name))

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RulesReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&squidv1.Rules{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}
