package squid

import (
	"fmt"
	"reflect"
	"strings"

	squidv1 "git.fr.clara.net/claranet/healthcare/buildops/projects/kubernetes/operators/squid-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
)

// Add new rules to current configmap
func NewRules(rules *squidv1.Rules, current *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	data, err := updateData(rules, current)
	if err != nil {
		return current, err
	}
	current.Data["squid.conf"] = data
	return current, nil
}

// Remove rules from current ConfigMap
func RemoveRules(rules *squidv1.Rules, current *corev1.ConfigMap) *corev1.ConfigMap {
	current.Data["squid.conf"] = removeData(rules, current)
	return current
}

// Init first configMap from SquidConfig CRD
func InitialConfigMap(sr *squidv1.Configs, data string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: ObjectMeta(sr),
		Data: map[string]string{
			"squid.conf": data,
		},
	}
}

// Utils
func cleanData(rules *squidv1.Rules, current *corev1.ConfigMap) map[string][]string {
	data := make(map[string][]string)
	newData := make(chan []string)
	currentData := make(chan []string)

	go func() {
		currentData <- formatData(current.Data["squid.conf"])
	}()

	go func() {
		newData <- formatData(rules.Spec.Rules)
	}()

	data["current"] = <-currentData
	data["newData"] = <-newData

	return data
}

func compare(data map[string][]string) {
	diff := make(chan bool)

	for _, current := range data["current"] {
		go func() {
			diff <- exist(current, data["newData"])
		}()
	}

}

func updateData(rules *squidv1.Rules, current *corev1.ConfigMap) (string, error) {
	data := cleanData(rules, current)

	return "", fmt.Errorf(noDiffError)
}

func removeData(rules *squidv1.Rules, current *corev1.ConfigMap) string {
	data := cleanData(rules, current)

	return strings.ReplaceAll(data["current"], data["newData"], "")
}

func exist(current string, incoming []string) bool {
	for line := range incoming {
		if reflect.DeepEqual(current, line) {
			return true
		}
	}

	return false
}

func formatData(data string) []string {
	var newString []string
	// Line by line
	for _, line := range strings.Split(strings.TrimRight(data, "\n"), "\n") {
		//TrimSpace
		newString = append(newString, strings.TrimSpace(line))

	}

	return newString
}
