// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package main

import (
	"io"

	"github.com/spf13/cobra"
	"github.com/wavecomtech/omlox-client-go/internal/cli"
)

func newCreateCmd(settings cli.EnvSettings, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{"add"},
		Short:   "Create hub resources",
	}

	cmd.AddCommand(newCreateTrackablesCmd(settings, out))
	cmd.AddCommand(newCreateProvidersCmd(settings, out))

	return cmd
}
