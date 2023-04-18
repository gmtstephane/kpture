//go:build agent || all
// +build agent all

/*
Copyright © 2023 Stephane Guillemot <gmtstephane@gmail.com>
*/

package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	capture "github.com/gmtstephane/kpture/api/kpture"
	"github.com/gmtstephane/kpture/cmd/utils"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	snapLen               int32
	device                string
	proxyPort             int
	proxyTarget           string
	enableTermMessagePath bool
	filter                string
	termMessagePath       string
)

const (
	defaultSnapLen    = int32(1500)
	defaultTargetPort = 10000
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Start Kpture packet sniffer agent",
	Long: `Kpture agent is a packet sniffer that sends packets to a proxy server.
It is meant to be run as a sidecar/ephemeral container in a pod but can be run standalone.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		t, err := utils.NewTerminationWriter(enableTermMessagePath, termMessagePath)
		if err != nil {
			return err
		}
		hostname, err := os.Hostname()
		if err != nil {
			return t.TerminationMessage(err)
		}

		if proxyTarget == "" {
			return t.TerminationMessage(errors.New("agentProxyTarget not set"))
		}

		// Open device to capture on
		handle, err := pcap.OpenLive(device, snapLen, false, -1)
		if err != nil {
			return t.TerminationMessage(err)
		}
		defer handle.Close()

		//Handle filter
		defaultfilter := fmt.Sprintf("port not %d", proxyPort)
		if filter == "" {
			filter = defaultfilter
		} else {
			filter = fmt.Sprintf("%s and %s", filter, defaultfilter)
		}
		if err = handle.SetBPFFilter(filter); err != nil {
			return t.TerminationMessage(err)
		}

		// Connect to proxy server
		target := fmt.Sprintf("%s:%d", proxyTarget, proxyPort)
		conn, err := grpc.Dial(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return t.TerminationMessage(err)
		}
		defer conn.Close()

		cli := capture.NewAgentServiceClient(conn)
		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

		// Check if proxy server is ready
		_, err = cli.Ready(context.Background(), &capture.Pod{})
		if err != nil {
			return t.TerminationMessage(err)
		}

		addPacketClient, err := cli.AddPacket(context.Background())
		if err != nil {
			return t.TerminationMessage(err)
		}

		stopchan := make(chan error, 1)
		// If we receive a message back, we close the connexion and exit
		go func() {
			for {
				_, err = addPacketClient.Recv()
				stopchan <- err
			}
		}()

		for {
			select {
			case <-stopchan:
				return nil

			// If we receive a packet, we send it to the proxy server
			case packet := <-packetSource.Packets():
				err = addPacketClient.Send(&capture.PacketDescriptor{
					Name: hostname,
					Packet: &capture.Packet{
						Data: packet.Data(),
						CaptureInfo: &capture.CaptureInfo{
							Timestamp:      packet.Metadata().Timestamp.Unix(),
							CaptureLength:  int64(packet.Metadata().CaptureLength),
							Length:         int64(packet.Metadata().Length),
							InterfaceIndex: int64(packet.Metadata().InterfaceIndex),
						},
					},
				})
				if err != nil {
					if errors.Is(err, io.EOF) {
						return nil
					}
					return t.TerminationMessage(err)
				}
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(agentCmd)
	initAgentFlags(agentCmd)
}

func initAgentFlags(cmd *cobra.Command) {
	cmd.Flags().Int32VarP(&snapLen, "snaplen", "l", defaultSnapLen, "Capture snapshot len")
	cmd.Flags().StringVarP(&device, "device", "d", "eth0", "Capture device")
	cmd.Flags().StringVarP(&proxyTarget, "target", "t", "", "Proxy server address")
	cmd.Flags().StringVarP(&termMessagePath, "messagePath", "m", utils.DefaultKubePath, "Termination message path")
	cmd.Flags().StringVarP(&filter, "filter", "f", "", "Capture filter")
	cmd.Flags().BoolVar(&enableTermMessagePath, "togglemessagePath", true, "Toggle  message path")
	cmd.Flags().IntVarP(&proxyPort, "port", "p", defaultTargetPort, "Proxy server port")
}
