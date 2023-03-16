package e2e

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/gmtstephane/kpture/pkg/k8s"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSetup(t *testing.T) {
	log.SetFlags(0)
	log.SetOutput(os.Stderr)
	client, err := k8s.GetClient("")
	assert.NoError(t, err)
	podlist, err := client.Clientset.CoreV1().Pods(client.Namespace).List(context.Background(), v1.ListOptions{})
	assert.NoError(t, err)
	assert.NotEmpty(t, podlist.Items)

	kptureID := uuid.New().String()
	agentOpts := k8s.LoadAgentOpts(k8s.WithAgentUUID(kptureID))
	proxyOpts := k8s.LoadProxyOpts(k8s.WithProxyUUID(kptureID))

	log.Println("Deploying Proxy")
	assert.NotNil(t, agentOpts)
	assert.NotNil(t, proxyOpts)
	ip, err := k8s.SetupProxy(client.Clientset.CoreV1().Pods(client.Namespace), proxyOpts)
	assert.NoError(t, err)
	assert.NotEmpty(t, ip)
	agentOpts = agentOpts.WithTargetIP(ip).WithTargetPort(int(proxyOpts.ServerPort))

	err = k8s.SetupEphemeralContainers(podlist.Items, client.Clientset.CoreV1().Pods(client.Namespace), agentOpts)
	assert.NoError(t, err)

	err = k8s.TearDownProxy(kptureID, client.Clientset.CoreV1().Pods(client.Namespace))
	assert.NoError(t, err)
}
