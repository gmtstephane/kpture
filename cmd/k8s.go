//go:build cli || all
// +build cli all

/*
Copyright © 2023 Stephane Guillemot <gmtstephane@gmail.com>
*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	capture "github.com/gmtstephane/kpture/api/kpture"
	"github.com/gmtstephane/kpture/cmd/utils"
	"github.com/gmtstephane/kpture/pkg/k8s"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var packetExample = []utils.CommandExample{
	{
		Command: "kpture packets nginx-679f748897-vmc5r nginx-6fdt248897-380f4  -o output",
		Title:   "Start kpture in separated pcap files",
		Additionnal: "This will create the following output directory: " +
			"\n```bash \noutput\n├── nginx-679f748897-vmc5r.pcap\n└── nginx-6fdt248897-380f4.pcap\n```",
	},
	{
		Command: "kpture packets nginx-679f748897-vmc5r nginx-6fdt248897-380f4  --raw | tshark -r -",
		Title:   "Start kpture and pipe the output to tshark",
	},
	{
		Command: "kpture --all -o output --raw | wireshark -k -i -",
		Title:   "Start kpture all packet in current namespace to ./output and pipe the output to wireshark",
	},
}

var packetsCmd = &cobra.Command{
	Use:   "packets",
	Short: "Capture packet from kubernetes pods",
	Long: `
Start a kubernetes packet kpture running these steps:
	
- Inject ephemeral containers to target pods
- Create temporary proxy pod
- Port forwarding proxy pod to local machine
- Retrieve packet via proxy
` + utils.CommandMardkown(packetExample),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !cmd.Flag("output").Changed && !cmd.Flag("raw").Changed {
			return errors.New("must provide output and/or raw flag")
		}

		log.SetFlags(0)
		log.SetOutput(os.Stderr)
		client, err := k8s.GetClient(namespace)
		if err != nil {
			return err
		}

		// Check if the cluster supports Ephemeral Containers
		if err = k8s.CheckEphemeralContainerSupport(client.Clientset.Discovery()); err != nil {
			return err
		}

		// Select pods based on cli args
		pods, err := k8s.SelectPods(args, all, client.Clientset.CoreV1().Pods(client.Namespace))
		if err != nil {
			return err
		}

		// Check if pods are ready and necessary security context for traffic capture
		if err = k8s.CheckPodsContext(pods); err != nil {
			return err
		}

		kptureID := uuid.New().String()

		agentOpts := k8s.LoadAgentOpts(k8s.WithAgentUUID(kptureID), k8s.WithAgentSnapLen(-1), k8s.WithAgentCaptureFilter(capturefilter))
		proxyOpts := k8s.LoadProxyOpts(k8s.WithProxyUUID(kptureID))

		writer, err := newpcapWriter(cmd, pods, uint32(agentOpts.SnapshotLen))
		if err != nil {
			return err
		}

		log.Println("Deploying Proxy")

		// cleanup the proxy at the end
		defer tearDown(client, kptureID)

		ip, err := k8s.SetupProxy(client.Clientset.CoreV1().Pods(client.Namespace), proxyOpts)
		if err != nil {
			return err
		}
		agentOpts = agentOpts.WithTargetIP(ip).WithTargetPort(int(proxyOpts.ServerPort))

		// or on interrupt
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			log.Println("")
			tearDown(client, kptureID)
			writer.cleanup()
			os.Exit(1)
		}()

		log.Println("Forwarding Proxy")

		readychan, stopchan := make(chan struct{}, 1), make(chan struct{}, 1)
		forwarder, port, err := k8s.GetKubeForwarder(
			client.RestConf,
			fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", client.Namespace, "kpture-proxy-"+kptureID),
			readychan,
			stopchan,
			proxyOpts.ServerPort,
		)
		if err != nil {
			return err
		}

		err = k8s.PortForward(forwarder, readychan, agentOpts.SetupTimeout)
		if err != nil {
			return err
		}
		defer close(stopchan)

		// inject the debug containers
		err = k8s.SetupEphemeralContainers(pods, client.Clientset.CoreV1().Pods(client.Namespace), agentOpts)
		if err != nil {
			return err
		}

		conn, err := grpc.Dial(
			fmt.Sprintf("%s:%d", "127.0.0.1", port),
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return err
		}
		defer conn.Close()

		cli := capture.NewClientServiceClient(conn)

		packets, err := cli.GetPackets(context.Background(), &capture.Empty{})
		if err != nil {
			return err
		}

		log.Println("Kpture started, press Ctrl+C to exit")
		if errWriteCapture := writer.WriteCapture(packets); err != nil {
			return errWriteCapture
		}

		return nil
	},

	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		client, err := k8s.GetClient("")
		if err != nil {
			log.Println(err)
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		pods, errList := client.Clientset.CoreV1().Pods(client.Namespace).List(cmd.Context(), v1.ListOptions{})
		if errList != nil {
			log.Println(err)
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		var podNames []string
		// append the pods that are not already in the list
		for _, pod := range pods.Items {
			isin := false
			for _, selected := range args {
				if selected == pod.Name {
					isin = true
				}
			}
			if !isin {
				podNames = append(podNames, pod.Name)
			}
		}

		return podNames, cobra.ShellCompDirectiveDefault
	},
}

func init() {
	RootCmd.AddCommand(packetsCmd)
	packetsCmd.Flags().BoolVarP(&all, "all", "a", false, "Capture from all pods in the selected namespace")
	packetsCmd.Flags().BoolVarP(&raw, "raw", "r", false, "Print raw packet to stdout (for tshark/wireshark)")
	packetsCmd.Flags().StringVarP(&output, "output", "o", "random-kpture-id", "output folder")
	packetsCmd.Flags().StringVarP(&capturefilter, "filter", "f", "", "capture filter")
	packetsCmd.Flags().BoolVarP(&split, "split", "s", true, "split pcap files per pod")
}

func tearDown(client *k8s.KubeClient, id string) {
	log.Println("tearing down")
	errteardown := k8s.TearDownProxy(id, client.Clientset.CoreV1().Pods(client.Namespace))
	if errteardown != nil {
		log.Println(errteardown)
	}
}

type pcapWriter struct {
	podMapWriter       map[string]*pcapgo.Writer
	additionnalWriters []*pcapgo.Writer
	files              []*os.File
	snaplen            uint32
}

func (p *pcapWriter) cleanup() {
	for _, file := range p.files {
		if err := file.Close(); err != nil {
			log.Println(err)
		}
	}
}

func (p *pcapWriter) WriteCapture(
	stream capture.ClientService_GetPacketsClient,
) error {
	for {
		pktStr, errReceive := stream.Recv()
		if errReceive != nil {
			return errReceive
		}
		gop := gopacket.NewPacket(pktStr.Packet.GetData(), layers.LayerTypeEthernet, gopacket.Default)
		if isArp(gop) || isicmpv6sol(gop) {
			continue
		}

		if p.podMapWriter != nil {
			if w, ok := p.podMapWriter[pktStr.GetName()]; ok {
				err := w.WritePacket(gopacket.CaptureInfo{
					Timestamp:      time.Now(),
					CaptureLength:  int(pktStr.Packet.GetCaptureInfo().GetCaptureLength()),
					Length:         int(pktStr.Packet.GetCaptureInfo().GetLength()),
					InterfaceIndex: int(pktStr.Packet.GetCaptureInfo().GetInterfaceIndex()),
				}, pktStr.Packet.GetData())
				if err != nil {
					return err
				}
			}
			for _, additionnal := range p.additionnalWriters {
				err := additionnal.WritePacket(gopacket.CaptureInfo{
					Timestamp:      time.Now(),
					CaptureLength:  int(pktStr.Packet.GetCaptureInfo().GetCaptureLength()),
					Length:         int(pktStr.Packet.GetCaptureInfo().GetLength()),
					InterfaceIndex: int(pktStr.Packet.GetCaptureInfo().GetInterfaceIndex()),
				}, pktStr.Packet.GetData())
				if err != nil {
					return err
				}
			}
		}
	}
}

func newpcapWriter(cmd *cobra.Command, pods []corev1.Pod, snaplen uint32) (*pcapWriter, error) {
	pw := pcapWriter{
		podMapWriter:       make(map[string]*pcapgo.Writer),
		additionnalWriters: []*pcapgo.Writer{},
		files:              []*os.File{},
		snaplen:            snaplen,
	}

	if cmd.Flag("output").Changed {
		err := os.MkdirAll(output, os.ModePerm)
		if err != nil {
			return nil, err
		}
		if errAddFile := pw.addGlobalFile(filepath.Join(output, "kpture.pcap")); errAddFile != nil {
			return nil, errAddFile
		}
	}

	if cmd.Flag("output").Changed && (split && len(pods) > 1) {
		if err := pw.buildpodMap(pods); err != nil {
			return nil, err
		}
	}

	if raw {
		if err := pw.addGlobalWriter(os.Stdout); err != nil {
			return nil, err
		}
	}

	return &pw, nil
}

func (p *pcapWriter) addGlobalWriter(o io.Writer) error {
	w := pcapgo.NewWriter(o)
	err := w.WriteFileHeader(p.snaplen, layers.LinkTypeEthernet)
	if err != nil {
		return err
	}
	p.additionnalWriters = append(p.additionnalWriters, w)
	return nil
}

func (p *pcapWriter) addGlobalFile(path string) error {
	f, errOpenFile := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
	if errOpenFile != nil {
		return errOpenFile
	}
	if erraddWriter := p.addGlobalWriter(f); erraddWriter != nil {
		return erraddWriter
	}
	p.files = append(p.files, f)
	return nil
}

func (p *pcapWriter) buildpodMap(pods []corev1.Pod) error {
	for _, pod := range pods {
		file := filepath.Join(output, pod.Name+".pcap")
		podfile, errOpenPodFile := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
		if errOpenPodFile != nil {
			return errOpenPodFile
		}
		podwriter := pcapgo.NewWriter(podfile)
		err := podwriter.WriteFileHeader(p.snaplen, layers.LinkTypeEthernet)
		if err != nil {
			return err
		}
		p.podMapWriter[pod.Name] = podwriter

		p.files = append(p.files, podfile)
	}

	return nil
}

func isArp(packet gopacket.Packet) bool {
	arpLayer := packet.Layer(layers.LayerTypeARP)
	if arpLayer == nil {
		return false
	}

	arp, ok := arpLayer.(*layers.ARP)
	if !ok {
		return false
	}

	if arp.Operation != layers.ARPRequest || len(arp.DstProtAddress) != 4 {
		return false
	}

	return true
}

func isicmpv6sol(packet gopacket.Packet) bool {
	ipv6Layer := packet.Layer(layers.LayerTypeIPv6)
	if ipv6Layer == nil {
		return false
	}
	_, ok := ipv6Layer.(*layers.IPv6)
	if !ok {
		return false
	}

	// Check if the packet contains an ICMPv6 layer
	icmpv6Layer := packet.Layer(layers.LayerTypeICMPv6)
	if icmpv6Layer == nil {
		return false
	}

	icmpv6, ok := icmpv6Layer.(*layers.ICMPv6)
	if !ok {
		return false
	}

	return icmpv6.TypeCode.Type() == layers.ICMPv6TypeRouterSolicitation
}
