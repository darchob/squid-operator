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
	"time"

	squidv1 "git.fr.clara.net/claranet/healthcare/buildops/projects/kubernetes/operators/squid-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// ConfigsReconciler reconciles a Configs object
type ConfigsReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

const (
	finalizerName = "squid.ckd.clara.net/finalizer"
	ruleField     = ".spec.squidConfig.name"
)

//+kubebuilder:rbac:groups=squid.cdk.clara.net,resources=configs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=squid.cdk.clara.net,resources=configs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=squid.cdk.clara.net,resources=configs/finalizers,verbs=update
//+kubebuilder:rbac:groups=squid.cdk.clara.net,resources=rules/status,verbs=get
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;update;patch;delete

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *ConfigsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	configs := &squidv1.Configs{}
	rules := &squidv1.Rules{}

	if err := r.Get(ctx, req.NamespacedName, configs); err != nil {
		if err := r.Get(ctx, req.NamespacedName, rules); err != nil {
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}

		if err := r.Client.Get(ctx, client.ObjectKey{Namespace: rules.Namespace, Name: rules.Spec.SquidConfig.Name}, configs); err != nil {
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}

		configs.Status.Deployment = squidv1.PhasePending
		if err := r.Update(ctx, configs); err != nil {
			return ctrl.Result{}, err
		}

		if err := r.rollingUpdateDeployment(ctx, rules); err != nil {
			r.Recorder.Event(configs, "Warning", "not merged", fmt.Sprintf("Rules with name %s couldn't be merged", rules.Name))
			configs.Status.ConfigMap = squidv1.PhaseError
			configs.Status.Deployment = squidv1.PhaseError
			return r.handlingConfigUpdate(ctx, configs)
		}

		r.Recorder.Event(configs, "Normal", "merged", fmt.Sprintf("Rules with name %s is merged and applied", rules.Name))
		return ctrl.Result{}, err
	}

	// Ensure Finalizer
	if !controllerutil.ContainsFinalizer(configs, finalizerName) {
		configs.ObjectMeta.Finalizers = append(configs.ObjectMeta.Finalizers, finalizerName)
		if err := r.Update(ctx, configs); err != nil {
			return ctrl.Result{}, err
		}
		return r.handlingConfigUpdate(ctx, configs)
	}

	// Delete deployment and ConfigMap
	if !configs.ObjectMeta.DeletionTimestamp.IsZero() {
		configs.ObjectMeta.Finalizers = []string{}
		if err := r.Update(ctx, configs); err != nil {
			return r.handlingConfigUpdate(ctx, configs)
		}

		return ctrl.Result{}, nil
	}

	configs.Status.ServiceAccount = squidv1.PhasePending
	configs.Status.ConfigMap = squidv1.PhasePending
	configs.Status.Deployment = squidv1.PhasePending

	if err := r.ensureServiceAccount(ctx, configs); err != nil {
		configs.Status.ServiceAccount = squidv1.PhaseError
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

func (r *ConfigsReconciler) findObjectsByNamespace(ctx context.Context, rules client.Object) []reconcile.Request {
	_ = log.FromContext(ctx)

	attachedConfigDeployments := &squidv1.RulesList{}

	listOps := &client.ListOptions{
		Namespace: rules.GetNamespace(),
	}

	err := r.List(ctx, attachedConfigDeployments, listOps)
	if err != nil {
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, len(attachedConfigDeployments.Items))
	for i, item := range attachedConfigDeployments.Items {
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
			},
		}
	}
	return requests
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConfigsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	log := log.Log.WithName("setup manager")

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &squidv1.Rules{}, ruleField, func(rawObj client.Object) []string {
		configRule := rawObj.(*squidv1.Rules)

		if configRule.Spec.SquidConfig.Name == "" {
			return nil
		}

		return []string{configRule.Spec.SquidConfig.Name}
	}); err != nil {
		log.Error(err, "Indexing error")
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&squidv1.Configs{}).
		Owns(&corev1.ConfigMap{}, builder.WithPredicates(predicate.Funcs{
			UpdateFunc: func(ue event.UpdateEvent) bool {
				ue.ObjectNew.SetAnnotations(map[string]string{
					"squid-operator.kubernetes.io/updatedAt": time.Now().Format(time.RFC3339),
				})
				return false
			},
		})).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.ServiceAccount{}).
		Watches(
			&squidv1.Rules{},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsByNamespace),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Complete(r)
}
