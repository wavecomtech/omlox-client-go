// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package main

import (
	"io"

	"github.com/spf13/cobra"
	"github.com/wavecomtech/omlox-client-go/internal/cli"
)

func newGetCmd(settings cli.EnvSettings, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get hub resources",
	}

	cmd.AddCommand(newGetTrackablesCmd(settings, out))
	cmd.AddCommand(newGetProvidersCmd(settings, out))

	return cmd
}
