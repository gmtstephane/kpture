package pcap

import (
	"context"
	"fmt"
	"os"

	"github.com/gmtstephane/kpture/api/capture"
	"github.com/gmtstephane/kpture/pkg/kpture"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/stats"
)

type Server struct {
	options kpture.Options
	packets chan gopacket.Packet
	handle  *pcap.Handle
	capture.UnimplementedKptureServer
}

func NewCaptureServer(os ...kpture.Option) (*Server, error) {
	var err error
	opts := kpture.LoadOptions(os...)

	s := Server{
		options: opts,
	}
	logrus.Info(opts)
	s.handle, err = pcap.OpenLive(s.options.Device, s.options.SnapshotLen, s.options.Promiscuous, s.options.Timeout)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	if errBPFFilter := s.handle.SetBPFFilter(fmt.Sprintf("port not %d", s.options.Port)); errBPFFilter != nil {
		logrus.Error(err)
		return nil, err
	}
	packetSource := gopacket.NewPacketSource(s.handle, s.handle.LinkType())

	s.packets = packetSource.Packets()

	return &s, nil
}

func (s *Server) Port() int {
	return s.options.Port
}

func (s *Server) PacketsStream(in *capture.Empty, stream capture.Kpture_PacketsStreamServer) error {
	logrus.Info("PacketsStream")
	for packet := range s.packets {
		err := stream.Send(&capture.Packet{
			Data: packet.Data(),
			CaptureInfo: &capture.CaptureInfo{
				Timestamp:      packet.Metadata().Timestamp.Unix(),
				CaptureLength:  int64(packet.Metadata().CaptureLength),
				Length:         int64(packet.Metadata().Length),
				InterfaceIndex: int64(packet.Metadata().InterfaceIndex),
			},
		})
		if err != nil {
			logrus.Info(err)
			return err
		}
	}
	return nil
}

// Stat Handler is used to detect when the client disconnects and exit the server.
type Handler struct{}

func (h *Handler) TagRPC(context.Context, *stats.RPCTagInfo) context.Context {
	return context.Background()
}
func (h *Handler) HandleRPC(context.Context, stats.RPCStats) {}
func (h *Handler) TagConn(context.Context, *stats.ConnTagInfo) context.Context {
	return context.Background()
}

func (h *Handler) HandleConn(c context.Context, s stats.ConnStats) {
	if _, ok := s.(*stats.ConnEnd); ok {
		logrus.Info("client disconnected")
		os.Exit(0)
	}
}
