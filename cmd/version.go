/*
Copyright © 2023 Stephane Guillemot <gmtstephane@gmail.com>
*/
package cmd

import (
	_ "embed"
	"log"
	"runtime"

	"github.com/spf13/cobra"
)

//go:embed version
var version string

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "show version",
	Run: func(cmd *cobra.Command, args []string) {
		log.SetFlags(0)
		log.Println("kpture:", version)
		log.Println("golang:", runtime.Version())
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
