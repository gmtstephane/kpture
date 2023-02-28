package kpture

import (
	"context"
	"io"
	"os"
	"strings"
	"time"

	"github.com/gmtstephane/kpture/api/capture"
	"github.com/gmtstephane/kpture/pkg/pcap"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PodDescriptor struct {
	Name      string
	Namespace string
}

const defaultPacketChanSize uint32 = 1500

type PacketCapture struct {
	*capture.Packet
	Pod string
}

type Kpture struct {
	client     *KubeClient
	packetChan chan *PacketCapture
	errChan    chan error
	kpturePods []*Pod
	opts       pcap.Options
}

func NewKpture(client *KubeClient, pods []PodDescriptor, opts ...pcap.Option) (*Kpture, error) {
	var err error
	k := &Kpture{
		client:     client,
		packetChan: make(chan *PacketCapture, defaultPacketChanSize),
		errChan:    make(chan error),
		kpturePods: []*Pod{},
	}

	id, err := uuid.NewUUID()
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	k.opts = pcap.LoadOptions(opts...)
	for _, pod := range pods {
		pod, errGetPod := client.Clientset.Get(context.Background(), pod.Name, v1.GetOptions{})
		if errGetPod != nil {
			logrus.Error(errGetPod)
			return nil, errGetPod
		}
		for _, eph := range pod.Status.EphemeralContainerStatuses {
			if strings.Contains("kpture", eph.Name) && eph.State.Running != nil {
				logrus.Error("kpture already running on pod ", pod.Name)
				return nil, err
			}
		}

		kpturePod, errcapturePod := NewKpturePod(pod.Name, pod.Namespace, id.String(), k.opts, k.errChan)
		if errcapturePod != nil {
			logrus.Error(errcapturePod)
			return nil, errcapturePod
		}
		k.kpturePods = append(k.kpturePods, kpturePod)
	}
	go k.handleErr()
	return k, nil
}

func (k *Kpture) handleErr() {
	for err := range k.errChan {
		logrus.Error("handleErr", err)
	}
}

func (k *Kpture) SetupEphemeralContainers() error {
	for _, kpturePod := range k.kpturePods {
		err := kpturePod.CreateDebugContainer(k.client.Clientset)
		if err != nil {
			logrus.Error(err)
			return err
		}
	}
	return nil
}

func (k *Kpture) SetupPortForwarding() error {
	for _, kpturePod := range k.kpturePods {
		err := kpturePod.PortForwardAPod(k.client.RestConf)
		if err != nil {
			logrus.Error(err)
			return err
		}
	}
	return nil
}

func (k *Kpture) ReadPacketsConn() {
	for _, kpturePod := range k.kpturePods {
		go kpturePod.ReadPackets(k.packetChan)
	}
}

func (k *Kpture) Stop() {
	for _, kpturePod := range k.kpturePods {
		kpturePod.Close()
	}
}

func (k *Kpture) HandlePackets(out io.Writer) error {
	pcapwriter := pcapgo.NewWriter(out)

	err := pcapwriter.WriteFileHeader(uint32(k.opts.SnapshotLen), layers.LinkTypeEthernet)
	if err != nil {
		return err
	}

	for packet := range k.packetChan {
		err = pcapwriter.WritePacket(gopacket.CaptureInfo{
			Timestamp:      time.Now(),
			CaptureLength:  int(packet.GetCaptureInfo().GetCaptureLength()),
			Length:         int(packet.GetCaptureInfo().GetLength()),
			InterfaceIndex: int(packet.GetCaptureInfo().GetInterfaceIndex()),
		}, packet.GetData())
		if err != nil {
			return err
		}
	}
	return nil
}

func (k *Kpture) HandlePacketsMultipleOutput(dest string) error {
	writers := map[string]*pcapgo.Writer{}

	for _, pod := range k.kpturePods {
		f, err := os.OpenFile(pod.name+".pcap", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
		if err != nil {
			logrus.Error(err)
			return err
		}
		defer f.Close()
		writers[pod.name] = pcapgo.NewWriter(f)
		err = writers[pod.name].WriteFileHeader(uint32(k.opts.SnapshotLen), layers.LinkTypeEthernet)
		if err != nil {
			return err
		}
	}
	// pcapwriter := pcapgo.NewWriter(out)
	for packet := range k.packetChan {
		err := writers[packet.Pod].WritePacket(gopacket.CaptureInfo{
			Timestamp:      time.Now(),
			CaptureLength:  int(packet.GetCaptureInfo().GetCaptureLength()),
			Length:         int(packet.GetCaptureInfo().GetLength()),
			InterfaceIndex: int(packet.GetCaptureInfo().GetInterfaceIndex()),
		}, packet.GetData())
		if err != nil {
			return err
		}
	}
	return nil
}
