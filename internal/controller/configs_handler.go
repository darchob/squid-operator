package controller

import (
	"context"
	"fmt"
	"time"

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
	_ = log.FromContext(ctx)

	if err := r.Status().Update(ctx, configs); err != nil {
		return ctrl.Result{
			Requeue: true,
		}, err
	}

	return ctrl.Result{Requeue: false}, nil
}

func (sr *ConfigsReconciler) ensureServiceAccount(ctx context.Context, configs *squidv1.Configs) error {
	_ = log.FromContext(ctx)

	sa := corev1.ServiceAccount{}
	if err := sr.Client.Get(ctx, client.ObjectKey{Namespace: configs.Namespace, Name: configs.Name}, &sa); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		newSa := squid.NewServiceAccount(configs)
		if err = sr.Create(ctx, newSa); err != nil {
			return err
		}

		if err := controllerutil.SetControllerReference(configs, newSa, sr.Scheme); err != nil {
			return err
		}
	}

	sr.Recorder.Event(configs, "Normal", "Deployed", fmt.Sprintf("ServiceAccount %s has been deployed", sa.Name))

	return nil
}

func (sr *ConfigsReconciler) ensureDeployment(ctx context.Context, configs *squidv1.Configs) error {
	_ = log.FromContext(ctx)

	deployment := appsv1.Deployment{}

	if err := sr.Client.Get(ctx, client.ObjectKey{Namespace: configs.Namespace, Name: configs.Name}, &deployment); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		newDeploy := squid.NewDeployment(configs)
		if err = sr.Create(ctx, newDeploy); err != nil {
			return err
		}

		if err := controllerutil.SetControllerReference(configs, newDeploy, sr.Scheme); err != nil {
			return err
		}
	}

	sr.Recorder.Event(configs, "Normal", "Deployed", fmt.Sprintf("Deployment %s has been deployed", deployment.Name))

	return nil
}

func (sr *ConfigsReconciler) rollingUpdateDeployment(ctx context.Context, rules *squidv1.Rules) error {
	_ = log.FromContext(ctx)

	deployment := appsv1.Deployment{}

	if err := sr.Client.Get(ctx, client.ObjectKey{Namespace: rules.Namespace, Name: rules.Spec.SquidConfig.Name}, &deployment); err != nil {
		return err
	}

	if !rules.Status.Merged {
		return fmt.Errorf("wait for configmap merged")
	}

	deployment.Spec.Template.ObjectMeta.Annotations["squid-operator.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)
	if err := sr.Update(ctx, &deployment); err != nil {
		return err
	}

	return nil
}

func (sr *ConfigsReconciler) ensureDefaultConfigMap(ctx context.Context, configs *squidv1.Configs) error {
	_ = log.FromContext(ctx)

	configmap := corev1.ConfigMap{}

	if err := sr.Client.Get(ctx, client.ObjectKey{Namespace: configs.Namespace, Name: configs.Name}, &configmap); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		newConfigMap := squid.InitialConfigMap(configs, configs.Spec.SquidConfig)
		if err = sr.Create(ctx, newConfigMap); err != nil {
			return err
		}

		if err := controllerutil.SetControllerReference(configs, newConfigMap, sr.Scheme); err != nil {
			return err
		}
	}

	sr.Recorder.Event(configs, "Normal", "Deployed", fmt.Sprintf("ConfigMap %s has been deployed", configmap.Name))
	return nil
}
