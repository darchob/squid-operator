package squid

import (
	"reflect"
	"testing"

	squidv1 "git.fr.clara.net/claranet/healthcare/buildops/projects/kubernetes/operators/squid-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	initRules = `#INIT RULES
Test rules`

	newRules = `
NEW RULES`

	appendRules = `
APPENDED RULES`

	allRules = `
NEW RULES

APPENDED RULES`

	expectedRules = `#INIT RULES
Test rules

NEW RULES`

	expectedNextRules = `#INIT RULES
Test rules

NEW RULES

APPENDED RULES`
)

func TestNewRules(t *testing.T) {
	type args struct {
		rules   *squidv1.Rules
		current *corev1.ConfigMap
	}
	tests := []struct {
		name    string
		args    args
		want    *corev1.ConfigMap
		wantErr bool
	}{
		{
			name: "Test new rule",
			args: args{
				current: &corev1.ConfigMap{
					Data: map[string]string{
						"squid.conf": initRules,
					},
				},
				rules: &squidv1.Rules{
					Spec: squidv1.RulesSpec{
						Name:  "test",
						Rules: newRules,
					},
				},
			},
			wantErr: false,
			want: &corev1.ConfigMap{
				Data: map[string]string{
					"squid.conf": expectedRules,
				},
			},
		},
		{
			name: "Test append rule",
			args: args{
				current: &corev1.ConfigMap{
					Data: map[string]string{
						"squid.conf": expectedRules,
					},
				},
				rules: &squidv1.Rules{
					Spec: squidv1.RulesSpec{
						Name:  "testNext",
						Rules: appendRules,
					},
				},
			},
			wantErr: false,
			want: &corev1.ConfigMap{
				Data: map[string]string{
					"squid.conf": expectedNextRules,
				},
			},
		},
		{
			name: "Test existing rule",
			args: args{
				current: &corev1.ConfigMap{
					Data: map[string]string{
						"squid.conf": expectedNextRules,
					},
				},
				rules: &squidv1.Rules{
					Spec: squidv1.RulesSpec{
						Name:  "testNext",
						Rules: appendRules,
					},
				},
			},
			wantErr: true,
			want: &corev1.ConfigMap{
				Data: map[string]string{
					"squid.conf": expectedNextRules,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewRules(tt.args.rules, tt.args.current)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRules() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRules() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoveRules(t *testing.T) {
	type args struct {
		rules   *squidv1.Rules
		current *corev1.ConfigMap
	}
	tests := []struct {
		name string
		args args
		want *corev1.ConfigMap
	}{
		{
			name: "Test delete rule",
			args: args{
				current: &corev1.ConfigMap{
					Data: map[string]string{
						"squid.conf": expectedRules,
					},
				},
				rules: &squidv1.Rules{
					Spec: squidv1.RulesSpec{
						Name:  "test",
						Rules: newRules,
					},
				},
			},
			want: &corev1.ConfigMap{
				Data: map[string]string{
					"squid.conf": initRules,
				},
			},
		},
		{
			name: "Test delete with different name rule",
			args: args{
				current: &corev1.ConfigMap{
					Data: map[string]string{
						"squid.conf": expectedNextRules,
					},
				},
				rules: &squidv1.Rules{
					Spec: squidv1.RulesSpec{
						Name:  "testOne",
						Rules: allRules,
					},
				},
			},
			want: &corev1.ConfigMap{
				Data: map[string]string{
					"squid.conf": initRules,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RemoveRules(tt.args.rules, tt.args.current); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveRules() = %v, want %v", got, tt.want)
			}
		})
	}
}
