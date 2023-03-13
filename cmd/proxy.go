//go:build proxy || all
// +build proxy all

/*
Copyright © 2023 Stephane Guillemot <gmtstephane@gmail.com>
*/

package cmd

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"

	capture "github.com/gmtstephane/kpture/api/kpture"
	"github.com/gmtstephane/kpture/cmd/utils"
	"github.com/gmtstephane/kpture/pkg/proxy"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

var (
	serverPort             int32
	bufferSize             int
	pTermMessagePath       string
	pEnableTermMessagePath bool
)

var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "start proxy to receive packet from agent",
	RunE: func(c *cobra.Command, args []string) error {
		t, err := utils.NewTerminationWriter(pEnableTermMessagePath, pTermMessagePath)
		if err != nil {
			return err
		}

		s := proxy.NewProxyServer(bufferSize, CleanUpExit)

		lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", serverPort))
		if err != nil {
			return t.TerminationMessage(err)
		}

		logrus.Info("starting gRPC server on port ", serverPort)
		var opts []grpc.ServerOption
		grpcServer := grpc.NewServer(opts...)
		capture.RegisterAgentServiceServer(grpcServer, s)
		capture.RegisterClientServiceServer(grpcServer, s)
		grpc_health_v1.RegisterHealthServer(grpcServer, health.NewServer())

		reflection.Register(grpcServer)
		return grpcServer.Serve(lis)
	},
}

const (
	defaultProxyPort       = 10000
	defaultProxyBufferSize = 1500
)

func init() {
	RootCmd.AddCommand(proxyCmd)
	proxyCmd.Flags().Int32VarP(&serverPort, "port", "p", defaultProxyPort, "gRPC server port")
	proxyCmd.Flags().IntVarP(&bufferSize, "packet buffer size", "s", defaultProxyBufferSize, "packet buffer size")
	proxyCmd.Flags().StringVarP(&pTermMessagePath, "messagePath", "m", utils.DefaultKubePath, "termination message path")
	proxyCmd.Flags().BoolVarP(&pEnableTermMessagePath, "togglemessagePath", "t", true, "toggle  message path")
}

// CleanUpExit is a function that will be called when the proxy is exiting
// It will wait for all the goroutines to finish before exiting.
var CleanUpExit = func(wg *sync.WaitGroup, cancel context.CancelFunc) {
	go func() {
		wg.Wait()
		os.Exit(0)
	}()
	cancel()
}
