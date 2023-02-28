package main

import (
	"fmt"
	"log"
	"net"

	"github.com/gmtstephane/kpture/api/capture"
	"github.com/gmtstephane/kpture/pkg/kpture"
	"github.com/gmtstephane/kpture/pkg/pcap"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	logrus.Info("loading env vars")
	// godotenv.Load()

	serverOptions, err := kpture.OptFromEnv()
	if err != nil {
		logrus.Error(err)
		return
	}
	s, err := pcap.NewCaptureServer(serverOptions...)
	if err != nil {
		logrus.Error(err)
		return
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", s.Port()))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	logrus.Info("starting gRPC server on port ", s.Port())
	var opts []grpc.ServerOption
	opts = append(opts, grpc.StatsHandler(&pcap.Handler{}))
	grpcServer := grpc.NewServer(opts...)
	capture.RegisterKptureServer(grpcServer, s)
	reflection.Register(grpcServer)
	logrus.Error(grpcServer.Serve(lis))
}
