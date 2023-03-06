package proxy

import (
	"context"
	"errors"
	"io"
	"sync"

	capture "github.com/gmtstephane/kpture/api/kpture"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"

	"google.golang.org/grpc/status"
)

type Proxy struct {
	packets   chan *capture.Packet
	started   bool
	cleanup   func(wg *sync.WaitGroup, cancel context.CancelFunc)
	readypods []*capture.Pod
	wg        *sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
	capture.UnimplementedAgentServiceServer
	capture.UnimplementedClientServiceServer
}

func NewProxyServer(bufferSize int, cleanup func(wg *sync.WaitGroup, cancel context.CancelFunc)) *Proxy {
	s := Proxy{
		packets:   make(chan *capture.Packet, bufferSize),
		readypods: make([]*capture.Pod, 0),
		started:   false,
		cleanup:   cleanup,
		wg:        &sync.WaitGroup{},
	}
	s.ctx, s.cancel = context.WithCancel(context.Background())
	return &s
}

func (s *Proxy) AddPacket(packetStream capture.AgentService_AddPacketServer) error {
	logrus.Info("AddPacket")

	s.wg.Add(1)
	defer s.wg.Done()

	go func() {
		for {
			<-s.ctx.Done()
			logrus.Info("Context is Done")
			if err := packetStream.Send(&capture.Empty{}); err != nil {
				logrus.Error(err)
				return
			}
		}
	}()

	for {
		packet, err := packetStream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return status.Error(codes.Internal, err.Error())
		}

		if s.started {
			select {
			case s.packets <- packet:
			default:
				logrus.Error("buffer full")
			}
		}
	}
}

func (s *Proxy) Ready(ctx context.Context, pod *capture.Pod) (*capture.Empty, error) {
	s.readypods = append(s.readypods, pod)
	return &capture.Empty{}, nil
}

func (s *Proxy) GetPackets(in *capture.Empty, stream capture.ClientService_GetPacketsServer) error {
	logrus.Info("GetPackets")
	s.started = true
	for {
		select {
		case <-stream.Context().Done():
			logrus.Info("client stopped connexion, cleaning up...")
			s.cleanup(s.wg, s.cancel)
			return nil
		case p := <-s.packets:
			err := stream.Send(p)
			if err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}
				return status.Error(codes.Internal, err.Error())
			}
		}
	}
}
