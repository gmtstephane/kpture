package k8s

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultPolling = 1 * time.Second
)

type KubeEphemeralHandler interface {
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.Pod, error)
	UpdateEphemeralContainers(ctx context.Context, podName string, pod *v1.Pod, opts metav1.UpdateOptions) (*v1.Pod, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1.PodList, error)
}

func SetupEphemeralContainers(pods []v1.Pod, h KubeEphemeralHandler, opts AgentOpts) error {
	errchan := make(chan error, len(pods))

	// create the debug container in all the pods
	// the wait group is done when all kubernetes api calls are done
	// not when the debug containers are actually running
	wg := sync.WaitGroup{}
	wg.Add(len(pods))
	for _, kpturePod := range pods {
		n := kpturePod
		go func() {
			createDebugContainer(n, errchan, &wg, h, opts)
		}()
	}
	wg.Wait()

	// If we have any error creating containers we return the first one
	for len(errchan) > 0 {
		err := <-errchan
		if err != nil {
			logrus.Error(err)
			return err
		}
	}

	// Readyness and Liveness probes are not supported for ephemeral containers
	// Kubernetes is quiet long to detect containers as running.
	// Start a watcher in case an ephemeral container fails to start
	go watcher(h, pods, opts)

	return nil
}

// watcher checks all the pods until they are all in a running state
// or until one of them is in a terminated state for an error.
func watcher(h KubeEphemeralHandler, pods []v1.Pod, opts AgentOpts) {
	for {
		list, err := h.List(context.Background(), metav1.ListOptions{})
		if err != nil {
			log.Println(err)
			return
		}
		countPod := 0
		for _, pod := range list.Items {
			if isPodInArray(pod.Name, pods) {
				var errStatus error
				if countPod, errStatus = checkpodStatus(pod, opts, countPod); errStatus != nil {
					log.Println(errStatus)
					return
				}
			}
		}
		if countPod == len(pods) {
			return
		}
		time.Sleep(defaultPolling)
	}
}

func checkpodStatus(pod v1.Pod, opts AgentOpts, countPod int) (int, error) {
	for _, eph := range pod.Status.EphemeralContainerStatuses {
		if eph.Name == "kpture-"+opts.UUID {
			if eph.State.Running != nil {
				countPod++
				break
			}
			if eph.State.Terminated != nil {
				return 0, errors.New("error setting pod " + eph.State.Terminated.Message)
			}
		}
	}
	return countPod, nil
}

func createDebugContainer(pod v1.Pod, errchan chan error, wg *sync.WaitGroup, h KubeEphemeralHandler, opts AgentOpts) {
	defer wg.Done()
	err := injectContainer(pod, h, opts, opts.UUID)
	if err != nil {
		errchan <- err
		return
	}
}

func injectContainer(pod v1.Pod, h KubeEphemeralHandler, opts AgentOpts, id string) error {
	// get the pod again to make sure we have the latest version
	// otherwise we might get a conflict error
	syncpod, err := h.Get(context.Background(), pod.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if _, err = h.UpdateEphemeralContainers(
		context.Background(),
		pod.Name,
		debugPod(syncpod, "kpture-"+id, opts),
		metav1.UpdateOptions{}); err != nil {
		return err
	}

	return nil
}

func debugPod(pod *v1.Pod, name string, opts AgentOpts) *v1.Pod {
	args := []string{
		"agent",
		fmt.Sprintf("-d%s", opts.Device),
		fmt.Sprintf("-t%s", opts.TargetIP),
		fmt.Sprintf("-l%d", opts.SnapshotLen),
		fmt.Sprintf("-p%d", opts.TargetPort),
		fmt.Sprintf("-f%s", opts.Filter),
	}
	p := true
	f := false
	user := int64(1000)
	ec := &v1.EphemeralContainer{
		EphemeralContainerCommon: v1.EphemeralContainerCommon{
			Name: name,
			SecurityContext: &v1.SecurityContext{
				RunAsUser: &user,
				Capabilities: &v1.Capabilities{
					Add: []v1.Capability{"NET_ADMIN", "NET_RAW"},
				},
				RunAsNonRoot:             &p,
				Privileged:               &f,
				AllowPrivilegeEscalation: &f,
			},
			Args:            args,
			Image:           "ghcr.io/gmtstephane/kpture:latest",
			ImagePullPolicy: v1.PullIfNotPresent,
		},
		TargetContainerName: pod.Spec.Containers[0].Name,
	}
	copied := pod.DeepCopy()
	copied.Spec.EphemeralContainers = append(copied.Spec.EphemeralContainers, *ec)
	return copied
}
