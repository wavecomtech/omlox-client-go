// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package main

import (
	"github.com/spf13/cobra"
)

func newGenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "gen",
		Aliases: []string{"generate"},
		Short:   "Generate commands",
	}

	cmd.AddCommand(newGenDocsCmd())

	return cmd
}
