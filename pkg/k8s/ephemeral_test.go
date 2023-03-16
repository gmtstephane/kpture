package k8s

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type kubeEphemeralMock struct {
	kubeProxyHandlerMockUpdateEphState
	kubeProxyHandlerMockGETState
	updatedpod *v1.Pod
}

func (k *kubeEphemeralMock) Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.Pod, error) {
	switch k.kubeProxyHandlerMockGETState {
	case getError:
		return nil, errors.New("Error getting pod")
	case getOkNotReadyThenReady:
		k.kubeProxyHandlerMockGETState = getOK
		k.updatedpod.Status.Phase = v1.PodPending
		return k.updatedpod, nil
	case getOkNotReadyTimeout:
		k.updatedpod.Status.Phase = v1.PodPending
		return k.updatedpod, nil
	case getOKNoEph:
		pod := v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: name}}
		pod.Spec.Containers = []v1.Container{
			{
				Name:  "container-test",
				Image: "image-test",
			},
		}
		pod.Status.Phase = v1.PodRunning
		return &pod, nil
	case getOK:
		if k.updatedpod != nil {
			k.updatedpod.Status.Phase = v1.PodRunning
			eph := []v1.ContainerStatus{}
			for _, s := range k.updatedpod.Spec.EphemeralContainers {
				eph = append(eph, v1.ContainerStatus{
					Name: s.Name,
					State: v1.ContainerState{
						Running: &v1.ContainerStateRunning{},
					},
				})
			}
			k.updatedpod.Status.EphemeralContainerStatuses = eph
			return k.updatedpod, nil
		}
		pod := v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: name}}
		pod.Spec.Containers = []v1.Container{
			{
				Name:  "container-test",
				Image: "image-test",
			},
		}
		pod.Status.Phase = v1.PodRunning
		return &pod, nil
	case getUnkown:
		return nil, errors.New("unkown state")
	default:
		return nil, errors.New("uninplemented state")
	}
}

func (k *kubeEphemeralMock) List(ctx context.Context, opts metav1.ListOptions) (*v1.PodList, error) {
	return &v1.PodList{
		Items: []v1.Pod{},
	}, nil
}

func (k *kubeEphemeralMock) UpdateEphemeralContainers(
	ctx context.Context, name string, pod *v1.Pod, opts metav1.UpdateOptions,
) (*v1.Pod, error) {
	if k.kubeProxyHandlerMockUpdateEphState == updateEphError {
		return nil, errors.New("error updating pod")
	}
	k.updatedpod = pod
	return pod, nil
}

func TestSetupEphemeralContainers(t *testing.T) {
	mock := &kubeEphemeralMock{}
	opts := LoadAgentOpts(WithAgentSetupTimeOut(200 * time.Millisecond))
	pods := []v1.Pod{{
		ObjectMeta: metav1.ObjectMeta{Name: "testpod"},
		Spec: v1.PodSpec{Containers: []v1.Container{
			{
				Name:  "container-test",
				Image: "image-test",
			},
		}},
	}}

	mock.kubeProxyHandlerMockGETState = getError
	err := SetupEphemeralContainers(pods, mock, opts)
	assert.Error(t, err)

	mock.kubeProxyHandlerMockGETState = getOK
	mock.kubeProxyHandlerMockUpdateEphState = updateEphError
	err = SetupEphemeralContainers(pods, mock, opts)
	assert.Error(t, err)

	mock.kubeProxyHandlerMockGETState = getOK
	mock.kubeProxyHandlerMockUpdateEphState = updateEphOK
	err = SetupEphemeralContainers(pods, mock, opts)
	assert.NoError(t, err)
}

func Test_createDebugContainer(t *testing.T) {
	opts := LoadAgentOpts(WithAgentSetupTimeOut(200 * time.Millisecond))
	errchan := make(chan error, 10)
	wg := sync.WaitGroup{}
	mock := &kubeEphemeralMock{}
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "testpod"},
		Spec: v1.PodSpec{Containers: []v1.Container{
			{
				Name:  "container-test",
				Image: "image-test",
			},
		}},
	}

	// mock an error when fetching pod
	wg.Add(1)
	mock.kubeProxyHandlerMockGETState = getError
	createDebugContainer(pod, errchan, &wg, mock, opts)
	errlen := len(errchan)
	assert.Equal(t, 1, errlen)
	for len(errchan) > 0 {
		<-errchan
	}

	// mock an error creating ephemeral container
	wg.Add(1)
	mock.kubeProxyHandlerMockGETState = getOKNoEph
	mock.kubeProxyHandlerMockUpdateEphState = updateEphError
	createDebugContainer(pod, errchan, &wg, mock, opts)
	time.Sleep(opts.SetupTimeout)
	errlen = len(errchan)
	assert.Equal(t, 1, errlen)
	for len(errchan) > 0 {
		<-errchan
	}

	wg.Add(1)
	mock.kubeProxyHandlerMockGETState = getOK
	mock.kubeProxyHandlerMockUpdateEphState = updateEphOK
	createDebugContainer(pod, errchan, &wg, mock, opts)
	errlen = len(errchan)
	assert.Equal(t, 0, errlen)
}

func Test_injectContainer(t *testing.T) {
	opts := LoadAgentOpts()

	mock := &kubeEphemeralMock{}
	mock.kubeProxyHandlerMockGETState = getError
	pod := v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "testpod"}}
	err := injectContainer(pod, mock, opts, "1234")
	assert.Error(t, err)

	mock.kubeProxyHandlerMockGETState = getOKNoEph
	mock.kubeProxyHandlerMockUpdateEphState = updateEphError
	err = injectContainer(pod, mock, opts, "1234")
	assert.Error(t, err)

	mock.kubeProxyHandlerMockUpdateEphState = updateEphOK
	err = injectContainer(pod, mock, opts, "1234")
	assert.NoError(t, err)
}
