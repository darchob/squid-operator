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

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	squidv1 "git.fr.clara.net/claranet/healthcare/buildops/projects/kubernetes/operators/squid-operator/api/v1"
)

// AclsReconciler reconciles a Acls object
type AclsReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=squid.cdk.clara.net,resources=acls,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=squid.cdk.clara.net,resources=acls/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=squid.cdk.clara.net,resources=acls/finalizers,verbs=update

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *AclsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	acl := &squidv1.Acls{}

	log.Info("Squid ACL reconciling %s", req.Name)
	if err := r.Get(ctx, req.NamespacedName, acl); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Delete deployment and ConfigMap
	if !acl.ObjectMeta.DeletionTimestamp.IsZero() {
		acl.ObjectMeta.Finalizers = nil

		if err := r.cleanACLs(ctx, acl); err != nil {
			return ctrl.Result{}, nil
		}

		if err := r.rollingUpdateDeployment(ctx, acl); err != nil {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, nil
	}

	// Change merged status to false
	acl.Status.Merged = false

	// Ensure Finalizer
	if !controllerutil.ContainsFinalizer(acl, finalizerName) {
		acl.ObjectMeta.Finalizers = append(acl.ObjectMeta.Finalizers, finalizerName)
	}

	if err := r.Update(ctx, acl); err != nil {
		log.Error(err, "unable to update Squid acl ", acl.Name)
		return ctrl.Result{}, err
	}

	if err := r.ensureACLs(ctx, acl); err != nil {
		log.Error(err, "unable to update Squid acls ", acl.Name)
		return ctrl.Result{}, err
	}

	//Ensure config is applyed
	if err := r.rollingUpdateDeployment(ctx, acl); err != nil {
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AclsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&squidv1.Acls{}).
		Complete(r)
}
