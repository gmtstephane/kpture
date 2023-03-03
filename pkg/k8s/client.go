package k8s

import (
	"context"
	"os"
	"strings"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/flowcontrol"
)

type KubeClient struct {
	Clientset PodInterface
	// Clientset *kubernetes.Clientset
	RestConf  *rest.Config
	Namespace string
}

type PodInterface interface {
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.Pod, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1.PodList, error)
	UpdateEphemeralContainers(ctx context.Context, podName string, pod *v1.Pod, opts metav1.UpdateOptions) (*v1.Pod, error)
	Create(ctx context.Context, pod *v1.Pod, opts metav1.CreateOptions) (*v1.Pod, error)
}

type PodDescriptor struct {
	Name      string
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
	restconf.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(40, 80)
	clientset, err := kubernetes.NewForConfig(restconf)
	if err != nil {
		return nil, errors.WithMessage(err, "could not create clientset")
	}

	if namespace == "" {
		namespace = rawConf.Contexts[rawConf.CurrentContext].Namespace
		if namespace == "" {
			namespace = "default"
		}
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
		pods, err := k.Clientset.List(context.Background(), metav1.ListOptions{})
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
		kpod, err := k.Clientset.Get(context.Background(), pod, metav1.GetOptions{})
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
