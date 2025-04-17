package controller

import (
	"fmt"

	squidv1 "git.fr.clara.net/claranet/healthcare/buildops/projects/kubernetes/operators/squid-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	duplicateError = "dupplicate key found %s"
)

const (
	finalizerName = "squid.ckd.clara.net/finalizer"
	configField   = "metadata.annotations[\"squid.ckd.clara.net/instance\"]"
)

func isDuplicateErr(err error) bool {
	return err.Error() == duplicateError
}

func handleConfigMap(obj client.Object, configs *squidv1.SquidConfigs, truncate bool) (*corev1.ConfigMap, error) {
	dataKey := fmt.Sprintf("%s.conf", configs.Name)
	configmap, ok := obj.DeepCopyObject().(*corev1.ConfigMap)
	if !ok {
		return nil, fmt.Errorf("object is not a ConfigMap")
	}

	if configmap.Data == nil {
		configmap.Data = make(map[string]string)
	}

	if truncate {
		delete(configmap.Data, dataKey)
		return configmap, nil
	}

	_, ok = configmap.Data[dataKey]
	if ok {
		return nil, fmt.Errorf(duplicateError, dataKey)
	}

	configmap.Data[dataKey] = configs.Spec.Rules

	return configmap, nil
}
