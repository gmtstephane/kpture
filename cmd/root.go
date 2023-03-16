/*
Copyright © 2023 Stephane Guillemot <gmtstephane@gmail.com>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	raw       bool
	output    string
	namespace string
	all       bool
	split     bool
)

// RootCmd represents the base command when called without any subcommands.
var RootCmd = &cobra.Command{
	Use:          "kpture",
	SilenceUsage: true,
	Short:        "Kubernetes packet and log capture tool",
}

func Execute() {
	// RootCmd.AddCommand(agent.Cmd, proxy.Cmd)
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
