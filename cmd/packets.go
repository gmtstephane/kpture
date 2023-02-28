/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"os/signal"
	"syscall"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/gmtstephane/kpture/pkg/kpture"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// packetsCmd represents the packets command
var packetsCmd = &cobra.Command{
	Use:   "packets",
	Short: "capture packet from pods",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := kpture.GetClient(Namespace)
		if err != nil {
			logrus.Error(err)
			return
		}
		pods, err := client.SelectPods(args, All)
		if err != nil {
			logrus.Error(err)
			return
		}

		logrus.SetFormatter(&nested.Formatter{
			HideKeys:       true,
			NoColors:       true,
			NoFieldsColors: true,
			FieldsOrder:    []string{"component", "category"},
		})

		kpture, err := kpture.NewKpture(client, pods)
		if err != nil {
			logrus.Error(err)
			return
		}

		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			kpture.Stop()
			os.Exit(1)
		}()

		err = kpture.SetupEphemeralContainers()
		if err != nil {
			logrus.Error(err)
			return
		}

		err = kpture.SetupPortForwarding()
		if err != nil {
			logrus.Error(err)
			return
		}

		kpture.ReadPacketsConn()

		if Raw {
			// logfile, errOpenfile := os.OpenFile("kpture.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
			// if errOpenfile != nil {
			// 	logrus.Error(errOpenfile)
			// 	return
			// }
			logrus.SetOutput(os.Stderr)
			err = kpture.HandlePackets(os.Stdout)
			if err != nil {
				logrus.Error(err)
				return
			}
		} else {
			logrus.SetOutput(os.Stdout)
			err = kpture.HandlePacketsMultipleOutput("")
			if err != nil {
				logrus.Error(err)
				return
			}
		}
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
