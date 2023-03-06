//go:build cli
// +build cli

/*
Copyright © 2023 Stephane Guillemot <gmtstephane@gmail.com>
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gernest/wow"
	"github.com/gernest/wow/spin"
	capture "github.com/gmtstephane/kpture/api/kpture"
	"github.com/gmtstephane/kpture/pkg/k8s"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var packetsCmd = &cobra.Command{
	Use:   "packets",
	Short: "capture packet from pods",
	Run: func(cmd *cobra.Command, args []string) {
		log.SetFlags(0)
		log.SetOutput(os.Stderr)
		client, err := k8s.GetClient(Namespace)
		if err != nil {
			log.Println(err)
			return
		}

		// select the pods to capture
		pods, err := k8s.SelectPods(args, All, client.Clientset.CoreV1().Pods(client.Namespace))
		if err != nil {
			log.Println(err)
			return
		}
		kptureID := uuid.New().String()
		agentOpts := k8s.LoadAgentOpts(k8s.WithAgentUUID(kptureID))
		proxyOpts := k8s.LoadProxyOpts(k8s.WithProxyUUID(kptureID))

		log.Println("Deploying Proxy")

		ip, err := k8s.SetupProxy(client.Clientset.CoreV1().Pods(client.Namespace), proxyOpts)
		if err != nil {
			log.Println(err)
			return
		}
		agentOpts = agentOpts.WithTargetIP(ip).WithTargetPort(int(proxyOpts.ServerPort))

		// cleanup the proxy at the end
		defer tearDown(client, kptureID)

		// or on interrupt
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			log.Println("")
			tearDown(client, kptureID)
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
			log.Println(err)
			return
		}

		err = k8s.PortForward(forwarder, readychan, agentOpts.SetupTimeout)
		if err != nil {
			log.Println(err)
			return
		}
		defer close(stopchan)

		// inject the debug containers
		err = k8s.SetupEphemeralContainers(pods, client.Clientset.CoreV1().Pods(client.Namespace), agentOpts)
		if err != nil {
			log.Println(err)
			return
		}

		conn, err := grpc.Dial(
			fmt.Sprintf("%s:%d", "127.0.0.1", port),
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Println(err)
			return
		}
		defer conn.Close()

		cli := capture.NewClientServiceClient(conn)

		packets, err := cli.GetPackets(context.Background(), &capture.Empty{})
		if err != nil {
			log.Println(err)
			return
		}

		if !Raw {
			counter := 0
			w := wow.New(os.Stderr, spin.Get(spin.GrowHorizontal), "Captured "+fmt.Sprint(counter)+" packets")
			w.Start()
			for {
				_, errReceive := packets.Recv()
				if errReceive != nil {
					log.Println(errReceive)
					return
				}
				counter++
				w.Text("Captured " + fmt.Sprint(counter) + " packets")
				if err != nil {
					log.Println(err)
					return
				}
			}

		} else {
			pcapwriter := pcapgo.NewWriter(os.Stdout)
			err = pcapwriter.WriteFileHeader(uint32(agentOpts.SnapshotLen), layers.LinkTypeEthernet)
			if err != nil {
				log.Println(err)
				return
			}
			for {
				p, errReceive := packets.Recv()
				if errReceive != nil {
					log.Println(errReceive)
					return
				}

				err = pcapwriter.WritePacket(gopacket.CaptureInfo{
					Timestamp:      time.Now(),
					CaptureLength:  int(p.GetCaptureInfo().GetCaptureLength()),
					Length:         int(p.GetCaptureInfo().GetLength()),
					InterfaceIndex: int(p.GetCaptureInfo().GetInterfaceIndex()),
				}, p.GetData())
				if err != nil {
					log.Println(err)
					return
				}

			}
		}
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
	packetsCmd.Flags().BoolVarP(&All, "all", "a", false, "Capture from all pods in the selected namespace")
	packetsCmd.Flags().BoolVarP(&Raw, "raw", "r", false, "Print raw packet to stdout (for tshark/wireshark)")
}

func tearDown(client *k8s.KubeClient, id string) {
	log.Println("tearing down")
	errteardown := k8s.TearDownProxy(id, client.Clientset.CoreV1().Pods(client.Namespace))
	if errteardown != nil {
		log.Println(errteardown)
	}
}
