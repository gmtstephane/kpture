package main

import (
	"fmt"
	"log"
	"net"

	"github.com/gmtstephane/kpture/api/capture"
	"github.com/gmtstephane/kpture/pkg/kpture"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	logrus.Info("loading env vars")
	godotenv.Load()

	serverOptions, err := kpture.OptFromEnv()
	if err != nil {
		logrus.Error(err)
		return
	}
	server, err := kpture.NewCaptureServer(serverOptions...)
	if err != nil {
		logrus.Error(err)
		return
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", server.Port()))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	logrus.Info("starting gRPC server on port ", server.Port())
	var opts []grpc.ServerOption
	opts = append(opts, grpc.StatsHandler(&kpture.Handler{}))
	grpcServer := grpc.NewServer(opts...)
	capture.RegisterKptureServer(grpcServer, server)
	reflection.Register(grpcServer)
	grpcServer.Serve(lis)
}
