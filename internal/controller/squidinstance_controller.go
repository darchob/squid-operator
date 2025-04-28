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

// ResourceGenerator définit une fonction qui crée une ressource Kubernetes
type ResourceGenerator func(*squidv1.SquidInstance) client.Object

// ManagedResource définit une ressource à gérer
type ManagedResource struct {
	Type     client.Object
	Generate ResourceGenerator
}

// Définition des ressources gérées
var managedResources = []ManagedResource{
	{
		Type: &appsv1.Deployment{},
		Generate: func(si *squidv1.SquidInstance) client.Object {
			return squid.NewDeployment(si)
		},
	},
	{
		Type: &corev1.ConfigMap{},
		Generate: func(si *squidv1.SquidInstance) client.Object {
			return squid.ConfigMap(si)
		},
	},
	{
		Type: &corev1.ServiceAccount{},
		Generate: func(si *squidv1.SquidInstance) client.Object {
			return squid.NewServiceAccount(si)
		},
	},
	{
		Type: &corev1.Service{},
		Generate: func(si *squidv1.SquidInstance) client.Object {
			return squid.Service(si)
		},
	},
	{
		Type: &corev1.PersistentVolumeClaim{},
		Generate: func(si *squidv1.SquidInstance) client.Object {
			return squid.PersistentVolumeClaim(si)
		},
	},
}

// +kubebuilder:rbac:groups=squid.cdk.clara.net,resources=squidinstances,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=squid.cdk.clara.net,resources=squidinstances/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=squid.cdk.clara.net,resources=squidinstances/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;update;patch;delete
func (r *SquidInstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.Log.WithName("squidInstances")

	log.Info("Reconcile", "request", req.NamespacedName)

	var squidInstance squidv1.SquidInstance
	if err := r.Get(ctx, req.NamespacedName, &squidInstance); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	switch {
	case !controllerutil.ContainsFinalizer(&squidInstance, finalizerName):
		return r.handlingInit(ctx, &squidInstance)
	case squidInstance.DeletionTimestamp != nil:
		return r.handlingDeletion(ctx, &squidInstance)
	default:
		return r.handlingReconciliation(ctx, &squidInstance)
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *SquidInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&squidv1.SquidInstance{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Complete(r)
}

func (r *SquidInstanceReconciler) handlingInit(ctx context.Context, squidInstance *squidv1.SquidInstance) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithName(squidInstance.Name)

	log.Info("initialize", "squidInstance", squidInstance.Name)

	squidInstance.ObjectMeta.Finalizers = append(squidInstance.ObjectMeta.Finalizers, finalizerName)
	if err := r.Update(ctx, squidInstance); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{Requeue: true}, nil
}

func (r *SquidInstanceReconciler) handlingUpdate(ctx context.Context, squidInstance *squidv1.SquidInstance) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithName(squidInstance.Name)

	if err := r.Status().Update(ctx, squidInstance); err != nil {
		if errors.IsConflict(err) {
			// Conflit de mise à jour, on requeue
			log.Info("Update conflict detected, requeing", "name", squidInstance.Name)
			return ctrl.Result{Requeue: true}, nil
		}
		log.Error(err, "Failed to update status", "name", squidInstance.Name)
		return ctrl.Result{}, fmt.Errorf("failed to update status: %w", err)
	}

	return ctrl.Result{Requeue: false}, nil
}

func (r *SquidInstanceReconciler) handlingReconciliation(ctx context.Context, instance *squidv1.SquidInstance) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithName(instance.Name)
	log.Info("reconcile", "squidInstance", instance.Name)

	for _, resource := range managedResources {
		if err := r.reconcileSingleResource(ctx, instance, resource); err != nil {
			log.Error(err, "Failed to reconcile resource")
			instance.Status.Health = squidv1.PhaseError
			return r.handlingUpdate(ctx, instance)
		}
	}

	instance.Status.Health = squidv1.PhaseDone
	return r.handlingUpdate(ctx, instance)
}

func (r *SquidInstanceReconciler) reconcileSingleResource(ctx context.Context, instance *squidv1.SquidInstance, resource ManagedResource) error {
	existing := resource.Type.DeepCopyObject().(client.Object)

	err := r.Get(ctx, client.ObjectKey{
		Namespace: instance.Namespace,
		Name:      instance.Name,
	}, existing)

	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		return r.createResource(ctx, instance, resource)
	}

	return r.updateResource(ctx, instance, resource)
}

func (r *SquidInstanceReconciler) createResource(ctx context.Context, instance *squidv1.SquidInstance, resource ManagedResource) error {
	newResource := resource.Generate(instance)

	if err := controllerutil.SetControllerReference(instance, newResource, r.Scheme); err != nil {
		return fmt.Errorf("failed to set controller reference: %w", err)
	}

	if err := r.Create(ctx, newResource); err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	r.Recorder.Event(instance, "Normal", "Created",
		fmt.Sprintf("Created %T %s", newResource, newResource.GetName()))

	return nil
}

func (r *SquidInstanceReconciler) updateResource(ctx context.Context, instance *squidv1.SquidInstance, resource ManagedResource) error {
	newResource := resource.Generate(instance)

	switch obj := newResource.(type) {
	case *corev1.PersistentVolumeClaim:
		return nil
	case *corev1.ConfigMap:
		if obj.Data != nil {
			return nil
		}
	case *corev1.Service:
		if len(obj.Status.LoadBalancer.Ingress) == 0 {
			return nil
		}
		instance.Status.LoadBalancer.Ingress = obj.Status.LoadBalancer.Ingress
	}

	if err := r.Update(ctx, newResource); err != nil {
		return fmt.Errorf("failed to update resource: %w", err)
	}

	r.Recorder.Event(instance, "Normal", "Updated",
		fmt.Sprintf("Updated %T %s", newResource, newResource.GetName()))

	return nil
}

func (r *SquidInstanceReconciler) handlingDeletion(ctx context.Context, instance *squidv1.SquidInstance) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithName(instance.Name)
	log.Info("deletion", "squidInstance", instance.Name)

	for _, resource := range managedResources {
		existing := resource.Type.DeepCopyObject().(client.Object)
		err := r.Get(ctx, client.ObjectKey{
			Namespace: instance.Namespace,
			Name:      instance.Name,
		}, existing)

		if err != nil {
			if !errors.IsNotFound(err) {
				return ctrl.Result{}, err
			}
			continue
		}

		if err := r.Delete(ctx, existing); err != nil {
			return ctrl.Result{}, err
		}
	}

	controllerutil.RemoveFinalizer(instance, finalizerName)
	if err := r.Update(ctx, instance); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}
