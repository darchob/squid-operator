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
	"reflect"

	squidv1 "git.fr.clara.net/claranet/healthcare/buildops/projects/kubernetes/operators/squid-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type state int

const (
	merged state = iota
	deletion
	applying
	failed
)

var states = map[state]string{
	merged:   "Merged",
	deletion: "Deletion",
	applying: "Applying",
	failed:   "Failed",
}

// SquidConfigsReconciler reconciles a SquidConfigs object
type SquidConfigsReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=squid.cdk.clara.net,resources=squidconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=squid.cdk.clara.net,resources=squidconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=squid.cdk.clara.net,resources=squidconfigs/finalizers,verbs=update

//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployment,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=persistentvolumeclaims ,verbs=get;list;watch;create;update;patch;delete

//+kubebuilder:rbac:groups="networking.k8s.io",resources=ingresses,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *SquidConfigsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.Log.WithName("squidConfigs")

	log.Info("Reconcile", "request", req.NamespacedName)

	var squidConfig squidv1.SquidConfigs
	if err := r.Get(ctx, req.NamespacedName, &squidConfig); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	switch {
	case !controllerutil.ContainsFinalizer(&squidConfig, finalizerName):
		return r.appendFinalizer(ctx, &squidConfig)
	default:
		return r.handlingReconciliation(ctx, &squidConfig)
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *SquidConfigsReconciler) SetupWithManager(mgr ctrl.Manager) error {

	return ctrl.NewControllerManagedBy(mgr).
		For(&squidv1.SquidConfigs{}).
		// WithEventFilter(predicate.ResourceVersionChangedPredicate{}).
		// WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}

func (r *SquidConfigsReconciler) appendFinalizer(ctx context.Context, configs *squidv1.SquidConfigs) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithName(configs.Name)
	log.Info("initialize", "squidConfig", configs.Name)

	configs.ObjectMeta.Finalizers = append(configs.ObjectMeta.Finalizers, finalizerName)

	if err := r.Update(ctx, configs); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{Requeue: false}, nil
}

func (r *SquidConfigsReconciler) handlingReconciliation(ctx context.Context, configs *squidv1.SquidConfigs) (ctrl.Result, error) {
	_ = log.FromContext(ctx).WithName(configs.Name)

	objectsList := []client.Object{
		// &appsv1.Deployment{},
		&corev1.ConfigMap{},
	}

	configMapName := configs.GetAnnotations()["squid.ckd.clara.net/instance"]
	for _, object := range objectsList {
		if err := r.Get(ctx, client.ObjectKey{Namespace: configs.Namespace, Name: configMapName}, object); err != nil {
			return ctrl.Result{}, err
		}

		truncate := false
		switch obj := object.(type) {
		case *corev1.ConfigMap:
			if !configs.ObjectMeta.DeletionTimestamp.IsZero() {
				truncate = true
			}

			configmap, err := handleConfigMap(obj, configs, truncate)
			if err != nil {
				return r.handlingUpdate(ctx, configs, err, failed)
			}

			if err := r.Update(ctx, configmap); err != nil {
				return r.handlingUpdate(ctx, configs, err, failed)
			}
			// case *appsv1.Deployment:
			// 	deployment := obj.DeepCopy()
			// 	deployment.Spec.Template.ObjectMeta.Annotations["squid-operator.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)
			// 	if err := r.Update(ctx, deployment); err != nil {
			// 		return r.handlingUpdate(ctx, configs, err, failed)
			// 	}
		}

		r.Recorder.Event(configs, "Normal", "Updated", fmt.Sprintf("%s %s has been updated", reflect.TypeOf(object).String(), object.GetName()))
	}

	return r.handlingUpdate(ctx, configs, nil, merged)
}

func (r *SquidConfigsReconciler) handlingUpdate(ctx context.Context, configs *squidv1.SquidConfigs, err error, state state) (ctrl.Result, error) {
	config := configs.DeepCopy()
	config.Status.State = states[state]

	if !isDuplicateErr(err) {
		config.Status.State = states[failed]
	}

	if !config.ObjectMeta.DeletionTimestamp.IsZero() {
		config.Status.State = states[deletion]
		config.ObjectMeta.Finalizers = []string{}
	}

	if err := r.Update(ctx, config); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{Requeue: true}, nil
}
