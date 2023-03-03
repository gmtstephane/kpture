package main

import (
	"fmt"
	"net"
	"os"

	"github.com/gmtstephane/kpture/api/capture"
	"github.com/gmtstephane/kpture/pkg/proxy"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	logrus.Info("loading env vars")
	s, err := proxy.NewProxyServer()
	if err != nil {
		terminationMessage(err)
		os.Exit(1)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", 10000))
	if err != nil {
		terminationMessage(err)
		os.Exit(1)
	}

	logrus.Info("starting gRPC server on port ", 10000)
	var opts []grpc.ServerOption
	// opts = append(opts, grpc.StatsHandler(&pcap.Handler{}))
	grpcServer := grpc.NewServer(opts...)
	capture.RegisterPackgetGetterServer(grpcServer, s)
	capture.RegisterPacketsReceiverServer(grpcServer, s)
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
