package squid

import (
	"fmt"
	"log"
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

// Remove block data
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

func generate(current ...string) <-chan int {
	index := make(chan int)

	go func() {
		defer close(index)
		for i := range current {
			index <- i
		}
	}()
	return index
}

func compare(incoming string, currentData []string, current <-chan int) <-chan int {
	index := make(chan int)

	go func() {
		defer close(index)
		for i := range current {
			if reflect.DeepEqual(incoming, currentData[i]) {
				log.Printf("DUPLICATED %s", incoming)
				index <- i
			}
		}
	}()

	return index
}

func exist(data map[string][]string) int {
	out := make(<-chan int)

	for _, line := range data["newData"] {
		gen := generate(data["current"]...)
		out = compare(line, data["current"], gen)
	}

	return <-out
}

func remove(data map[string][]string) []string {

	for _, line := range data["newData"] {
		gen := generate(data["current"]...)
		out := compare(line, data["current"], gen)
		index := 0
		index = <-out

		if index < len(data["current"]) {
			log.Printf("RULES to remove %s", data["current"][index])
			// Remove the element at index i from a.
			data["current"][index] = data["current"][len(data["current"])-1] // Copy last element to index i.
			data["current"][len(data["current"])-1] = ""                     // Erase last element (write zero value).
			data["current"] = data["current"][:len(data["current"])-1]       // Truncate slice.
		}
	}

	return data["current"]
}

func updateData(rules *squidv1.Rules, current *corev1.ConfigMap) (string, error) {
	data := cleanData(rules, current)

	duplicateLine := exist(data)
	if duplicateLine > 0 {
		return "", fmt.Errorf(fmt.Sprintf("%s : %s", DuplicateError, data["current"][duplicateLine]))
	}

	return fmt.Sprintf("%s\n%s",
		strings.Join(data["current"], "\n"),
		strings.Join(data["newData"], "\n")), nil
}

func removeData(rules *squidv1.Rules, current *corev1.ConfigMap) string {
	data := cleanData(rules, current)
	cleanedData := remove(data)
	return strings.TrimSpace(strings.Join(cleanedData, "\n"))
}

// Trim Space
func formatData(data string) []string {
	var newString []string
	lines := strings.Split(strings.TrimRight(data, "\n"), "\n")

	// Line by line
	for _, line := range lines {
		//TrimSpace
		newString = append(newString, strings.TrimSpace(fmt.Sprintf("%s\n", strings.TrimSpace(line))))

	}

	return newString
}
