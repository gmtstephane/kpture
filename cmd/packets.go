/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/gmtstephane/kpture/pkg/k8s"
	"github.com/gmtstephane/kpture/pkg/kpture"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// func printlines(m map[string]int, i *uilive.Writer) {
// 	i.Flush()
// 	for key, value := range m {
// 	}
// }

// packetsCmd represents the packets command
var packetsCmd = &cobra.Command{
	Use:   "packets",
	Short: "capture packet from pods",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := k8s.GetClient(Namespace)
		if err != nil {
			logrus.Error(err)
			return
		}
		err = client.SetupProxy("1234")
		if err != nil {
			logrus.Error(err)
			return
		}
		// client, err := kpture.GetClient(Namespace)
		// if err != nil {
		// 	logrus.Error(err)
		// 	return
		// }
		// pods, err := client.SelectPods(args, All)
		// if err != nil {
		// 	logrus.Error(err)
		// 	return
		// }

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
