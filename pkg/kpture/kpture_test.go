package kpture

import (
	"errors"
	"flag"
	"io"
	"math/rand"
	"testing"
	"time"

	"github.com/gmtstephane/kpture/api/capture"

	"github.com/stretchr/testify/assert"
)

var kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")

func TestNewKpture(t *testing.T) {
	flag.Parse()
	t.Setenv("KUBECONFIG", *kubeconfig)

	client, err := GetClient()
	assert.Nil(t, err)

	t.Run("test pod not found ", func(t *testing.T) {
		got, errNewKpture := NewKpture(client, []PodDescriptor{{Name: "test"}})
		assert.NotNil(t, errNewKpture)
		assert.Nil(t, got)
	})

	t.Run("test valid kpture", func(t *testing.T) {
		got, errNewKpture := NewKpture(client, []PodDescriptor{{Name: "podsample"}})
		assert.Nil(t, errNewKpture)
		assert.NotNil(t, got)
		assert.NotNil(t, got.client)
		assert.NotNil(t, got.packetChan)
		assert.NotNil(t, got.errChan)
		assert.NotNil(t, got.kpturePods)
		assert.Equal(t, 1, len(got.kpturePods))
		assert.Equal(t, "podsample", got.kpturePods[0].name)
	})
}

func TestKpture_handleErr(t *testing.T) {
	t.Run("handle error", func(t *testing.T) {
		k := &Kpture{
			client:     nil,
			packetChan: make(chan *capture.Packet),
			errChan:    make(chan error, 1),
			kpturePods: []*Pod{},
		}
		go func() {
			k.handleErr()
		}()
		k.errChan <- errors.New("test error")
	})
}

func TestKpture_SetupEphemeralContainers(t *testing.T) {
	flag.Parse()
	t.Setenv("KUBECONFIG", *kubeconfig)
	// generate random port for test
	rand.Seed(time.Now().UnixNano())
	port := rand.Intn(30000) + 1024
	client, err := GetClient()
	assert.Nil(t, err)
	t.Run("setup ephemeral container", func(t *testing.T) {
		k, errNewKpture := NewKpture(client, []PodDescriptor{{Name: "podsample"}}, WithPort(port))
		assert.Nil(t, errNewKpture)
		errSetupContainer := k.SetupEphemeralContainers()
		assert.Nil(t, errSetupContainer)
	})
}

func TestKpture_SetupPortForwarding(t *testing.T) {
	flag.Parse()
	t.Setenv("KUBECONFIG", *kubeconfig)

	client, err := GetClient()
	assert.Nil(t, err)

	t.Run("port forward", func(t *testing.T) {
		k, errNewKpture := NewKpture(client, []PodDescriptor{{Name: "podsample", Namespace: "default"}})
		assert.Nil(t, errNewKpture)
		errPf := k.SetupPortForwarding()
		assert.Nil(t, errPf)
		k.Stop()
	})
}

func TestKpture_ReadPacketsConn(t *testing.T) {
	flag.Parse()
	t.Setenv("KUBECONFIG", *kubeconfig)

	client, err := GetClient()
	assert.Nil(t, err)

	k, errNewKpture := NewKpture(client, []PodDescriptor{{Name: "podsample", Namespace: "default"}})
	assert.Nil(t, errNewKpture)

	t.Run("start read packets", func(t *testing.T) {
		k.ReadPacketsConn()
	})

	k.Stop()
}

func TestKpture_HandlePackets(t *testing.T) {
	flag.Parse()
	t.Setenv("KUBECONFIG", *kubeconfig)

	client, err := GetClient()
	assert.Nil(t, err)

	k, errNewKpture := NewKpture(client, []PodDescriptor{{Name: "podsample", Namespace: "default"}})
	assert.Nil(t, errNewKpture)

	t.Run("start  handle packet", func(t *testing.T) {
		go func() {
			errHandlePackets := k.HandlePackets(io.Discard)
			assert.Nil(t, errHandlePackets)
		}()
		k.packetChan <- &capture.Packet{}
		close(k.packetChan)
	})
}
