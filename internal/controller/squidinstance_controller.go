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
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// SquidInstanceReconciler reconciles a SquidInstance object
type SquidInstanceReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

const (
	configMapField = "metadata.name"
)

//+kubebuilder:rbac:groups=squid.cdk.clara.net,resources=squidinstances,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=squid.cdk.clara.net,resources=squidinstances/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=squid.cdk.clara.net,resources=squidinstances/finalizers,verbs=update

//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=service,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking,resources=ingress,verbs=get;list;watch;create;update;patch;delete

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
		if err := r.deletion(ctx, &squidInstance); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	if err := r.create(ctx, &squidInstance); err != nil {
		log.Error(err, "resource creating failed")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SquidInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&squidv1.SquidInstance{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&corev1.ConfigMap{}).
		WithEventFilter(
			predicate.ResourceVersionChangedPredicate{}).
		Complete(r)
}

func (r *SquidInstanceReconciler) handlingUpdate(ctx context.Context, squidInstance *squidv1.SquidInstance) error {
	if err := r.Status().Update(ctx, squidInstance); err != nil {
		return err
	}

	return nil
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
	if err := r.serviceAccount(ctx, squidInstance); err != nil {
		squidInstance.Status.ServiceAccount = squidv1.PhaseError
		log.Error(err, "createOrUpdate", "sa", squidInstance.Name)
		return r.handlingUpdate(ctx, squidInstance)
	}

	if err := r.configMap(ctx, squidInstance); err != nil {
		squidInstance.Status.ConfigMap = squidv1.PhaseError
		log.Error(err, "createOrUpdate", "configmap", squidInstance.Name)
		return r.handlingUpdate(ctx, squidInstance)
	}

	if err := r.deployment(ctx, squidInstance); err != nil {
		squidInstance.Status.Deployment = squidv1.PhaseError
		log.Error(err, "createOrUpdate", "deployment", squidInstance.Name)
		return r.handlingUpdate(ctx, squidInstance)
	}

	if err := r.service(ctx, squidInstance); err != nil {
		squidInstance.Status.Deployment = squidv1.PhaseError
		log.Error(err, "createOrUpdate", "deployment", squidInstance.Name)
		return r.handlingUpdate(ctx, squidInstance)
	}

	if err := r.ingress(ctx, squidInstance); err != nil {
		squidInstance.Status.Deployment = squidv1.PhaseError
		log.Error(err, "createOrUpdate", "deployment", squidInstance.Name)
		return r.handlingUpdate(ctx, squidInstance)
	}

	squidInstance.Status.ServiceAccount = squidv1.PhaseDeployed
	squidInstance.Status.ConfigMap = squidv1.PhaseDeployed
	squidInstance.Status.Deployment = squidv1.PhaseDeployed

	if err := r.Status().Update(ctx, squidInstance); err != nil {
		return err
	}

	return nil
}

func (r *SquidInstanceReconciler) serviceAccount(ctx context.Context, instance *squidv1.SquidInstance) error {
	sa := corev1.ServiceAccount{}
	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: instance.Namespace, Name: instance.Name}, &sa); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		newSa := squid.NewServiceAccount(instance)
		if err = r.Create(ctx, newSa); err != nil {
			return err
		}

		if err := controllerutil.SetControllerReference(instance, newSa, r.Scheme); err != nil {
			return err
		}

		r.Recorder.Event(instance, "Normal", "Deployed", fmt.Sprintf("ServiceAccount %s has been deployed", sa.Name))
		return nil
	}

	return nil
}

func (r *SquidInstanceReconciler) deployment(ctx context.Context, instance *squidv1.SquidInstance) error {
	log := log.Log.WithName("squidInstance")

	deployment := appsv1.Deployment{}

	if err := r.Get(ctx, client.ObjectKey{Namespace: instance.Namespace, Name: instance.Name}, &deployment); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		newDeploy := squid.NewDeployment(instance)
		if err = r.Create(ctx, newDeploy); err != nil {
			log.Error(err, "unable to create new deployment")
			return err
		}

		if err := controllerutil.SetControllerReference(instance, newDeploy, r.Scheme); err != nil {
			return err
		}

		r.Recorder.Event(instance, "Normal", "Deployed", fmt.Sprintf("Deployment %s has been deployed", deployment.Name))
		return nil
	}

	return nil
}

func (r *SquidInstanceReconciler) configMap(ctx context.Context, instance *squidv1.SquidInstance) error {
	log := log.Log.WithName("squidInstance")

	configmap := corev1.ConfigMap{}

	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: instance.Namespace, Name: instance.Name}, &configmap); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		newConfigMap := squid.ConfigMap(instance)
		if err = r.Create(ctx, newConfigMap); err != nil {
			return err
		}

		if err := controllerutil.SetControllerReference(instance, newConfigMap, r.Scheme); err != nil {
			return err
		}

		r.Recorder.Event(instance, "Normal", "Deployed", fmt.Sprintf("Secondary ConfigMap %s has been deployed", configmap.Name))
	}

	log.Info("update", "configmap", configmap.Name)

	// This ConfigMap couldn't be updated see func : SetupWithManager(mgr ctrl.Manager)
	return nil
}

func (r *SquidInstanceReconciler) service(ctx context.Context, instance *squidv1.SquidInstance) error {
	svc := corev1.Service{}
	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: instance.Namespace, Name: instance.Name}, &svc); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		newSvc := squid.Service(instance)
		if err = r.Create(ctx, newSvc); err != nil {
			return err
		}

		if err := controllerutil.SetControllerReference(instance, newSvc, r.Scheme); err != nil {
			return err
		}

		r.Recorder.Event(instance, "Normal", "Deployed", fmt.Sprintf("Service %s has been deployed", svc.Name))
		return nil
	}

	return nil
}

func (r *SquidInstanceReconciler) ingress(ctx context.Context, instance *squidv1.SquidInstance) error {
	ing := networkingv1.Ingress{}
	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: instance.Namespace, Name: instance.Name}, &ing); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		newIng := squid.Ingress(instance)
		if err = r.Create(ctx, newIng); err != nil {
			return err
		}

		if err := controllerutil.SetControllerReference(instance, newIng, r.Scheme); err != nil {
			return err
		}

		r.Recorder.Event(instance, "Normal", "Deployed", fmt.Sprintf("Ingress %s has been deployed", ing.Name))
		return nil
	}

	return nil
}
