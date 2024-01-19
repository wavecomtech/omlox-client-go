// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package main

import (
	"io"

	"github.com/spf13/cobra"
	"github.com/wavecomtech/omlox-client-go/internal/cli"
)

func newDeleteCmd(settings cli.EnvSettings, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete",
		Aliases: []string{"del", "remove"},
		Short:   "Delete hub resources",
	}

	cmd.AddCommand(newDeleteTrackablesCmd(settings, out))
	cmd.AddCommand(newDeleteProvidersCmd(settings, out))

	return cmd
}
