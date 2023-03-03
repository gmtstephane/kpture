package k8s

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gmtstephane/kpture/pkg/kpture"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultPullTimeout = 30
)

func (k *KubeClient) SetupEphemeralContainers(pods []PodDescriptor, opts kpture.Options) error {
	wg := sync.WaitGroup{}
	errchan := make(chan error, len(pods))
	wg.Add(len(pods))
	for _, kpturePod := range pods {
		go k.CreateDebugContainer(kpturePod.Name, errchan, &wg, opts)
	}
	wg.Wait()
	// empty the error channel
	for len(errchan) > 0 {
		err := <-errchan
		if err != nil {
			logrus.Error(err)
			return err
		}
	}
	return nil
}

func (k *KubeClient) CreateDebugContainer(name string, errchan chan error, wg *sync.WaitGroup, opts kpture.Options) {
	defer wg.Done()
	err := k.InjectContainer(name, opts)
	if err != nil {
		errchan <- err
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(defaultPullTimeout)*time.Second)

	defer cancel()
	for {
		select {
		case <-ctx.Done():
			errchan <- errors.New("timeout waiting for debug container to be ready")
			return
		default:
			pod, errGetPod := k.Clientset.Get(context.Background(), name, metav1.GetOptions{})
			if errGetPod != nil {
				errchan <- errGetPod
				return
			}
			for _, eph := range pod.Status.EphemeralContainerStatuses {
				if eph.Name == "kpture-"+opts.UUID {
					if eph.State.Running != nil {
						log.Println("debug container is ready for pod", name)
						return
					}
				}
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func (k *KubeClient) InjectContainer(name string, opts kpture.Options) error {
	pod, err := k.Clientset.Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	debugContainer := debugContainer(pod, "kpture-"+opts.UUID, opts)

	_, err = k.Clientset.UpdateEphemeralContainers(context.Background(), pod.Name, debugContainer, metav1.UpdateOptions{})
	if err != nil {
		logrus.Error(err)
		return err
	}

	return nil
}

func debugContainer(pod *corev1.Pod, name string, opts kpture.Options) *corev1.Pod {
	ec := &corev1.EphemeralContainer{
		EphemeralContainerCommon: corev1.EphemeralContainerCommon{
			Name: name,
			Env: []corev1.EnvVar{
				{
					Name:  "Kpture_PORT",
					Value: fmt.Sprintf("%d", opts.Port),
				},
				{
					Name:  "Kpture_DEVICE",
					Value: opts.Device,
				},
				{
					Name:  "Kpture_SNAPSHOT_LEN",
					Value: fmt.Sprintf("%d", opts.SnapshotLen),
				},
				{
					Name:  "Kpture_PROMISCUOUS",
					Value: fmt.Sprintf("%t", opts.Promiscuous),
				},
				{
					Name:  "Kpture_TIMEOUT",
					Value: fmt.Sprintf("%d", opts.Timeout),
				},
				{
					Name:  "Kpture_PROXY",
					Value: opts.Proxy,
				},
			},
			Image:                    "ghcr.io/gmtstephane/kpture:latest",
			ImagePullPolicy:          corev1.PullNever,
			TerminationMessagePolicy: corev1.TerminationMessageReadFile,
		},
		TargetContainerName: pod.Spec.Containers[0].Name,
	}
	copied := pod.DeepCopy()
	copied.Spec.EphemeralContainers = append(copied.Spec.EphemeralContainers, *ec)
	return copied
}
