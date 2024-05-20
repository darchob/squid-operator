package squid

import (
	"crypto/md5"
	"fmt"
	"strings"

	squidv1 "git.fr.clara.net/claranet/healthcare/buildops/projects/kubernetes/operators/squid-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
)

func DefaultConfigMap(sr *squidv1.Configs) *corev1.ConfigMap {

	return &corev1.ConfigMap{
		ObjectMeta: ObjectMeta(sr),
		Data: map[string]string{
			"squid.conf": sr.Spec.SquidConfig,
		},
	}
}

func updateConfigMap(currentData string, data string) map[string]string {
	return map[string]string{
		"squid.conf": fmt.Sprintf("%s\n%s", currentData, data),
	}
}

func deleteContent(currentData string, data string) map[string]string {
	return map[string]string{
		"suqid.conf": strings.ReplaceAll(currentData, data, ""),
	}
}

func NeedUpdate(acls *squidv1.Acls, current *corev1.ConfigMap) bool {
	currentByte := []byte(current.Data["squid.conf"])
	expectedByte := []byte(updateConfigMap(
		current.Data["squid.conf"],
		acls.Spec.ACLs)["squid.conf"])

	return md5.Sum(currentByte) != md5.Sum(expectedByte)
}

func NewACLS(acls *squidv1.Acls, current *corev1.ConfigMap) *corev1.ConfigMap {
	current.Data = updateConfigMap(current.Data["squid.conf"], acls.Spec.ACLs)
	return current
}

func RemoveACLs(acls *squidv1.Acls, current *corev1.ConfigMap) *corev1.ConfigMap {
	current.Data = deleteContent(current.Data["squid.conf"], acls.Spec.ACLs)
	return current
}
