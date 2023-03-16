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
	packets   chan *capture.PacketDescriptor
	started   bool
	mu        sync.Mutex
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
		packets:   make(chan *capture.PacketDescriptor, bufferSize),
		readypods: make([]*capture.Pod, 0),
		started:   false,
		cleanup:   cleanup,
		mu:        sync.Mutex{},
		wg:        &sync.WaitGroup{},
	}
	s.ctx, s.cancel = context.WithCancel(context.Background())
	return &s
}

func (s *Proxy) hasStarted() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.started
}

func (s *Proxy) setStarted() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.started = true
}

func (s *Proxy) AddPacket(packetStream capture.AgentService_AddPacketServer) error {
	logrus.Info("AddPacket")

	s.wg.Add(1)
	defer s.wg.Done()

	go func() {
		<-s.ctx.Done()
		logrus.Info("Context is Done")
		if err := packetStream.Send(&capture.Empty{}); err != nil {
			logrus.Error(err)
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

		if s.hasStarted() {
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
	s.setStarted()
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
