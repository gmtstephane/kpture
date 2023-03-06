package k8s

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetKubeForwarder(t *testing.T) {
	t.Setenv("KUBECONFIG", "../../ci/samples/kubeconfig")
	client, err := GetClient("")
	assert.NoError(t, err)
	assert.NotNil(t, client)
	readychan, stopchan := make(chan struct{}, 1), make(chan struct{}, 1)
	serverport := int32(10000)
	fw, port, err := GetKubeForwarder(
		client.RestConf,
		fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", client.Namespace, "kpture-proxy-"+"12345"),
		readychan,
		stopchan,
		serverport,
	)
	assert.NoError(t, err)
	assert.NotEmpty(t, port)
	assert.NotNil(t, fw)
}

type mockForwarder struct {
	mu  sync.Mutex
	Err bool
}

func (m *mockForwarder) getErr() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Err
}

func (m *mockForwarder) setErr(b bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Err = b
}

func (m *mockForwarder) ForwardPorts() error {
	if m.getErr() {
		return errors.New("Error forwarding pod")
	}
	return nil
}

func TestPortForward(t *testing.T) {
	fw := &mockForwarder{Err: false, mu: sync.Mutex{}}
	ch := make(chan struct{}, 1)

	// PortForward with timeout
	err := PortForward(fw, ch, 1*time.Second)
	assert.Error(t, err)

	// Port forward with ready
	ch <- struct{}{}
	err = PortForward(fw, ch, 1*time.Second)
	assert.NoError(t, err)

	// Port forward with error
	fw.setErr(true)
	ch = make(chan struct{}, 1)
	err = PortForward(fw, ch, 1*time.Second)
	assert.Error(t, err)
}
