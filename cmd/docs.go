//go:build docs || all
// +build docs all

/*
Copyright © 2023 Stephane Guillemot <gmtstephane@gmail.com>
*/

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "generate cli markdown docs",

	RunE: func(cmd *cobra.Command, args []string) error {
		err := doc.GenMarkdownTree(RootCmd, "./docs")
		return err
	},
}

func init() {
	RootCmd.AddCommand(docsCmd)
}
