/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
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
	"github.com/gmtstephane/kpture/api/capture"
	"github.com/gmtstephane/kpture/pkg/k8s"
	"github.com/gmtstephane/kpture/pkg/kpture"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
	"github.com/sirupsen/logrus"
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
			logrus.Error(err)
			return
		}

		// select the pods to capture
		pods, err := client.SelectPods(args, All)
		if err != nil {
			logrus.Error(err)
			return
		}

		log.Println("Starting the proxy...")
		options := kpture.LoadOptions()
		ip, err := client.SetupProxy(options.UUID, 10000)
		if err != nil {
			logrus.Error(err)
			return
		}
		options.Proxy = ip

		// cleanup the proxy at the end or on interrupt
		defer tearDown(client, options)

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			tearDown(client, options)
			os.Exit(1)
		}()

		log.Println("Port forwarding the proxy...")
		// port forward the proxy
		stopch, port, err := client.PortForward(options.UUID)
		if err != nil {
			log.Println(err)
			return
		}
		defer close(stopch)

		log.Println("Injecting the debug containers...")
		// inject the debug container
		err = client.SetupEphemeralContainers(pods, options)
		if err != nil {
			logrus.Error(err)
			return
		}

		// time.Sleep(1 * time.Second)
		// stopch <- struct{}{}

		conn, err := grpc.Dial(
			fmt.Sprintf("%s:%d", "127.0.0.1", port),
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			logrus.Error(err)
			return
		}
		defer conn.Close()

		cli := capture.NewPackgetGetterClient(conn)

		packets, err := cli.GetPackets(context.Background(), &capture.Empty{})
		if err != nil {
			logrus.Error(err)
			return
		}

		log.Println("Capturing packets 🚀")

		if !Raw {
			w := wow.New(os.Stderr, spin.Get(spin.GrowHorizontal), "")
			w.Start()
			counter := 0
			for {
				_, errReceive := packets.Recv()
				if errReceive != nil {
					logrus.Error(errReceive)
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
			err = pcapwriter.WriteFileHeader(uint32(options.SnapshotLen), layers.LinkTypeEthernet)
			if err != nil {
				log.Println(err)
				return
			}
			for {
				p, errReceive := packets.Recv()
				if errReceive != nil {
					logrus.Error(errReceive)
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

		// logrus.SetFormatter(&nested.Formatter{
		// 	HideKeys:       true,
		// 	NoColors:       true,
		// 	NoFieldsColors: true,
		// 	FieldsOrder:    []string{"component", "category"},
		// })

		// kpture, err := kpture.NewKpture(client, pods)
		// if err != nil {
		// 	logrus.Error("error setting up kpture")
		// 	return
		// }

		// c := make(chan os.Signal)
		// signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		// go func() {
		// 	<-c
		// 	kpture.Stop()
		// 	os.Exit(1)
		// }()

		// count := len(pods)
		// p := uiprogress.New()
		// p.SetOut(os.Stderr)
		// p.Start()
		// bar := p.AddBar(count)
		// bar.PrependFunc(func(b *uiprogress.Bar) string {
		// 	return fmt.Sprintf("Containers (%d/%d) 🚀", b.Current(), count)
		// })
		// bar.AppendCompleted()

		// readychan := make(chan struct{}, count)
		// go func() {
		// 	for {
		// 		select {
		// 		case <-readychan:
		// 			bar.Incr()
		// 			// bar.Add(1)

		// 		}
		// 	}
		// }()

		// err = kpture.SetupEphemeralContainers(readychan)
		// if err != nil {
		// 	logrus.Error(err)
		// 	return
		// }

		// bar2 := p.AddBar(count)
		// bar2.PrependFunc(func(b *uiprogress.Bar) string {
		// 	return fmt.Sprintf("Forwarding (%d/%d) 🔭", b.Current(), count)
		// })
		// bar2.AppendCompleted()
		// readychan2 := make(chan struct{}, count)

		// go func() {
		// 	for {
		// 		select {
		// 		case <-readychan2:
		// 			bar2.Incr()
		// 		}
		// 	}
		// }()

		// err = kpture.SetupPortForwarding(readychan2)
		// if err != nil {
		// 	logrus.Error(err)
		// 	return
		// }
		// kpture.ReadPacketsConn()
		// logrus.SetOutput(io.Discard)
		// bar2.Set(len(pods))
		// p.Stop()
		// writers := map[string]int{}
		// u := uilive.New()
		// u.Flush()
		// u.RefreshInterval = 1000
		// u.Start()
		// // for _, p := range pods {
		// // 	writers[p.Name] = uilive.New()
		// // 	writers[p.Name].RefreshInterval = 1000
		// // 	writers[p.Name].Start()

		// // }
		// // start listening for updates and render
		// go func() {
		// 	for {
		// 		o := ""
		// 		for key, value := range writers {
		// 			o += fmt.Sprintf("%s %d\n", key, value)
		// 		}
		// 		fmt.Fprintf(u, "%s \n", o)
		// 		time.Sleep(1 * time.Second)
		// 	}
		// }()
		// for packet := range kpture.PacketChan() {
		// 	// printlines(writers, writer)
		// 	// fmt.Fprintf(u, "%d \n\r", len(packet.Data))
		// 	writers[packet.Pod]++
		// 	// fmt.Println(writers)
		// }
	},

	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		client, err := kpture.GetClient("")
		if err != nil {
			logrus.Error(err)
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		pods, errList := client.Clientset.List(cmd.Context(), v1.ListOptions{})
		if errList != nil {
			logrus.Error(err)
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
	rootCmd.AddCommand(packetsCmd)
	packetsCmd.Flags().BoolVarP(&All, "all", "a", false, "Capture from all pods in the selected namespace")
	packetsCmd.Flags().BoolVarP(&Raw, "raw", "r", false, "Print raw packet to stdout (for tshark/wireshark)")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// packetsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// packetsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func tearDown(client *k8s.KubeClient, options kpture.Options) {
	logrus.Info("tearing down")
	errteardown := client.TearDownProxy(options.UUID)
	if errteardown != nil {
		logrus.Error(errteardown)
	}
}
