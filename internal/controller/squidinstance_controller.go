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
	squid "git.fr.clara.net/claranet/healthcare/buildops/projects/kubernetes/operators/squid-operator/pkg/squid"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// SquidInstanceReconciler reconciles a SquidInstance object
type SquidInstanceReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=squid.cdk.clara.net,resources=squidinstances,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=squid.cdk.clara.net,resources=squidinstances/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=squid.cdk.clara.net,resources=squidinstances/finalizers,verbs=update

//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete

//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *SquidInstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.Log.WithName("squidInstance")

	log.Info("Reconcile", "request", req.NamespacedName)

	var squidInstance squidv1.SquidInstance
	if err := r.Get(ctx, req.NamespacedName, &squidInstance); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !controllerutil.ContainsFinalizer(&squidInstance, finalizerName) {
		if err := r.init(ctx, &squidInstance); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{Requeue: true}, nil
	}

	if !squidInstance.ObjectMeta.DeletionTimestamp.IsZero() {
		log.Info("resource deleting", "resource", req.NamespacedName)
		if err := r.deletion(ctx, &squidInstance); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	if err := r.create(ctx, &squidInstance); err != nil {
		log.Error(err, "resource creating failed")
		squidInstance.Status.Health = squidv1.PhaseError
		return r.handlingUpdate(ctx, &squidInstance)
	}

	squidInstance.Status.Health = squidv1.PhaseDone
	return r.handlingUpdate(ctx, &squidInstance)
}

// SetupWithManager sets up the controller with the Manager.
func (r *SquidInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&squidv1.SquidInstance{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&networkingv1.Ingress{}).
		// WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}

func (r *SquidInstanceReconciler) handlingUpdate(ctx context.Context, squidInstance *squidv1.SquidInstance) (ctrl.Result, error) {
	if err := r.Status().Update(ctx, squidInstance); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{Requeue: false}, nil
}

func (r *SquidInstanceReconciler) init(ctx context.Context, squidInstance *squidv1.SquidInstance) error {
	log := log.Log.WithName("squidInstance")

	log.Info("Reconcile", "init", squidInstance.Name)

	squidInstance.ObjectMeta.Finalizers = append(squidInstance.ObjectMeta.Finalizers, finalizerName)
	if err := r.Update(ctx, squidInstance); err != nil {
		log.Error(err, "init", "finalizer", squidInstance.Name)
		return err
	}

	return nil
}

func (r *SquidInstanceReconciler) deletion(ctx context.Context, squidInstance *squidv1.SquidInstance) error {
	log := log.Log.WithName("squidInstance")

	log.Info("Reconcile", "deletion", squidInstance.Name)

	objectsList := []client.Object{
		&appsv1.Deployment{},
		&corev1.ConfigMap{},
		&corev1.ServiceAccount{},
		&corev1.Service{},
		&networkingv1.Ingress{},
	}

	for _, object := range objectsList {
		if err := r.Get(ctx, client.ObjectKey{Namespace: squidInstance.Namespace, Name: squidInstance.Name}, object); err != nil {
			return client.IgnoreNotFound(err)
		}

		log.Info("Reconcile", "deletion", object.GetName())
		if err := r.Delete(ctx, object); err != nil {
			return err
		}
	}

	squidInstance.ObjectMeta.Finalizers = []string{}
	if err := r.Update(ctx, squidInstance); err != nil {
		log.Error(err, "init", "finalizer", squidInstance.Name)
		return err
	}

	return nil
}

func (r *SquidInstanceReconciler) create(ctx context.Context, squidInstance *squidv1.SquidInstance) error {
	log := log.Log.WithName("squidInstance")
	log.Info("Reconcile", "creatOrUpdate", squidInstance.Name)

	objectsList := []client.Object{
		&appsv1.Deployment{},
		&corev1.ConfigMap{},
		&corev1.ServiceAccount{},
		&corev1.Service{},
		&networkingv1.Ingress{},
	}

	for _, object := range objectsList {
		if err := r.Get(ctx, client.ObjectKey{Namespace: squidInstance.Namespace, Name: squidInstance.Name}, object); err != nil {
			if !errors.IsNotFound(err) {
				return err
			}

			switch object.(type) {
			case *appsv1.Deployment:
				object = squid.NewDeployment(squidInstance)
			case *corev1.ConfigMap:
				object = squid.ConfigMap(squidInstance)
			case *corev1.ServiceAccount:
				object = squid.NewServiceAccount(squidInstance)
			case *corev1.Service:
				object = squid.Service(squidInstance)
			case *networkingv1.Ingress:
				object = squid.Ingress(squidInstance)
			}

			log.Info("Reconcile", "create", object.GetName())
			if err := r.Create(ctx, object); err != nil {
				log.Error(err, "createOrUpdate", reflect.TypeOf(object).String(), squidInstance.Name)
				return err
			}

			if err := controllerutil.SetControllerReference(squidInstance, object, r.Scheme); err != nil {
				return err
			}

			r.Recorder.Event(squidInstance, "Normal", "Deployed", fmt.Sprintf("%s %s has been deployed", reflect.TypeOf(object).String(), object.GetName()))
		}
	}

	return nil
}
