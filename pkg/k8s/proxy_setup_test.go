package k8s

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type kubeProxyHandlerMock struct {
	kubeProxyHandlerMockGETState
	kubeProxyHandlerMockCREATEState
	kubeProxyHandlerMockDELETEState
	createdPod *v1.Pod
}

func (k *kubeProxyHandlerMock) Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.Pod, error) {
	switch k.kubeProxyHandlerMockGETState {
	case getError:
		return nil, errors.New("Error getting pod")
	case getOkNotReadyThenReady:
		k.kubeProxyHandlerMockGETState = getOK
		k.createdPod.Status.Phase = v1.PodPending
		return k.createdPod, nil
	case getOkNotReadyTimeout:
		k.createdPod.Status.Phase = v1.PodPending
		return k.createdPod, nil
	case getOK, getOKNoEph:
		k.createdPod.Status.Phase = v1.PodRunning
		k.createdPod.Status.PodIP = "10.102.0.4"
		return k.createdPod, nil
	case getUnkown:
		return nil, errors.New("uninplemented state")
	default:
		return nil, errors.New("uninplemented state")
	}
}

func (k *kubeProxyHandlerMock) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	if k.kubeProxyHandlerMockDELETEState == deleteError {
		return errors.New("Error deleting pod")
	}
	return nil
}

func (k *kubeProxyHandlerMock) Create(ctx context.Context, pod *v1.Pod, opts metav1.CreateOptions) (*v1.Pod, error) {
	if k.kubeProxyHandlerMockCREATEState == createError {
		return nil, errors.New("Error creating pod")
	}
	k.createdPod = pod
	return pod, nil
}

func TestSetupProxy(t *testing.T) {
	mock := &kubeProxyHandlerMock{}
	mock.kubeProxyHandlerMockCREATEState = createError
	opts := LoadProxyOpts(WithProxyUUID("1234"), WithProxySetupTimeout(1*time.Second))
	_, err := SetupProxy(mock, opts)
	assert.Error(t, err)

	mock.kubeProxyHandlerMockCREATEState = createOK
	mock.kubeProxyHandlerMockGETState = getError
	_, err = SetupProxy(mock, opts)
	assert.Error(t, err)

	mock.kubeProxyHandlerMockCREATEState = createOK
	mock.kubeProxyHandlerMockGETState = getOkNotReadyThenReady
	ip, err := SetupProxy(mock, opts)
	assert.NoError(t, err)
	assert.Equal(t, ip, "10.102.0.4")
	mock.kubeProxyHandlerMockCREATEState = createOK
	mock.kubeProxyHandlerMockGETState = getOkNotReadyTimeout
	_, err = SetupProxy(mock, opts)
	assert.Error(t, err)
}

func TestTearDownProxy(t *testing.T) {
	mock := &kubeProxyHandlerMock{}
	mock.kubeProxyHandlerMockDELETEState = deleteError
	err := TearDownProxy("1234", mock)
	assert.Error(t, err)

	mock.kubeProxyHandlerMockDELETEState = deleteOk
	err = TearDownProxy("1234", mock)
	assert.NoError(t, err)
}
