package proxy

import (
	"context"
	"errors"
	"io"
	"os"
	"sync"

	"github.com/gmtstephane/kpture/api/capture"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Proxy struct {
	packets   chan *capture.Packet
	started   bool
	readypods []*capture.Pod
	wg        *sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
	capture.UnimplementedPacketsReceiverServer
	capture.UnimplementedPackgetGetterServer
}

func NewProxyServer() (*Proxy, error) {
	s := Proxy{
		packets:   make(chan *capture.Packet, 1500),
		readypods: make([]*capture.Pod, 0),
		started:   false,
		wg:        &sync.WaitGroup{},
	}
	s.ctx, s.cancel = context.WithCancel(context.Background())
	return &s, nil
}

func (s *Proxy) AddPacket(packetStream capture.PacketsReceiver_AddPacketServer) error {
	logrus.Info("AddPacket")

	s.wg.Add(1)
	defer s.wg.Done()

	go func() {
		for {
			select {
			case <-s.ctx.Done():
				logrus.Info("Context is Done")
				if err := packetStream.Send(&capture.Empty{}); err != nil {
					logrus.Error(err)
					return
				}
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
			logrus.Error(err)
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

func (s *Proxy) GetPackets(in *capture.Empty, stream capture.PackgetGetter_GetPacketsServer) error {
	logrus.Info("GetPackets")
	s.started = true
	for {
		select {
		case <-stream.Context().Done():
			logrus.Info("client stopped connexion, cleaning up...")
			s.cleanup()
			return nil
		case p := <-s.packets:
			err := stream.Send(p)
			if err != nil {
				if errors.Is(err, io.EOF) {
					logrus.Error("iof")
					return nil
				}
				logrus.Error(err)
				return status.Error(codes.Internal, err.Error())
			}
		}
	}
}

func (s *Proxy) cleanup() {
	logrus.Info("Cleanup")
	go func() {
		logrus.Info("Waiting contexts...")
		s.wg.Wait()
		logrus.Info("Waiting contexts Done")
		os.Exit(0)
	}()
	s.cancel()
}
