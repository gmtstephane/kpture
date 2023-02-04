package kpture

import (
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

type ServerOption func(o ServerOptions) ServerOptions

type ServerOptions struct {
	snapshotLen int32
	promiscuous bool
	device      string
	timeout     time.Duration
	port        int
}

func defaultOptions() ServerOptions {
	return ServerOptions{
		snapshotLen: int32(defaultSnapLen),
		promiscuous: defaultPromiscuous,
		device:      defaultDevice,
		timeout:     time.Duration(defaultTimeout),
		port:        defaultPort,
	}
}

func loadOptions(os ...ServerOption) ServerOptions {
	opts := defaultOptions()
	for _, o := range os {
		opts = o(opts)
	}
	return opts
}

func WithInterface(n string) ServerOption {
	return func(o ServerOptions) ServerOptions {
		o.device = n
		return o
	}
}

func WithPort(n int) ServerOption {
	return func(o ServerOptions) ServerOptions {
		o.port = n
		return o
	}
}

func WithTimeOut(n time.Duration) ServerOption {
	return func(o ServerOptions) ServerOptions {
		o.timeout = n
		return o
	}
}

func WithSnapLen(n int32) ServerOption {
	return func(o ServerOptions) ServerOptions {
		o.snapshotLen = n
		return o
	}
}

func WithPromiscuous(n bool) ServerOption {
	return func(o ServerOptions) ServerOptions {
		o.promiscuous = n
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

func OptFromEnv() ([]ServerOption, error) {
	opts := []ServerOption{}

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
