package kpture

import (
	"errors"
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func Test_defaultOptions(t *testing.T) {
	tests := []struct {
		name string
		want ServerOptions
	}{
		{
			name: "default options",
			want: ServerOptions{
				snapshotLen: int32(defaultSnapLen),
				promiscuous: defaultPromiscuous,
				device:      defaultDevice,
				timeout:     time.Duration(defaultTimeout),
				port:        defaultPort,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := defaultOptions(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("defaultOptions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_loadOptions(t *testing.T) {
	type args struct {
		os []ServerOption
	}
	tests := []struct {
		name string
		args args
		want ServerOptions
	}{
		{
			name: "load options",
			args: args{
				os: []ServerOption{
					WithInterface("eth0"),
					WithPort(8080),
					WithTimeOut(5 * time.Second),
					WithSnapLen(65535),
					WithPromiscuous(true),
				},
			},
			want: ServerOptions{
				snapshotLen: 65535,
				promiscuous: true,
				device:      "eth0",
				timeout:     5 * time.Second,
				port:        8080,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := loadOptions(tt.args.os...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("loadOptions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithInterface(t *testing.T) {
	type args struct {
		n string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "with interface",
			args: args{
				n: "eth3",
			},
			want: "eth3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, WithInterface(tt.args.n)(defaultOptions()).device)
		})
	}
}

func TestWithPort(t *testing.T) {
	type args struct {
		n int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "with port",
			args: args{
				n: 8080,
			},
			want: 8080,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, WithPort(tt.args.n)(defaultOptions()).port)
		})
	}
}

func TestWithTimeOut(t *testing.T) {
	type args struct {
		n time.Duration
	}
	tests := []struct {
		name string
		args args
		want time.Duration
	}{
		{
			name: "with timeout",
			args: args{
				n: 5 * time.Second,
			},
			want: 5 * time.Second,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, WithTimeOut(tt.args.n)(defaultOptions()).timeout)
		})
	}
}

func TestWithSnapLen(t *testing.T) {
	logrus.SetOutput(io.Discard)
	type args struct {
		n int32
	}
	tests := []struct {
		name string
		args args
		want int32
	}{
		{
			name: "with snaplen",
			args: args{
				n: 65535,
			},
			want: 65535,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, WithSnapLen(tt.args.n)(defaultOptions()).snapshotLen)
		})
	}
}

func TestWithPromiscuous(t *testing.T) {
	type args struct {
		n bool
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "with promiscuous",
			args: args{
				n: true,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, WithPromiscuous(tt.args.n)(defaultOptions()).promiscuous)
		})
	}
}

func TestOptFromEnv(t *testing.T) {
	tests := []struct {
		name    string
		want    ServerOptions
		wantErr bool
		envs    map[string]string
	}{
		{
			name: "opt from env",
			envs: map[string]string{
				EnvSnapLen:     "65535",
				EnvPromiscuous: "true",
				EnvInterface:   "eth0",
				EnvTimeOut:     "5",
				EnvPort:        "8080",
			},
			want: ServerOptions{
				snapshotLen: 65535,
				promiscuous: true,
				device:      "eth0",
				timeout:     5 * time.Second,
				port:        8080,
			},
			wantErr: false,
		},
		{
			name: "invalid snaplen",
			envs: map[string]string{
				EnvSnapLen: "not a number",
			},
			wantErr: true,
		},
		{
			name: "invalid promiscuous",
			envs: map[string]string{
				EnvPromiscuous: "not a bool",
			},
			wantErr: true,
		},
		{
			name: "invalid timeout",
			envs: map[string]string{
				EnvTimeOut: "not a duration",
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			envs: map[string]string{
				EnvPort: "not a number",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envs {
				t.Setenv(k, v)
			}
			got, err := OptFromEnv()

			if err != nil && tt.wantErr {
				var e InvalidEnvParamError
				if !errors.As(err, &e) {
					t.Errorf("OptFromEnv() error = %v, wantErr %v", err, e)
				}
				return
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("OptFromEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			g := loadOptions(got...)
			if !reflect.DeepEqual(g, tt.want) {
				t.Errorf("OptFromEnv() = %v, want %v", g, tt.want)
			}
		})
	}
}
