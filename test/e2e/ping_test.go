package e2e

import (
	"context"
	"flag"
	"testing"
	"time"

	"github.com/gmtStephane/kpture/pkg/kpture"

	"github.com/stretchr/testify/assert"
)

var kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")

type customWriter struct {
	Reveived bool
}

func (cw *customWriter) Write(p []byte) (int, error) {
	cw.Reveived = true
	return 0, nil
}

func TestInvalidEnvParamError_Error(t *testing.T) {
	flag.Parse()
	t.Setenv("KUBECONFIG", *kubeconfig)

	client, err := kpture.GetClient()
	assert.Nil(t, err)

	podList := []kpture.PodDescriptor{{Name: "podsample-integration", Namespace: "default"}}

	kpture, err := kpture.NewKpture(client, podList)
	assert.Nil(t, err)

	err = kpture.SetupEphemeralContainers()
	assert.Nil(t, err)

	err = kpture.SetupPortForwarding()
	assert.Nil(t, err)

	kpture.ReadPacketsConn()

	cw := &customWriter{}

	go func() {
		errHandlePacket := kpture.HandlePackets(cw)
		assert.Nil(t, errHandlePacket)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	for range time.Tick(1 * time.Second) {
		select {
		case <-ctx.Done():
			if !cw.Reveived {
				t.Error("Expected to receive packets")
			} else {
				t.Log("Received packets")
				return
			}
		default:
		}
	}
}
