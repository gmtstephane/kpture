package kpture

import (
	"context"
	"fmt"
	"os"

	"github.com/gmtstephane/kpture/api/capture"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/stats"
)

type Server struct {
	options ServerOptions
	packets chan gopacket.Packet
	handle  *pcap.Handle
	capture.UnimplementedKptureServer
}

func NewCaptureServer(os ...ServerOption) (*Server, error) {
	var err error
	opts := loadOptions(os...)

	s := Server{
		options: opts,
	}
	logrus.Info(opts)
	s.handle, err = pcap.OpenLive(s.options.device, s.options.snapshotLen, s.options.promiscuous, s.options.timeout)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	if errBPFFilter := s.handle.SetBPFFilter(fmt.Sprintf("port not %d", s.options.port)); errBPFFilter != nil {
		logrus.Error(err)
		return nil, err
	}
	packetSource := gopacket.NewPacketSource(s.handle, s.handle.LinkType())

	s.packets = packetSource.Packets()

	return &s, nil
}

func (s *Server) Port() int {
	return s.options.port
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
