package pcap

import (
	"context"
	"errors"
	"flag"
	"io"
	"net"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/gmtstephane/kpture/api/capture"
	"github.com/gmtstephane/kpture/pkg/kpture"
	"github.com/gmtstephane/kpture/pkg/kpture/mocks"

	"github.com/golang/mock/gomock"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/stats"
)

func TestMain(m *testing.M) {
	logrus.SetOutput(io.Discard)
	exitVal := m.Run()
	os.Exit(exitVal)
}

var interfaceName = flag.String("interface", "en0", "interface name")

func TestNewCaptureServer(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping testing in CI environment")
	}
	flag.Parse()
	type args struct {
		os []kpture.Option
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				os: []kpture.Option{
					kpture.WithPort(8888),
					kpture.WithInterface(*interfaceName),
					kpture.WithSnapLen(1500),
					kpture.WithPromiscuous(true),
					kpture.WithTimeOut(5 * time.Second),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCaptureServer(tt.args.os...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCaptureServer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.NotNil(t, got)
			assert.Equal(t, got.Port(), 8888)
			assert.Equal(t, got.options.Device, *interfaceName)
			assert.Equal(t, got.options.SnapshotLen, int32(1500))
			assert.Equal(t, got.options.Promiscuous, true)
			assert.Equal(t, got.options.Timeout, 5*time.Second)
			assert.NotNil(t, got.handle)
			assert.NotNil(t, got.packets)
		})
	}
}

func TestServer_Port(t *testing.T) {
	type fields struct {
		options                   kpture.Options
		packets                   chan gopacket.Packet
		handle                    *pcap.Handle
		UnimplementedKptureServer capture.UnimplementedKptureServer
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "valid port",
			fields: fields{
				options: kpture.Options{
					Port: 8888,
				},
			},
			want: 8888,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				options:                   tt.fields.options,
				packets:                   tt.fields.packets,
				handle:                    tt.fields.handle,
				UnimplementedKptureServer: tt.fields.UnimplementedKptureServer,
			}
			if got := s.Port(); got != tt.want {
				t.Errorf("Server.Port() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServer_PacketsStream(t *testing.T) {
	t.Run("test valid packet", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockServer := mocks.NewMockKpture_PacketsStreamServer(ctrl)
		packetch := make(chan gopacket.Packet)
		s := &Server{
			packets: packetch,
			options: kpture.Options{
				Port: 8888,
			},
		}
		mockServer.EXPECT().Send(gomock.Any()).DoAndReturn(
			func(p *capture.Packet) error {
				assert.Equal(t, p.Data, []byte{0x00, 0x01, 0x02, 0x03, 0x04})
				return nil
			},
		)

		go func() {
			if err := s.PacketsStream(&capture.Empty{}, mockServer); err != nil {
				t.Errorf("Server.PacketsStream() error = %v, wantErr %v", err, false)
			}
		}()
		packetch <- gopacket.NewPacket([]byte{0x00, 0x01, 0x02, 0x03, 0x04}, layers.LinkTypeEthernet, gopacket.Default)
		close(packetch)
	})

	t.Run("test err empty packet", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockServer := mocks.NewMockKpture_PacketsStreamServer(ctrl)
		packetch := make(chan gopacket.Packet)
		s := &Server{
			packets: packetch,
			options: kpture.Options{
				Port: 8888,
			},
		}
		mockServer.EXPECT().Send(gomock.Any()).DoAndReturn(
			func(p *capture.Packet) error {
				assert.Equal(t, p.Data, []byte{})
				return errors.New("Empty packet")
			},
		)

		go func() {
			err := s.PacketsStream(&capture.Empty{}, mockServer)
			assert.NotNil(t, err)
		}()
		packetch <- gopacket.NewPacket([]byte{}, layers.LinkTypeEthernet, gopacket.Default)
		close(packetch)
	})
}

func TestHandler_TagRPC(t *testing.T) {
	type args struct {
		in0 context.Context
		in1 *stats.RPCTagInfo
	}
	tests := []struct {
		name string
		h    *Handler
		args args
		want context.Context
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{}
			if got := h.TagRPC(tt.args.in0, tt.args.in1); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handler.TagRPC() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandler_HandleRPC(t *testing.T) {
	type args struct {
		in0 context.Context
		in1 stats.RPCStats
	}
	tests := []struct {
		name string
		h    *Handler
		args args
	}{
		{
			name: "test valid rpc",
			h:    &Handler{},
			args: args{
				in0: context.Background(),
				in1: &stats.Begin{
					Client: true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{}
			h.HandleRPC(tt.args.in0, tt.args.in1)
		})
	}
}

func TestHandler_TagConn(t *testing.T) {
	type args struct {
		in0 context.Context
		in1 *stats.ConnTagInfo
	}
	tests := []struct {
		name string
		h    *Handler
		args args
		want context.Context
	}{
		{
			name: "test valid tag",
			h:    &Handler{},
			want: context.Background(),
			args: args{
				in0: context.Background(),
				in1: &stats.ConnTagInfo{
					RemoteAddr: &net.UDPAddr{
						IP: net.IPv4(127, 0, 0, 1), Port: 8888,
					},
					LocalAddr: &net.UDPAddr{
						IP: net.IPv4(127, 0, 0, 1), Port: 8888,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{}
			if got := h.TagConn(tt.args.in0, tt.args.in1); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handler.TagConn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandler_HandleConn(t *testing.T) {
	type args struct {
		c context.Context
		s stats.ConnStats
	}
	tests := []struct {
		name string
		h    *Handler
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{}
			h.HandleConn(tt.args.c, tt.args.s)
		})
	}
}
