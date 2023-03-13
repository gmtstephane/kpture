package k8s

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
)

type versionGetterMock struct {
	Major string
	Minor string
	err   error
}

func (v *versionGetterMock) ServerVersion() (*version.Info, error) {
	if v.err != nil {
		return nil, v.err
	}

	return &version.Info{
		Major: v.Major,
		Minor: v.Minor,
	}, nil
}

func TestCheckEphemeralContainerSupport(t *testing.T) {
	versionMock := &versionGetterMock{
		err: errors.New("Error fetching kubernetes api"),
	}

	err := CheckEphemeralContainerSupport(versionMock)
	assert.Error(t, err)

	versionMock.err = nil

	versionMock.Major = "1"
	versionMock.Minor = "21"

	err = CheckEphemeralContainerSupport(versionMock)
	assert.Error(t, err)

	versionMock.Major = "1"
	versionMock.Minor = "23"

	err = CheckEphemeralContainerSupport(versionMock)
	assert.NoError(t, err)

	versionMock.Major = "azr"
	versionMock.Minor = "23"

	err = CheckEphemeralContainerSupport(versionMock)
	assert.Error(t, err)

	versionMock.Major = "1"
	versionMock.Minor = "qsd"

	err = CheckEphemeralContainerSupport(versionMock)
	assert.Error(t, err)
}

type podListerMock struct {
	kubeProxyHandlerMockLISTState
}

var podlist []v1.Pod = []v1.Pod{
	{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod1",
			Namespace: "ns-test",
		},
	},
	{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod2",
			Namespace: "ns-test",
		},
	},
}

func (p *podListerMock) List(ctx context.Context, opts metav1.ListOptions) (*v1.PodList, error) {
	if p.kubeProxyHandlerMockLISTState == listError {
		return nil, errors.New("could not list pods")
	}
	if p.kubeProxyHandlerMockLISTState == listOK {
		podlist := v1.PodList{
			Items: podlist,
		}
		return &podlist, nil
	}
	return nil, errors.New("test case not implemented")
}

func TestSelectPods(t *testing.T) {
	mock := podListerMock{}
	mock.kubeProxyHandlerMockLISTState = listError

	// test with fetch error
	_, err := SelectPods([]string{}, false, &mock)
	assert.Error(t, err)

	// test with no selection
	mock.kubeProxyHandlerMockLISTState = listOK
	pods, err := SelectPods([]string{}, false, &mock)
	assert.NoError(t, err)
	assert.Equal(t, pods, []v1.Pod{})

	// test with one pod
	mock.kubeProxyHandlerMockLISTState = listOK
	pods, err = SelectPods([]string{"pod1"}, false, &mock)
	assert.NoError(t, err)
	assert.Equal(t, pods, []v1.Pod{{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod1",
			Namespace: "ns-test",
		},
	}})

	// test with two pods
	mock.kubeProxyHandlerMockLISTState = listOK
	pods, err = SelectPods([]string{"pod1", "pod2"}, false, &mock)
	assert.NoError(t, err)
	assert.Equal(t, pods, []v1.Pod{{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod1",
			Namespace: "ns-test",
		},
	}, {
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod2",
			Namespace: "ns-test",
		},
	}})

	// test with two pods
	mock.kubeProxyHandlerMockLISTState = listOK
	pods, err = SelectPods([]string{""}, true, &mock)
	assert.NoError(t, err)
	assert.Equal(t, pods, podlist)
}

func Test_isInArray(t *testing.T) {
	type args struct {
		s     string
		array []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isInArray(tt.args.s, tt.args.array); got != tt.want {
				t.Errorf("isInArray() = %v, want %v", got, tt.want)
			}
		})
	}
}
