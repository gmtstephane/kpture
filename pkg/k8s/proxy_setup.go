package k8s

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	readinessProbeInitialDelay = int32(5)
	livenessProbeInitialDelay  = int32(10)
)

type KubeProxyHandler interface {
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.Pod, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	Create(ctx context.Context, pod *v1.Pod, opts metav1.CreateOptions) (*v1.Pod, error)
}

func SetupProxy(h KubeProxyHandler, opts ProxyOpts) (string, error) {
	name := "kpture-proxy-" + opts.UUID
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				debugContainer(opts),
			},
		},
	}
	_, err := h.Create(context.TODO(), &pod, metav1.CreateOptions{})
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), opts.SetupTimeout)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
			j, errgetPod := h.Get(context.Background(), name, metav1.GetOptions{})
			if errgetPod != nil {
				return "", errgetPod
			}
			if j.Status.Phase == v1.PodRunning {
				return j.Status.PodIP, nil
			}
		}
	}
}

func TearDownProxy(id string, h KubeProxyHandler) error {
	err := h.Delete(context.Background(), "kpture-proxy-"+id, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	return nil
}

func debugContainer(opts ProxyOpts) v1.Container {
	return v1.Container{
		Name:            "kpture-proxy",
		ImagePullPolicy: v1.PullIfNotPresent,
		Image:           "ghcr.io/gmtstephane/kpture_proxy:latest",
		Args:            []string{"proxy"},
		Ports: []v1.ContainerPort{
			{
				Name:          "grpc",
				ContainerPort: opts.ServerPort,
				Protocol:      v1.ProtocolTCP,
			},
		},
		LivenessProbe: &v1.Probe{
			InitialDelaySeconds: livenessProbeInitialDelay,
			ProbeHandler: v1.ProbeHandler{
				GRPC: &v1.GRPCAction{
					Port: opts.ServerPort,
				},
			},
		},
		ReadinessProbe: &v1.Probe{
			InitialDelaySeconds: readinessProbeInitialDelay,
			ProbeHandler: v1.ProbeHandler{
				GRPC: &v1.GRPCAction{
					Port:    opts.ServerPort,
					Service: nil,
				},
			},
		},
	}
}
