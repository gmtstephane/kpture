package kpture

import (
	"context"
	"io"
	"time"

	"github.com/gmtstephane/kpture/api/capture"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultPacketChanSize uint32 = 1500
	defaultSnapLen        uint32 = 1500
	defaultPromiscuous    bool   = true
	defaultDevice         string = "eth0"
	defaultTimeout        int    = -1
	defaultPort           int    = 10000
)

type PodDescriptor struct {
	Name      string
	Namespace string
}

type Kpture struct {
	client     *KubeClient
	packetChan chan *capture.Packet
	errChan    chan error
	kpturePods []*Pod
}

func NewKpture(client *KubeClient, pods []PodDescriptor, opts ...ServerOption) (*Kpture, error) {
	var err error
	k := &Kpture{
		client:     client,
		packetChan: make(chan *capture.Packet, defaultPacketChanSize),
		errChan:    make(chan error),
		kpturePods: []*Pod{},
	}

	id, err := uuid.NewUUID()
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	options := loadOptions(opts...)
	for _, pod := range pods {
		_, errGetPod := client.Clientset.Get(context.Background(), pod.Name, v1.GetOptions{})
		if errGetPod != nil {
			logrus.Error(errGetPod)
			return nil, errGetPod
		}

		kpturePod, errcapturePod := NewKpturePod(pod.Name, pod.Namespace, id.String(), options, k.errChan)
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
	err := pcapwriter.WriteFileHeader(defaultSnapLen, layers.LinkTypeEthernet)
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
