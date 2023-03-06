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
	Clientset *kubernetes.Clientset
	RestConf  *rest.Config
	Namespace string
}

const (
	defaultQPS   = 40
	defaultBurst = 80
)

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
	restconf.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(defaultQPS, defaultBurst)
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

type PodLister interface {
	List(ctx context.Context, opts metav1.ListOptions) (*v1.PodList, error)
}

func SelectPods(pods []string, all bool, h PodLister) ([]v1.Pod, error) {
	podList, err := h.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	if all {
		return podList.Items, nil
	}
	resp := []v1.Pod{}
	for _, pod := range podList.Items {
		if isInArray(pod.Name, pods) {
			resp = append(resp, pod)
		}
	}
	return resp, nil
}

func isInArray(s string, array []string) bool {
	for _, a := range array {
		if a == s {
			return true
		}
	}
	return false
}
