package main

import (
	"fmt"
	"net"
	"os"

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
		terminationMessage(err)
		os.Exit(1)
	}
	s, err := pcap.NewCaptureServer(serverOptions...)
	if err != nil {
		terminationMessage(err)
		os.Exit(1)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", s.Port()))
	if err != nil {
		terminationMessage(err)
		os.Exit(1)
	}

	logrus.Info("starting gRPC server on port ", s.Port())
	var opts []grpc.ServerOption
	opts = append(opts, grpc.StatsHandler(&pcap.Handler{}))
	grpcServer := grpc.NewServer(opts...)
	capture.RegisterKptureServer(grpcServer, s)
	reflection.Register(grpcServer)
	logrus.Error(grpcServer.Serve(lis))
}

func terminationMessage(e error) {
	logrus.Error(e)
	terminationFile, err := os.OpenFile("/dev/termination-log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
	if err != nil {
		logrus.Error(err)
	}
	defer terminationFile.Close()
	_, err = terminationFile.WriteString(e.Error())
	if err != nil {
		logrus.Error(err)
	}
}
