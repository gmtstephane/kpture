package kpture

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const DebugContainerName = "kpture-agent-debug"

type KubeClient struct {
	Clientset PodInterface
	RestConf  *rest.Config
	Namespace string
}

func GetClient(namespace string) (*KubeClient, error) {
	configFiles := strings.Split(os.Getenv("KUBECONFIG"), ":")
	config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{Precedence: configFiles},
		&clientcmd.ConfigOverrides{})

	rawConf, err := config.RawConfig()
	if err != nil {
		return nil, errors.New("could not get raw config: " + err.Error())
	}

	restconf, err := config.ClientConfig()
	if err != nil {
		return nil, errors.WithMessage(err, "could not generate clientConfig")
	}

	clientset, err := kubernetes.NewForConfig(restconf)
	if err != nil {
		return nil, errors.WithMessage(err, "could not create clientset")
	}
	if namespace == "" {
		namespace = rawConf.Contexts[rawConf.CurrentContext].Namespace
	}
	return &KubeClient{
		Namespace: namespace,
		Clientset: clientset.CoreV1().Pods(namespace),
		RestConf:  restconf,
	}, nil
}

func (k *KubeClient) SelectPods(pods []string, all bool) ([]PodDescriptor, error) {
	podDescriptors := []PodDescriptor{}
	if all {
		pods, err := k.Clientset.List(context.Background(), v1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, pod := range pods.Items {
			podDescriptors = append(podDescriptors, PodDescriptor{
				Name:      pod.Name,
				Namespace: pod.Namespace,
			})
		}
		return podDescriptors, nil
	}
	for _, pod := range pods {
		kpod, err := k.Clientset.Get(context.Background(), pod, v1.GetOptions{})
		if err != nil {
			return nil, err
		}
		podDescriptors = append(podDescriptors, PodDescriptor{
			Name:      kpod.Name,
			Namespace: kpod.Namespace,
		})
	}
	return podDescriptors, nil
}

func generateDebugContainer(pod *corev1.Pod, name string, opts Options) *corev1.Pod {
	ec := &corev1.EphemeralContainer{
		EphemeralContainerCommon: corev1.EphemeralContainerCommon{
			Name: name,
			Env: []corev1.EnvVar{
				{
					Name:  "Kpture_PORT",
					Value: fmt.Sprintf("%d", opts.Port),
				}, {
					Name:  "Kpture_DEVICE",
					Value: opts.Device,
				}, {
					Name:  "Kpture_SNAPSHOT_LEN",
					Value: fmt.Sprintf("%d", opts.SnapshotLen),
				}, {
					Name:  "Kpture_PROMISCUOUS",
					Value: fmt.Sprintf("%t", opts.Promiscuous),
				}, {
					Name:  "Kpture_TIMEOUT",
					Value: fmt.Sprintf("%d", opts.Timeout),
				},
			},
			Image:                    "docker.io/gmtstephane/agent:latest",
			ImagePullPolicy:          "IfNotPresent",
			Stdin:                    true,
			TerminationMessagePolicy: corev1.TerminationMessageReadFile,
			TTY:                      true,
		},
		TargetContainerName: pod.Spec.Containers[0].Name,
	}
	copied := pod.DeepCopy()
	copied.Spec.EphemeralContainers = append(copied.Spec.EphemeralContainers, *ec)
	return copied
}
