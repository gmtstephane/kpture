package agent

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/gmtstephane/kpture/api/capture"
	"github.com/gmtstephane/kpture/pkg/kpture"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Start() {
	logrus.Info("loading env vars")
	godotenv.Load()
	envs, err := kpture.OptFromEnv()
	if err != nil {
		terminationMessage(err)
		os.Exit(1)
	}

	opt := kpture.LoadOptions(envs...)

	handle, err := pcap.OpenLive(opt.Device, opt.SnapshotLen, opt.Promiscuous, opt.Timeout)
	if err != nil {
		terminationMessage(err)
		os.Exit(1)
	}
	// if errBPFFilter := handle.SetBPFFilter(fmt.Sprintf("port not %d", opt.Port)); errBPFFilter != nil {
	// 	terminationMessage(err)
	// 	os.Exit(1)
	// }
	if errBPFFilter := handle.SetBPFFilter(fmt.Sprintf("icmp")); errBPFFilter != nil {
		terminationMessage(err)
		os.Exit(1)
	}
	conn, err := grpc.Dial(opt.Target(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		terminationMessage(err)
		os.Exit(1)
	}
	defer conn.Close()

	cli := capture.NewPacketsReceiverClient(conn)
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	addPacketClient, err := cli.AddPacket(context.Background())
	if err != nil {
		terminationMessage(err)
		os.Exit(1)
	}

	//If we receive a message back, we close the connexion and exit
	go func() {
		for {
			packet, _ := addPacketClient.Recv()
			fmt.Println(packet)
			logrus.Info("Capture is done")
			conn.Close()
			os.Exit(0)
		}
	}()

	for packet := range packetSource.Packets() {
		err := addPacketClient.Send(&capture.Packet{
			Data: packet.Data(),
			CaptureInfo: &capture.CaptureInfo{
				Timestamp:      packet.Metadata().Timestamp.Unix(),
				CaptureLength:  int64(packet.Metadata().CaptureLength),
				Length:         int64(packet.Metadata().Length),
				InterfaceIndex: int64(packet.Metadata().InterfaceIndex),
			},
		})
		if err != nil {
			if errors.Is(err, io.EOF) {
				terminationMessage(errors.New("Kpture ended " + err.Error()))
				os.Exit(0)
			}
			terminationMessage(err)
			os.Exit(1)
		}
	}
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
