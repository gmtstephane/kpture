package k8s

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/flowcontrol"
)

// KubeClient is a wrapper around the kubernetes clientset
type KubeClient struct {
	Clientset *kubernetes.Clientset
	RestConf  *rest.Config
	Namespace string
}

const (
	defaultQPS   = 40
	defaultBurst = 80
)

// GetClient returns a new KubeClient from the KUBECONFIG with a default namespace if none is provided
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

type VersionGetter interface {
	ServerVersion() (*version.Info, error)
}

// CheckEphemeralContainerSupport checks if the cluster supports ephemeral containers
func CheckEphemeralContainerSupport(v VersionGetter) error {
	version, err := v.ServerVersion()
	if err != nil {
		return err
	}

	major, err := strconv.ParseInt(version.Major, 10, 64)
	if err != nil {
		return err
	}
	minor, err := strconv.ParseInt(version.Minor, 10, 64)
	if err != nil {
		return err
	}

	if major < 1 || (major == 1 && minor < 22) {
		return errors.New("Ephemeral containers are not supported")
	}

	return nil
}

type PodLister interface {
	List(ctx context.Context, opts metav1.ListOptions) (*v1.PodList, error)
}

// SelectPods select pods from a list of pods or all pods
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

// CheckPodsContext check if all pods are in running state
func CheckPodsContext(pods []v1.Pod) error {
	// Check if all pod are in running state
	// Check if all pod have security context thath allow to run as root and capture packets
	for _, pod := range pods {
		if pod.Status.Phase != "Running" {
			return errors.New("pod " + pod.Name + " is not in running state")
		}
	}
	return nil
}

// isInArray check if a string is in an array of string
func isInArray(s string, array []string) bool {
	for _, a := range array {
		if a == s {
			return true
		}
	}
	return false
}

// isPodInArray check if a pod is in an array of pod
func isPodInArray(s string, array []v1.Pod) bool {
	for _, a := range array {
		if a.Name == s {
			return true
		}
	}
	return false
}
