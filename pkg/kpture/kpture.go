package kpture

import (
	"context"
	"io"
	"os"
	"sync"
	"time"

	"github.com/gmtstephane/kpture/api/capture"

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
	opts       Options
}

func NewKpture(client *KubeClient, pods []PodDescriptor, opts ...Option) (*Kpture, error) {
	var err error
	k := &Kpture{
		client:     client,
		packetChan: make(chan *PacketCapture, defaultPacketChanSize),
		errChan:    make(chan error),
		kpturePods: []*Pod{},
	}
	k.opts = LoadOptions(opts...)

	id, err := uuid.NewUUID()
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	for _, pod := range pods {
		pod, errGetPod := client.Clientset.Get(context.Background(), pod.Name, v1.GetOptions{})
		if errGetPod != nil {
			logrus.Error(errGetPod)
			return nil, errGetPod
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

// func (k *Kpture) SetupProxy(readychan chan struct{}) error {

// }

func (k *Kpture) SetupEphemeralContainers(readychan chan struct{}) error {

	wg := sync.WaitGroup{}
	wg.Add(len(k.kpturePods))
	for _, kpturePod := range k.kpturePods {
		go kpturePod.CreateDebugContainer(k.client.Clientset, k.errChan, readychan, &wg)
	}
	wg.Wait()
	return nil
}

func (k *Kpture) SetupPortForwarding(readychan chan struct{}) error {
	wg := sync.WaitGroup{}
	wg.Add(len(k.kpturePods))
	for _, kpturePod := range k.kpturePods {
		go kpturePod.PortForwardAPod(k.client.RestConf, k.errChan, readychan, &wg)

	}
	wg.Wait()
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

func (k *Kpture) PacketChan() chan *PacketCapture {
	return k.packetChan
}
