package controller

import (
	"context"

	squidv1 "git.fr.clara.net/claranet/healthcare/buildops/projects/kubernetes/operators/squid-operator/api/v1"
	squid "git.fr.clara.net/claranet/healthcare/buildops/projects/kubernetes/operators/squid-operator/pkg/squid"
	corev1 "k8s.io/api/core/v1"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *RulesReconciler) handlingStatusRules(ctx context.Context, rules *squidv1.Rules) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	controllerlog.Info("RULES Squid Rules update", "name", rules.Name)
	if err := r.Status().Update(ctx, rules); err != nil {
		controllerlog.Error(err, "On Rules CRD update task", "name", rules.Name)
		return ctrl.Result{
			Requeue: true,
		}, err
	}

	return ctrl.Result{}, nil
}

func (r RulesReconciler) EnsureRules(ctx context.Context, rules *squidv1.Rules) error {
	_ = log.FromContext(ctx)

	controllerlog.Info("Check squid configmap to append new Rules", "name", rules.Name)
	currentConfig := corev1.ConfigMap{}

	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: rules.Namespace, Name: rules.Spec.SquidConfig.Name}, &currentConfig); err != nil {
		return err
	}

	controllerlog.Info("Append new Rules", "name", rules.Name)
	newConfigMap, err := squid.NewRules(rules, &currentConfig)
	if err != nil {
		return err
	}

	if err := r.Update(ctx, newConfigMap); err != nil {
		return err
	}

	return nil
}

func (r RulesReconciler) CleanRules(ctx context.Context, rules *squidv1.Rules) error {
	_ = log.FromContext(ctx)

	controllerlog.Info("Check squid configmap to remove Rules", "name", rules.Name)
	currentConfig := corev1.ConfigMap{}

	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: rules.Namespace, Name: rules.Spec.SquidConfig.Name}, &currentConfig); err != nil {
		return err
	}

	newConfigMap := squid.RemoveRules(rules, &currentConfig)

	if err := r.Update(ctx, newConfigMap); err != nil {
		return err
	}

	return nil
}
