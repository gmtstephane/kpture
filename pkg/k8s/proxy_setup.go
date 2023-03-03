package k8s

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
)

func (k KubeClient) SetupProxy(id string) error {

	name := "kpture-proxy-" + id
	var backOffLimit int32 = 0

	jobSpec := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  name,
							Image: "ghcr.io/gmtstephane/kpture_proxy:latest",
						},
					},
					RestartPolicy: v1.RestartPolicyNever,
				},
			},
			BackoffLimit: &backOffLimit,
		},
	}

	for {
		j, err := k.Clientset.BatchV1().Jobs(k.Namespace).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			return err
		}
	}

	_, err := k.Clientset.BatchV1().Jobs(k.Namespace).Create(context.TODO(), jobSpec, metav1.CreateOptions{})
	return err
}
