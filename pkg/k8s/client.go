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
)

type KubeClient struct {
	// Clientset PodInterface
	Clientset *kubernetes.Clientset
	RestConf  *rest.Config
	Namespace string
}

type PodInterface interface {
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.Pod, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1.PodList, error)
	UpdateEphemeralContainers(ctx context.Context, podName string, pod *v1.Pod, opts metav1.UpdateOptions) (*v1.Pod, error)
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
		if namespace == "" {
			namespace = "default"
		}
	}
	return &KubeClient{
		Namespace: namespace,
		Clientset: clientset,
		RestConf:  restconf,
	}, nil
}
