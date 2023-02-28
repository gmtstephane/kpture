/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	Raw       bool
	File      string
	Namespace string
	Pods      []string
	All       bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kpture",
	Short: "Kubernetes packet and log capture tool",
	// 	Long: `
	// Kpture is a kubernetes packet and log capture tool.
	// It allows you to capture packets and logs from multiple pods and write them to a pcap file/logs files.`,
	// ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// },

	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) {
	// },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// func init() {
// 	// Here you will define your flags and configuration settings.
// 	// Cobra supports persistent flags, which, if defined here,
// 	// will be global for your application.

// 	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kpture.yaml)")

// 	// Cobra also supports local flags, which will only run
// 	// when this action is called directly.
// 	client, err := kpture.GetClient("")
// 	if err != nil {
// 		logrus.Error(err)
// 		return
// 	}
// 	rootCmd.Flags().BoolVarP(&Raw, "raw", "r", false, "Print raw packet to stdout (for tshark/wireshark)")
// 	rootCmd.Flags().StringVarP(&File, "write", "w", "", "Write the pcap to a file")
// 	rootCmd.Flags().StringVarP(&Namespace, "namespace", "n", client.Namespace, "Kubernetes namespace to capture from")
// 	rootCmd.Flags().StringArrayVarP(&Pods, "pods", "p", []string{}, "Kubernetes pods to capture from")
// 	rootCmd.Flags().BoolVarP(&All, "all", "a", false, "Capture from all pods in the selected namespace")
// 	rootCmd.RegisterFlagCompletionFunc("pods", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
// 		pods, errList := client.Clientset.List(cmd.Context(), v1.ListOptions{})
// 		if errList != nil {
// 			logrus.Error(err)
// 			return nil, cobra.ShellCompDirectiveNoFileComp
// 		}
// 		var podNames []string
// 		for _, pod := range pods.Items {
// 			podNames = append(podNames, pod.Name)
// 		}
// 		return podNames, cobra.ShellCompDirectiveDefault
// 	})
// }
