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

func LoadProxyOpts(os ...ProxyOpt) ProxyOpts {
	opts := defaultProxyOpts()
	for _, o := range os {
		opts = o(opts)
	}
	return opts
}

func WithProxyUUID(u string) ProxyOpt {
	return func(o ProxyOpts) ProxyOpts {
		o.UUID = u
		return o
	}
}

func WithProxySetupTimeout(u time.Duration) ProxyOpt {
	return func(o ProxyOpts) ProxyOpts {
		o.SetupTimeout = u
		return o
	}
}

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

type AgentOpts struct {
	SnapshotLen  int32         // https://www.tcpdump.org/manpages/pcap_set_snaplen.3pcap.html
	Promiscuous  bool          // https://www.tcpdump.org/manpages/pcap_set_promisc.3pcap.html
	Device       string        // https://www.tcpdump.org/manpages/pcap_create.3pcap.html
	Timeout      time.Duration // https://www.tcpdump.org/manpages/pcap_set_timeout.3pcap.html
	TargetIP     string        // proxy endpoint address to send packet via gRPC
	TargetPort   int           // proxy endpoint port to send packet via gRPC
	UUID         string        // kpture uuid used in  ephemeral container name
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

func (a AgentOpts) WithTargetIP(s string) AgentOpts {
	a.TargetIP = s
	return a
}

func (a AgentOpts) WithTargetPort(p int) AgentOpts {
	a.TargetPort = p
	return a
}

func WithAgentDevice(n string) AgentOpt {
	return func(o AgentOpts) AgentOpts {
		o.Device = n
		return o
	}
}

func WithAgentPromiscuous(n bool) AgentOpt {
	return func(o AgentOpts) AgentOpts {
		o.Promiscuous = n
		return o
	}
}

func WithAgentSnapLen(n int32) AgentOpt {
	return func(o AgentOpts) AgentOpts {
		o.SnapshotLen = n
		return o
	}
}

func WithAgentCaptureTimeOut(n time.Duration) AgentOpt {
	return func(o AgentOpts) AgentOpts {
		o.Timeout = n
		return o
	}
}

func WithAgentSetupTimeOut(n time.Duration) AgentOpt {
	return func(o AgentOpts) AgentOpts {
		o.SetupTimeout = n
		return o
	}
}

func WithAgentUUID(u string) AgentOpt {
	return func(o AgentOpts) AgentOpts {
		o.UUID = u
		return o
	}
}
