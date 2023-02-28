package kpture

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gmtstephane/kpture/pkg/pcap"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestGetClient(t *testing.T) {
	t.Run("invalid kubeconfig", func(t *testing.T) {
		t.Setenv("KUBECONFIG", "invalid")
		_, err := GetClient("")
		assert.NotNil(t, err)
	})

	t.Run("valid kubeconfig", func(t *testing.T) {
		t.Setenv("KUBECONFIG", "../../test/kubeconfigs/sample1.yaml")
		kc, err := GetClient("")
		assert.Nil(t, err)
		assert.NotNil(t, kc)
		assert.NotNil(t, kc.Clientset)
		assert.NotEmpty(t, kc.Namespace)
		assert.NotNil(t, kc.RestConf)
	})

	t.Run("invalid kubeconfig", func(t *testing.T) {
		t.Setenv("KUBECONFIG", "../../test/kubeconfigs/invalidConfig.yaml")
		kc, err := GetClient("")
		assert.NotNil(t, err)
		assert.Nil(t, kc)
	})
}

func Test_generateDebugContainer(t *testing.T) {
	type args struct {
		pod  *corev1.Pod
		name string
		opts pcap.Options
	}
	tests := []struct {
		name string
		args args
		want *corev1.Pod
	}{
		{
			name: "generate debug container",
			args: args{
				pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "test-ns",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "container",
							},
						},
					},
				},
				name: "test-container",
				opts: pcap.Options{
					Port:        8080,
					SnapshotLen: 1024,
					Timeout:     -1,
				},
			},
			want: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "test-ns",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "container",
						},
					},
					EphemeralContainers: []corev1.EphemeralContainer{
						{
							EphemeralContainerCommon: corev1.EphemeralContainerCommon{
								Name: "test-container",
								Env: []corev1.EnvVar{
									{
										Name:  "Kpture_PORT",
										Value: "8080",
									}, {
										Name: "Kpture_DEVICE",
									}, {
										Name:  "Kpture_SNAPSHOT_LEN",
										Value: "1024",
									}, {
										Name:  "Kpture_PROMISCUOUS",
										Value: "false",
									}, {
										Name:  "Kpture_TIMEOUT",
										Value: "-1",
									},
								},
								Image:                    "docker.io/gmtstephane/agent:latest",
								ImagePullPolicy:          "IfNotPresent",
								Stdin:                    true,
								TerminationMessagePolicy: corev1.TerminationMessageReadFile,
								TTY:                      true,
							},
							TargetContainerName: "container",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateDebugContainer(tt.args.pod, tt.args.name, tt.args.opts); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("generateDebugContainer() = %v, want %v", got, tt.want)
			}
		})
	}
}
