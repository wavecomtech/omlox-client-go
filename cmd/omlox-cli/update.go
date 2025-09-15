// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package main

import (
	"io"

	"github.com/spf13/cobra"
	"github.com/wavecomtech/omlox-client-go/internal/cli"
)

func newUpdateCmd(settings cli.EnvSettings, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update",
		Aliases: []string{"set"},
		Short:   "Update hub resources",
	}

	cmd.AddCommand(newUpdateTrackablesCmd(settings, out))
	cmd.AddCommand(newUpdateProvidersCmd(settings, out))

	return cmd
}
