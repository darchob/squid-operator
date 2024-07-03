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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	squidv1 "git.fr.clara.net/claranet/healthcare/buildops/projects/kubernetes/operators/squid-operator/api/v1"
)

// SquidConfigsReconciler reconciles a SquidConfigs object
type SquidConfigsReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

var (
	duplicateError = "dupplicate key found"
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
		if err := r.deletion(ctx, &squidConfig); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	if err := r.Get(ctx, req.NamespacedName, &squidConfig); err != nil {
		log.Error(err, "unable to fetch SquidConfig")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := r.handlingConfigMapUpgrade(ctx, &squidConfig); err != nil {
		if isDuplicateErr(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to update configmap")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
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

	log.Info("Reconcile", "deletion", configs.Name)
	configmap := corev1.ConfigMap{}
	deployment := appsv1.Deployment{}

	configMapName := configs.GetAnnotations()["squid.ckd.clara.net/instance"]

	if err := r.Get(ctx, client.ObjectKey{Namespace: configs.Namespace, Name: configMapName}, &configmap); err != nil {
		return client.IgnoreNotFound(err)
	}

	if err := r.Get(ctx, client.ObjectKey{Namespace: configs.Namespace, Name: configMapName}, &deployment); err != nil {
		return err
	}

	if configmap.Data == nil {
		return fmt.Errorf("empty configmap")
	}

	dataKey := fmt.Sprintf("%s.conf", configs.Name)
	_, ok := configmap.Data[dataKey]
	if !ok {
		return fmt.Errorf("key not found")
	}
	delete(configmap.Data, dataKey)

	if err := r.Update(ctx, &configmap); err != nil {
		return err
	}

	deployment.Spec.Template.ObjectMeta.Annotations["squid-operator.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)
	if err := r.Update(ctx, &deployment); err != nil {
		return err
	}

	configs.ObjectMeta.Finalizers = []string{}
	if err := r.Update(ctx, configs); err != nil {
		log.Error(err, "init", "finalizer", configs.Name)
		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SquidConfigsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	log := log.Log.WithName("squidConfigs")

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &squidv1.SquidConfigs{}, configField, func(rawObj client.Object) []string {
		// Extract the ConfigMap name from the config Spec, if one is provided
		config := rawObj.(*squidv1.SquidConfigs)
		if config.GetAnnotations()["squid.ckd.clara.net/instance"] == "" {
			return nil
		}
		return []string{config.GetAnnotations()["squid.ckd.clara.net/instance"]}
	}); err != nil {
		log.Error(err, "Indexer", "indexing configs")
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&squidv1.SquidConfigs{}).
		Watches(
			&corev1.ConfigMap{},
			handler.EnqueueRequestsFromMapFunc(r.findConfigMapByName),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Complete(r)
}

func (r *SquidConfigsReconciler) findConfigMapByName(ctx context.Context, configMap client.Object) []reconcile.Request {
	log := log.Log.WithName("SquidConfigs")

	attachedConfigsList := &squidv1.SquidConfigsList{}
	listOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(configField, configMap.GetName()),
		Namespace:     configMap.GetNamespace(),
	}

	err := r.List(ctx, attachedConfigsList, listOps)
	if err != nil {
		log.Error(err, "list")
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, len(attachedConfigsList.Items))
	for i, item := range attachedConfigsList.Items {
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
			},
		}
	}
	return requests
}

func (r *SquidConfigsReconciler) handlingConfigMapUpgrade(ctx context.Context, configs *squidv1.SquidConfigs) error {
	_ = log.FromContext(ctx)

	configmap := corev1.ConfigMap{}
	deployment := appsv1.Deployment{}

	configMapName := configs.GetAnnotations()["squid.ckd.clara.net/instance"]

	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: configs.Namespace, Name: configMapName}, &configmap); err != nil {
		return client.IgnoreNotFound(err)
	}

	if err := r.Get(ctx, client.ObjectKey{Namespace: configs.Namespace, Name: configMapName}, &deployment); err != nil {
		return err
	}

	if configmap.Data == nil {
		configmap.Data = make(map[string]string)
	}

	dataKey := fmt.Sprintf("%s.conf", configs.Name)
	_, ok := configmap.Data[dataKey]
	if ok {
		return fmt.Errorf("dupplicate key found")
	}

	configmap.Data[dataKey] = configs.Spec.Rules

	if err := r.Update(ctx, &configmap); err != nil {
		return err
	}

	// if err := controllerutil.SetControllerReference(configs, &configmap, r.Scheme); err != nil {
	// 	return err
	// }

	deployment.Spec.Template.ObjectMeta.Annotations["squid-operator.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)
	if err := r.Update(ctx, &deployment); err != nil {
		return err
	}

	configs.Status.Merged = true
	if err := r.Status().Update(ctx, configs); err != nil {
		return err
	}

	r.Recorder.Event(&configmap, "Normal", "Updated", fmt.Sprintf("ConfigMap %s has been updated", configmap.Name))

	return nil
}

func isDuplicateErr(err error) bool {
	return err.Error() == duplicateError
}
