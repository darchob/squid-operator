package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"

	squidv1 "git.fr.clara.net/claranet/healthcare/buildops/projects/kubernetes/operators/squid-operator/api/v1"
	squid "git.fr.clara.net/claranet/healthcare/buildops/projects/kubernetes/operators/squid-operator/pkg/squid"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *ConfigsReconciler) handlingConfigUpdate(ctx context.Context, configs *squidv1.Configs) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	log.Info("Squid Configs update", "name", configs.Name)
	if err := r.Status().Update(ctx, configs); err != nil {
		log.Error(err, "On Configs CRD update task")
		return ctrl.Result{
			Requeue: true,
		}, err
	}

	return ctrl.Result{}, nil
}

func (sr *ConfigsReconciler) ensureServiceAccount(ctx context.Context, configs *squidv1.Configs) error {
	log := log.FromContext(ctx)
	log.Info("Check squid service account", "name", configs.Name)
	sa := corev1.ServiceAccount{}

	if err := sr.Client.Get(ctx, client.ObjectKey{Namespace: configs.Namespace, Name: configs.Name}, &sa); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		log.Info("Create new Squid ServiceAccount", "name", configs.Name)
		newSa := squid.NewServiceAccount(configs)
		if err = sr.Create(ctx, newSa); err != nil {
			return err
		}

		if err := controllerutil.SetControllerReference(configs, newSa, sr.Scheme); err != nil {
			return err
		}
	}

	log.Info("Check squid ServiceAccount already deployed", "name", configs.Name)

	return nil
}

func (sr *ConfigsReconciler) ensureDeployment(ctx context.Context, configs *squidv1.Configs) error {
	log := log.FromContext(ctx)

	log.Info("Check squid deployment ", "name", configs.Name)
	deployment := appsv1.Deployment{}

	if err := sr.Client.Get(ctx, client.ObjectKey{Namespace: configs.Namespace, Name: configs.Name}, &deployment); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		log.Info("Create new Squid Deployment", "name", configs.Name)
		newDeploy := squid.NewDeployment(configs)
		if err = sr.Create(ctx, newDeploy); err != nil {
			return err
		}

		if err := controllerutil.SetControllerReference(configs, newDeploy, sr.Scheme); err != nil {
			return err
		}
	}

	log.Info("Check squid deployment already deployed", "name", configs.Name)

	return nil
}

func (sr *ConfigsReconciler) ensureDefaultConfigMap(ctx context.Context, configs *squidv1.Configs) error {
	log := log.FromContext(ctx)

	log.Info("Check squid configmap ", "name", configs.Name)
	configmap := corev1.ConfigMap{}

	if err := sr.Client.Get(ctx, client.ObjectKey{Namespace: configs.Namespace, Name: configs.Name}, &configmap); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		log.Info("Create new Squid Default ConfigMap", "name", configs.Name)
		newConfigMap := squid.InitialConfigMap(configs, configs.Spec.SquidConfig)
		if err = sr.Create(ctx, newConfigMap); err != nil {
			return err
		}

		if err := controllerutil.SetControllerReference(configs, newConfigMap, sr.Scheme); err != nil {
			return err
		}
	}

	log.Info("Squid configMap already exist", "name", configs.Name)

	return nil
}
