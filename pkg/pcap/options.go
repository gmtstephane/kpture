package pcap

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

type Option func(o Options) Options

type Options struct {
	SnapshotLen int32
	Promiscuous bool
	Device      string
	Timeout     time.Duration
	Port        int
}

func defaultOptions() Options {
	return Options{
		SnapshotLen: int32(defaultSnapLen),
		Promiscuous: defaultPromiscuous,
		Device:      defaultDevice,
		Timeout:     time.Duration(defaultTimeout),
		Port:        defaultPort,
	}
}

func test() {
	fmt.Println("test")
}
func LoadOptions(os ...Option) Options {
	opts := defaultOptions()
	for _, o := range os {
		opts = o(opts)
	}
	return opts
}

func WithInterface(n string) Option {
	return func(o Options) Options {
		o.Device = n
		return o
	}
}

func WithPort(n int) Option {
	return func(o Options) Options {
		o.Port = n
		return o
	}
}

func WithTimeOut(n time.Duration) Option {
	return func(o Options) Options {
		o.Timeout = n
		return o
	}
}

func WithSnapLen(n int32) Option {
	return func(o Options) Options {
		o.SnapshotLen = n
		return o
	}
}

func WithPromiscuous(n bool) Option {
	return func(o Options) Options {
		o.Promiscuous = n
		return o
	}
}

const (
	EnvPort        = "Kpture_PORT"
	EnvPromiscuous = "Kpture_PROMISCUOUS"
	EnvSnapLen     = "Kpture_SNAPLEN"
	EnvTimeOut     = "Kpture_TIMEOUT"
	EnvInterface   = "Kpture_INTERFACE"
)

func OptFromEnv() ([]Option, error) {
	opts := []Option{}

	if port := os.Getenv(EnvPort); port != "" {
		p, err := strconv.Atoi(port)
		if err != nil {
			logrus.Error("invalid port", err)
			return nil, InvalidEnvParamError{param: EnvPort}
		}
		opts = append(opts, WithPort(p))
	}

	if promiscuous := os.Getenv(EnvPromiscuous); promiscuous != "" {
		p, err := strconv.ParseBool(promiscuous)
		if err != nil {
			logrus.Error("invalid promiscuous", err)
			return nil, InvalidEnvParamError{param: EnvPromiscuous}
		}
		opts = append(opts, WithPromiscuous(p))
	}

	if snapLen := os.Getenv(EnvSnapLen); snapLen != "" {
		p, err := strconv.ParseInt(snapLen, 10, 32)
		if err != nil {
			logrus.Error("invalid snaplen", err)
			return nil, InvalidEnvParamError{param: EnvSnapLen}
		}
		opts = append(opts, WithSnapLen(int32(p)))
	}

	if timeout := os.Getenv(EnvTimeOut); timeout != "" {
		t, err := strconv.Atoi(timeout)
		if err != nil {
			logrus.Error("invalid timeout", err)
			return nil, InvalidEnvParamError{param: EnvTimeOut}
		}
		opts = append(opts, WithTimeOut(time.Duration(t)*time.Second))
	}

	if iface := os.Getenv(EnvInterface); iface != "" {
		opts = append(opts, WithInterface(iface))
	}

	return opts, nil
}
