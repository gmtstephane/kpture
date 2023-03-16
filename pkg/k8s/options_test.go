package k8s

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_defaultProxyOpts(t *testing.T) {
	opts := defaultProxyOpts()
	assert.Equal(t, opts.ServerPort, proxyDefaultServerPort)
	assert.Equal(t, opts.SetupTimeout, proxyDefaultSetupTimeout)
	assert.Empty(t, opts.UUID)
}

func TestLoadProxyOpts(t *testing.T) {
	opts := LoadProxyOpts()
	assert.Equal(t, opts, defaultProxyOpts())
	opts = LoadProxyOpts(WithProxyServerPort(8080))
	assert.Equal(t, opts.ServerPort, int32(8080))
}

func Test_defaultAgentOpts(t *testing.T) {
	opts := defaultAgentOpts()
	assert.Equal(t, opts.Device, agentDefaultDevice)
	assert.Equal(t, opts.Promiscuous, agentDefaultPromiscuous)
	assert.Equal(t, opts.SetupTimeout, agentDefaultSetupTimeout)
	assert.Equal(t, opts.SnapshotLen, agentDdefaultSnapLen)
	assert.Empty(t, opts.TargetIP)
	assert.Empty(t, opts.TargetPort)
	assert.Empty(t, opts.UUID)
}

func TestLoadAgentOpts(t *testing.T) {
	id := "1234"
	timeout := 10 * time.Second
	captureTimeout := 5 * time.Second
	snaplen := int32(10)
	promisc := false
	device := "en0"

	opts := LoadAgentOpts(
		WithAgentUUID(id),
		WithAgentSetupTimeOut(timeout),
		WithAgentCaptureTimeOut(captureTimeout),
		WithAgentSnapLen(snaplen),
		WithAgentPromiscuous(promisc),
		WithAgentDevice(device),
	)
	assert.Equal(t, opts.Device, device)
	assert.Equal(t, opts.UUID, id)
	assert.Equal(t, opts.SetupTimeout, timeout)
	assert.Equal(t, opts.Timeout, captureTimeout)
	assert.Equal(t, opts.Promiscuous, promisc)
	assert.Equal(t, opts.SnapshotLen, opts.SnapshotLen)

	opts = opts.WithTargetIP("localhost").WithTargetPort(8080)
	assert.Equal(t, opts.TargetIP, "localhost")
	assert.Equal(t, opts.TargetPort, 8080)
}
