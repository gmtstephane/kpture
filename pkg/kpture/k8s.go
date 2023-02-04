package kpture

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
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

func GetClient() (*KubeClient, error) {
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

	ns := rawConf.Contexts[rawConf.CurrentContext].Namespace
	return &KubeClient{
		Namespace: ns,
		Clientset: clientset.CoreV1().Pods(ns),
		RestConf:  restconf,
	}, nil
}

func generateDebugContainer(pod *corev1.Pod, name string, opts ServerOptions) *corev1.Pod {
	ec := &corev1.EphemeralContainer{
		EphemeralContainerCommon: corev1.EphemeralContainerCommon{
			Name: name,
			Env: []corev1.EnvVar{
				{
					Name:  "Kpture_PORT",
					Value: fmt.Sprintf("%d", opts.port),
				}, {
					Name:  "Kpture_DEVICE",
					Value: opts.device,
				}, {
					Name:  "Kpture_SNAPSHOT_LEN",
					Value: fmt.Sprintf("%d", opts.snapshotLen),
				}, {
					Name:  "Kpture_PROMISCUOUS",
					Value: fmt.Sprintf("%t", opts.promiscuous),
				}, {
					Name:  "Kpture_TIMEOUT",
					Value: fmt.Sprintf("%d", opts.timeout),
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
