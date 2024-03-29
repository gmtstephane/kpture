package k8s

import (
	"time"
)

type (
	ProxyOpt func(o ProxyOpts) ProxyOpts
	AgentOpt func(o AgentOpts) AgentOpts
)

// Proxy Options.
const (
	proxyDefaultServerPort   int32         = 10000
	proxyDefaultSetupTimeout time.Duration = 20 * time.Second
)

// ProxyOpts are the options for the proxy server
type ProxyOpts struct {
	ServerPort   int32
	UUID         string
	SetupTimeout time.Duration // timeout for pod creation
}

func defaultProxyOpts() ProxyOpts {
	return ProxyOpts{
		ServerPort:   proxyDefaultServerPort,
		SetupTimeout: proxyDefaultSetupTimeout,
	}
}

// LoadProxyOpts loads the proxy options
func LoadProxyOpts(os ...ProxyOpt) ProxyOpts {
	opts := defaultProxyOpts()
	for _, o := range os {
		opts = o(opts)
	}
	return opts
}

// WithProxyUUID sets the proxy uuid
func WithProxyUUID(u string) ProxyOpt {
	return func(o ProxyOpts) ProxyOpts {
		o.UUID = u
		return o
	}
}

// WithProxySetupTimeout sets the proxy setup timeout
func WithProxySetupTimeout(u time.Duration) ProxyOpt {
	return func(o ProxyOpts) ProxyOpts {
		o.SetupTimeout = u
		return o
	}
}

// WithProxyServerPort sets the proxy server port
func WithProxyServerPort(u int32) ProxyOpt {
	return func(o ProxyOpts) ProxyOpts {
		o.ServerPort = u
		return o
	}
}

// Agent Options.
const (
	agentDdefaultSnapLen     int32         = 1500
	agentDefaultPromiscuous  bool          = true
	agentDefaultDevice       string        = "eth0"
	agentDefaultTimeout      time.Duration = -1 * time.Second
	agentDefaultSetupTimeout time.Duration = 20 * time.Second
)

// AgentOpts are the options for the capture agent
type AgentOpts struct {
	SnapshotLen  int32         // https://www.tcpdump.org/manpages/pcap_set_snaplen.3pcap.html
	Promiscuous  bool          // https://www.tcpdump.org/manpages/pcap_set_promisc.3pcap.html
	Device       string        // https://www.tcpdump.org/manpages/pcap_create.3pcap.html
	Timeout      time.Duration // https://www.tcpdump.org/manpages/pcap_set_timeout.3pcap.html
	TargetIP     string        // proxy endpoint address to send packet via gRPC
	TargetPort   int           // proxy endpoint port to send packet via gRPC
	UUID         string        // kpture uuid used in  ephemeral container name
	Filter       string        // https://www.tcpdump.org/manpages/pcap_compile.3pcap.html
	SetupTimeout time.Duration // timeout for ephemeral container injection
}

func defaultAgentOpts() AgentOpts {
	return AgentOpts{
		SnapshotLen:  agentDdefaultSnapLen,
		Promiscuous:  agentDefaultPromiscuous,
		Device:       agentDefaultDevice,
		Timeout:      agentDefaultTimeout,
		SetupTimeout: agentDefaultSetupTimeout,
	}
}

func LoadAgentOpts(os ...AgentOpt) AgentOpts {
	opts := defaultAgentOpts()
	for _, o := range os {
		opts = o(opts)
	}
	return opts
}

// WithTargetIP sets the target IP
func (a AgentOpts) WithTargetIP(s string) AgentOpts {
	a.TargetIP = s
	return a
}

// WithTargetPort sets the target port
func (a AgentOpts) WithTargetPort(p int) AgentOpts {
	a.TargetPort = p
	return a
}

// WithAgentDevice sets the device to capture on
func WithAgentDevice(n string) AgentOpt {
	return func(o AgentOpts) AgentOpts {
		o.Device = n
		return o
	}
}

// WithAgentPromiscuous sets the promiscuous mode
func WithAgentPromiscuous(n bool) AgentOpt {
	return func(o AgentOpts) AgentOpts {
		o.Promiscuous = n
		return o
	}
}

// WithAgentCaptureFilter sets the capture filter
func WithAgentCaptureFilter(n string) AgentOpt {
	return func(o AgentOpts) AgentOpts {
		o.Filter = n
		return o
	}
}

// WithAgentSnapLen sets the snapshot length
func WithAgentSnapLen(n int32) AgentOpt {
	return func(o AgentOpts) AgentOpts {
		o.SnapshotLen = n
		return o
	}
}

// WithAgentCaptureTimeOut sets the capture timeout
func WithAgentCaptureTimeOut(n time.Duration) AgentOpt {
	return func(o AgentOpts) AgentOpts {
		o.Timeout = n
		return o
	}
}

// WithAgentSetupTimeOut sets the agent setup timeout
func WithAgentSetupTimeOut(n time.Duration) AgentOpt {
	return func(o AgentOpts) AgentOpts {
		o.SetupTimeout = n
		return o
	}
}

// WithAgentUUID sets the agent uuid
func WithAgentUUID(u string) AgentOpt {
	return func(o AgentOpts) AgentOpts {
		o.UUID = u
		return o
	}
}
