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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	squidv1 "git.fr.clara.net/claranet/healthcare/buildops/projects/kubernetes/operators/squid-operator/api/v1"
)

// ConfigsReconciler reconciles a Configs object
type ConfigsReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const (
	finalizerName = "squid.ckd.clara.net/finalizer"
)

//+kubebuilder:rbac:groups=squid.cdk.clara.net,resources=configs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=squid.cdk.clara.net,resources=configs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=squid.cdk.clara.net,resources=configs/finalizers,verbs=update

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *ConfigsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	configs := &squidv1.Configs{}

	log.Info("Squid Configs reconciling ", "name", req.NamespacedName.Name)
	if err := r.Get(ctx, req.NamespacedName, configs); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Delete deployment and ConfigMap
	if !configs.ObjectMeta.DeletionTimestamp.IsZero() {
		//TODO !!! Delete Squid Deployment
		return ctrl.Result{}, nil
	}

	// Ensure Finalizer
	if !controllerutil.ContainsFinalizer(configs, finalizerName) {
		configs.ObjectMeta.Finalizers = append(configs.ObjectMeta.Finalizers, finalizerName)
		if err := r.Update(ctx, configs); err != nil {
			return ctrl.Result{}, err
		}
		return r.handlingConfigUpdate(ctx, configs)
	}

	configs.Status.ServiceAccount = squidv1.PhasePending
	configs.Status.ConfigMap = squidv1.PhasePending
	configs.Status.Deployment = squidv1.PhasePending

	log.Info("Check squid service account", "name", req.NamespacedName.Name)
	if err := r.ensureServiceAccount(ctx, configs); err != nil {
		configs.Status.ServiceAccount = squidv1.PhaseError
		log.Error(err, "On ServiceAccount create task", "name", req.NamespacedName.Name)
		return r.handlingConfigUpdate(ctx, configs)
	}

	if err := r.ensureDefaultConfigMap(ctx, configs); err != nil {
		configs.Status.ConfigMap = squidv1.PhaseError
		return r.handlingConfigUpdate(ctx, configs)
	}

	if err := r.ensureDeployment(ctx, configs); err != nil {
		configs.Status.Deployment = squidv1.PhaseError
		return r.handlingConfigUpdate(ctx, configs)
	}

	configs.Status.ServiceAccount = squidv1.PhaseDeployed
	configs.Status.ConfigMap = squidv1.PhaseDeployed
	configs.Status.Deployment = squidv1.PhaseDeployed

	return r.handlingConfigUpdate(ctx, configs)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConfigsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&squidv1.Configs{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.ServiceAccount{}).
		Complete(r)
}
