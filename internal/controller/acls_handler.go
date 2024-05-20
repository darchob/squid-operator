package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"

	squidv1 "git.fr.clara.net/claranet/healthcare/buildops/projects/kubernetes/operators/squid-operator/api/v1"
	"git.fr.clara.net/claranet/healthcare/buildops/projects/kubernetes/operators/squid-operator/pkg/squid"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (sr *AclsReconciler) ensureACLs(ctx context.Context, configs *squidv1.Acls) error {
	log := log.FromContext(ctx)

	log.Info("Check squid configmap for new ACLs", configs.Name)
	currentConfig := corev1.ConfigMap{}

	if err := sr.Client.Get(ctx, client.ObjectKey{Namespace: configs.Namespace, Name: configs.Name}, &currentConfig); err != nil {
		return err
	}

	log.Info("Check %s configmap with new ACLs", configs.Name)
	if !squid.NeedUpdate(configs, &currentConfig) {
		configs.Status.Merged = true
		return nil
	}

	log.Info("Update %s configmap with new ACLs", configs.Name)
	newConfigMap := squid.NewACLS(configs, &currentConfig)
	if err := sr.Update(ctx, newConfigMap); err != nil {
		return err
	}

	configs.Status.Merged = true

	log.Info("Check squid acls %s updated", configs.Name)

	return nil
}

func (sr *AclsReconciler) cleanACLs(ctx context.Context, configs *squidv1.Acls) error {
	log := log.FromContext(ctx)

	log.Info("Check squid configmap to remove ACLs", configs.Name)
	currentConfig := corev1.ConfigMap{}

	if err := sr.Client.Get(ctx, client.ObjectKey{Namespace: configs.Namespace, Name: configs.Name}, &currentConfig); err != nil {
		return err
	}

	log.Info("Check %s configmap with new ACLs", configs.Name)
	if !squid.NeedUpdate(configs, &currentConfig) {
		configs.Status.Merged = true
		return nil
	}

	log.Info("Update %s configmap with new ACLs", configs.Name)
	newConfigMap := squid.RemoveACLs(configs, &currentConfig)
	if err := sr.Update(ctx, newConfigMap); err != nil {
		return err
	}

	return nil
}

func (sr *AclsReconciler) rollingUpdateDeployment(ctx context.Context, configs *squidv1.Acls) error {
	log := log.FromContext(ctx)

	log.Info("Check squid deployment for rolling update ", configs.Name)
	deployment := appsv1.Deployment{}

	if err := sr.Client.Get(ctx, client.ObjectKey{Namespace: configs.Namespace, Name: configs.Name}, &deployment); err != nil {
		if errors.IsNotFound(err) {
			return err
		}
		if err = squid.UpdatedDeployment(&deployment); err != nil {
			log.Info("RollingUpdate Squid %s Deployment", configs.Name)
			if err = sr.Update(ctx, &deployment); err != nil {
				return err
			}
		}
	}

	return nil
}
