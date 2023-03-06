package k8s

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gernest/wow"
	"github.com/gernest/wow/spin"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultPolling = time.Second
)

type KubeEphemeralHandler interface {
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.Pod, error)
	UpdateEphemeralContainers(ctx context.Context, podName string, pod *v1.Pod, opts metav1.UpdateOptions) (*v1.Pod, error)
}

func SetupEphemeralContainers(pods []v1.Pod, h KubeEphemeralHandler, opts AgentOpts) error {
	wg := sync.WaitGroup{}
	errchan := make(chan error, len(pods))
	wg.Add(len(pods))

	readychan := make(chan bool, len(pods))
	c := 0
	w := wow.New(os.Stderr, spin.Get(spin.GrowHorizontal), fmt.Sprintf("Creating debug container %d/%d", c, len(pods)))
	w.Start()
	defer w.Stop()
	go func() {
		for {
			<-readychan
			c++
			w.Text(fmt.Sprintf("Creating debug container %d/%d", c, len(pods)))
		}
	}()

	for _, kpturePod := range pods {
		n := kpturePod
		go func() {
			createDebugContainer(n, errchan, &wg, h, opts)
			readychan <- true
		}()
	}
	wg.Wait()
	w.PersistWith(spin.Spinner{}, fmt.Sprintf("Created debug container %d/%d", len(pods), len(pods)))
	log.Println("All debug containers are ready")
	for len(errchan) > 0 {
		err := <-errchan
		if err != nil {
			logrus.Error(err)
			return err
		}
	}
	return nil
}

func createDebugContainer(pod v1.Pod, errchan chan error, wg *sync.WaitGroup, h KubeEphemeralHandler, opts AgentOpts) {
	defer wg.Done()
	err := injectContainer(pod, h, opts, opts.UUID)
	if err != nil {
		errchan <- err
		return
	}

	// wait for the debug container to be ready by polling because watch is not implemented for ephemeral containers
	ctx, cancel := context.WithTimeout(context.Background(), opts.SetupTimeout)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			errchan <- errors.New("timeout waiting for debug container to be ready")
			return
		default:
			pod, errGetPod := h.Get(context.Background(), pod.Name, metav1.GetOptions{})
			if errGetPod != nil {
				errchan <- errGetPod
				return
			}
			for _, eph := range pod.Status.EphemeralContainerStatuses {
				if eph.Name == "kpture-"+opts.UUID {
					if eph.State.Running != nil {
						// log.Println("debug container is Running for pod", pod.Name)
						return
					}
				}
			}
			time.Sleep(defaultPolling)
		}
	}
}

func injectContainer(pod v1.Pod, h KubeEphemeralHandler, opts AgentOpts, id string) error {
	// get the pod to make sure we have the latest version
	// otherwise we might get a conflict error
	syncpod, err := h.Get(context.Background(), pod.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	debugContainer := debugPod(syncpod, "kpture-"+id, opts)

	_, err = h.UpdateEphemeralContainers(context.Background(), pod.Name, debugContainer, metav1.UpdateOptions{})
	if err != nil {
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
	}
	ec := &v1.EphemeralContainer{
		EphemeralContainerCommon: v1.EphemeralContainerCommon{
			Name:            name,
			Args:            args,
			Image:           "ghcr.io/gmtstephane/kpture:latest",
			ImagePullPolicy: v1.PullAlways,
		},
		TargetContainerName: pod.Spec.Containers[0].Name,
	}
	copied := pod.DeepCopy()
	copied.Spec.EphemeralContainers = append(copied.Spec.EphemeralContainers, *ec)
	return copied
}
