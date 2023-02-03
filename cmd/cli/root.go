/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"io"
	"os"
	"time"

	"github.com/gmtStephane/kpture/api/capture"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "agent",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		var opts []grpc.DialOption
		opts = append(opts, grpc.WithInsecure())
		target := "localhost:8080"
		conn, err := grpc.Dial(target, opts...)
		if err != nil {
			logrus.Error(err)
			return
		}

		defer conn.Close()

		client := capture.NewKptureClient(conn)

		stream, err := client.PacketsStream(cmd.Context(), &capture.Empty{})
		if err != nil {
			logrus.Error(err)
			return
		}
		pcapwriter := pcapgo.NewWriter(os.Stdout)
		pcapwriter.WriteFileHeader(uint32(1500), layers.LinkTypeEthernet)

		for {
			packet, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				logrus.Fatalf("%v.PacketsStream(_) = _, %v", client, err)
			}

			pcapwriter.WritePacket(gopacket.CaptureInfo{
				Timestamp:      time.Now(),
				CaptureLength:  int(packet.GetCaptureInfo().GetCaptureLength()),
				Length:         int(packet.GetCaptureInfo().GetLength()),
				InterfaceIndex: int(packet.GetCaptureInfo().GetInterfaceIndex()),
			}, packet.GetData())

		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.agent.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
