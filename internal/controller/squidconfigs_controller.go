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
	"time"

	squidv1 "git.fr.clara.net/claranet/healthcare/buildops/projects/kubernetes/operators/squid-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// SquidConfigsReconciler reconciles a SquidConfigs object
type SquidConfigsReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

var (
	duplicateError = "dupplicate key found"
	notFoundError  = "key not found"
	emptyError     = "configmap is empty"
)

const (
	finalizerName = "squid.ckd.clara.net/finalizer"
	configField   = "metadata.annotations[\"squid.ckd.clara.net/instance\"]"
)

//+kubebuilder:rbac:groups=squid.cdk.clara.net,resources=squidconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=squid.cdk.clara.net,resources=squidconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=squid.cdk.clara.net,resources=squidconfigs/finalizers,verbs=update

//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployment,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
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

	if !controllerutil.ContainsFinalizer(&squidConfig, finalizerName) {
		if err := r.init(ctx, &squidConfig); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{Requeue: true}, nil
	}

	if !squidConfig.ObjectMeta.DeletionTimestamp.IsZero() {
		log.Info("resource deleting", "resource", req.NamespacedName)
		if err := r.deletion(ctx, &squidConfig); err != nil {
			if errors.IsConflict(err) {
				return r.handlingRequeuUpdate(ctx, &squidConfig)
			}

			if isEmptyErr(err) {
				return ctrl.Result{}, nil
			}

			if isNotFoundErr(err) {
				return ctrl.Result{}, nil
			}

			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	if err := r.configMapUpgrade(ctx, &squidConfig); err != nil {
		squidConfig.Status.Apply = false
		if errors.IsConflict(err) {
			return r.handlingRequeuUpdate(ctx, &squidConfig)
		}

		if !isDuplicateErr(err) {
			return r.handlingUpdate(ctx, &squidConfig)
		}
	}

	squidConfig.Status.Apply = true
	return r.handlingUpdate(ctx, &squidConfig)
}

// SetupWithManager sets up the controller with the Manager.
func (r *SquidConfigsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&squidv1.SquidConfigs{}).
		// WithEventFilter(predicate.ResourceVersionChangedPredicate{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}

func (r *SquidConfigsReconciler) handlingRequeuUpdate(ctx context.Context, configs *squidv1.SquidConfigs) (ctrl.Result, error) {
	if err := r.Status().Update(ctx, configs); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{Requeue: true}, nil
}

func (r *SquidConfigsReconciler) handlingUpdate(ctx context.Context, configs *squidv1.SquidConfigs) (ctrl.Result, error) {
	if err := r.Status().Update(ctx, configs); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{Requeue: false}, nil
}

func (r *SquidConfigsReconciler) configMapUpgrade(ctx context.Context, configs *squidv1.SquidConfigs) error {
	_ = log.Log.WithName("squidConfigs")

	objectsList := []client.Object{
		&appsv1.Deployment{},
		&corev1.ConfigMap{},
	}

	configMapName := configs.GetAnnotations()["squid.ckd.clara.net/instance"]
	for _, object := range objectsList {
		if err := r.Get(ctx, client.ObjectKey{Namespace: configs.Namespace, Name: configMapName}, object); err != nil {
			return err
		}

		switch obj := object.(type) {
		case *corev1.ConfigMap:
			if obj.Data == nil {
				obj.Data = make(map[string]string)
			}

			dataKey := fmt.Sprintf("%s.conf", configs.Name)
			_, ok := obj.Data[dataKey]
			if ok {
				return fmt.Errorf(duplicateError)
			}

			obj.Data[dataKey] = configs.Spec.Rules

		case *appsv1.Deployment:
			obj.Spec.Template.ObjectMeta.Annotations["squid-operator.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)
		}

		if err := r.Update(ctx, object); err != nil {
			return err
		}

		r.Recorder.Event(configs, "Normal", "Updated", fmt.Sprintf("%s %s has been updated", reflect.TypeOf(object).String(), object.GetName()))
	}

	return nil
}

func (r *SquidConfigsReconciler) init(ctx context.Context, squidConfigs *squidv1.SquidConfigs) error {
	log := log.Log.WithName("squidConfigs")

	log.Info("Reconcile", "init", squidConfigs.Name)

	squidConfigs.ObjectMeta.Finalizers = append(squidConfigs.ObjectMeta.Finalizers, finalizerName)
	if err := r.Update(ctx, squidConfigs); err != nil {
		log.Error(err, "init", "finalizer", squidConfigs.Name)
		return err
	}

	return nil
}

func (r *SquidConfigsReconciler) deletion(ctx context.Context, configs *squidv1.SquidConfigs) error {
	log := log.Log.WithName("squidConfigs")

	objectsList := []client.Object{
		&appsv1.Deployment{},
		&corev1.ConfigMap{},
	}

	configMapName := configs.GetAnnotations()["squid.ckd.clara.net/instance"]
	for _, object := range objectsList {
		if err := r.Get(ctx, client.ObjectKey{Namespace: configs.Namespace, Name: configMapName}, object); err != nil {
			return err
		}

		switch obj := object.(type) {
		case *corev1.ConfigMap:
			if obj.Data == nil {
				err := fmt.Errorf(emptyError)
				log.Error(err, "deletion", configs.Name)
				return err
			}

			dataKey := fmt.Sprintf("%s.conf", configs.Name)
			_, ok := obj.Data[dataKey]
			if !ok {
				err := fmt.Errorf(notFoundError)
				log.Error(err, "error on delete")
				return err
			}

			delete(obj.Data, dataKey)
		case *appsv1.Deployment:
			obj.Spec.Template.ObjectMeta.Annotations["squid-operator.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)
		}

		if err := r.Update(ctx, object); err != nil {
			return err
		}
	}

	configs.ObjectMeta.Finalizers = []string{}
	if err := r.Update(ctx, configs); err != nil {
		log.Error(err, "init", "finalizer", configs.Name)
		return err
	}

	return nil
}

func isDuplicateErr(err error) bool {
	return err.Error() == duplicateError
}

func isNotFoundErr(err error) bool {
	return err.Error() == notFoundError
}

func isEmptyErr(err error) bool {
	return err.Error() == emptyError
}
