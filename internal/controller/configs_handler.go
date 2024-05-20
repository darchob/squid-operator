package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"

	squidv1 "git.fr.clara.net/claranet/healthcare/buildops/projects/kubernetes/operators/squid-operator/api/v1"
	squid "git.fr.clara.net/claranet/healthcare/buildops/projects/kubernetes/operators/squid-operator/pkg/squid"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (sr *ConfigsReconciler) ensureDeployment(ctx context.Context, configs *squidv1.Configs) error {
	log := log.FromContext(ctx)

	log.Info("Check squid deployment ", configs.Name)
	deployment := appsv1.Deployment{}

	if err := sr.Client.Get(ctx, client.ObjectKey{Namespace: configs.Namespace, Name: configs.Name}, &deployment); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		log.Info("Create new Squid %s Deployment", configs.Name)
		newDeploy := squid.NewDeployment(configs)
		if err = sr.Create(ctx, newDeploy); err != nil {
			return err
		}
	}

	log.Info("Check squid deployment %s already deployed", configs.Name)

	return nil
}

func (sr *ConfigsReconciler) ensureDefaultConfigMap(ctx context.Context, configs *squidv1.Configs) error {
	log := log.FromContext(ctx)

	log.Info("Check squid configmap ", configs.Name)
	configmap := corev1.ConfigMap{}

	if err := sr.Client.Get(ctx, client.ObjectKey{Namespace: configs.Namespace, Name: configs.Name}, &configmap); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		log.Info("Create new Squid %s Default ConfigMap", configs.Name)
		neConfigMap := squid.DefaultConfigMap(configs)
		if err = sr.Create(ctx, neConfigMap); err != nil {
			return err
		}
	}

	log.Info("Squid configMap %s already exist", configs.Name)

	return nil
}
