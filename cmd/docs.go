//go:build docs || all
// +build docs all

/*
Copyright © 2023 Stephane Guillemot <gmtstephane@gmail.com>
*/

package cmd

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

const fmTemplate = `---
title: "%s"
---
`

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "generate cli markdown docs",

	RunE: func(cmd *cobra.Command, args []string) error {
		filePrepender := func(filename string) string {
			name := filepath.Base(filename)
			base := strings.TrimSuffix(name, path.Ext(name))
			return fmt.Sprintf(fmTemplate, strings.Replace(base, "_", " ", -1))
		}

		err := doc.GenMarkdownTreeCustom(RootCmd, "./docs", filePrepender, func(s string) string { return s })

		return err
	},
}

func init() {
	RootCmd.AddCommand(docsCmd)
}
