package k8s

import (
	"context"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	readinessProbeInitialDelay = int32(5)
	livenessProbeInitialDelay  = int32(10)
	defaultTimeout             = 10 * time.Second
)

func (k KubeClient) SetupProxy(id string, port int32) (string, error) {
	name := "kpture-proxy-" + id
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:            "kpture-proxy",
					ImagePullPolicy: v1.PullIfNotPresent,
					Image:           "ghcr.io/gmtstephane/kpture_proxy:latest",
					Ports: []v1.ContainerPort{
						{
							Name:          "grpc",
							ContainerPort: port,
							Protocol:      v1.ProtocolTCP,
						},
					},
					LivenessProbe: &v1.Probe{
						InitialDelaySeconds: livenessProbeInitialDelay,
						ProbeHandler: v1.ProbeHandler{
							GRPC: &v1.GRPCAction{
								Port: port,
							},
						},
					},
					ReadinessProbe: &v1.Probe{
						InitialDelaySeconds: readinessProbeInitialDelay,
						ProbeHandler: v1.ProbeHandler{
							GRPC: &v1.GRPCAction{
								Port:    port,
								Service: nil,
							},
						},
					},
				},
			},
		},
	}
	_, err := k.Clientset.Create(context.TODO(), &pod, metav1.CreateOptions{})
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
			j, errgetPod := k.Clientset.Get(context.Background(), name, metav1.GetOptions{})
			if errgetPod != nil {
				return "", err
			}
			if j.Status.Phase == "Running" {
				return j.Status.PodIP, nil
			}
		}
	}
}

func (k KubeClient) TearDownProxy(id string) error {
	name := "kpture-proxy-" + id
	err := k.Clientset.Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}
