// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"fmt"
	"io"

	"github.com/wavecomtech/omlox-client-go/internal/cli"
	"github.com/wavecomtech/omlox-client-go/internal/cli/output"

	"github.com/spf13/cobra"
)

const getTrackablesHelp = `
This command retrieves trackables from the Omlox Hub.
`

func newGetTrackablesCmd(settings cli.EnvSettings, out io.Writer) *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "trackables",
		Short: "Retrieves trackables from the Hub",
		Long:  getTrackablesHelp,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newOmloxClient(&settings)
			if err != nil {
				return err
			}

			trackables, err := c.Trackables.List(context.Background())
			if err != nil {
				return err
			}

			o, err := output.ParseFormat(format)
			if err != nil {
				return err
			}

			return o.Write(out, &output.TrackableFormater{Trackables: trackables})
		},
	}

	f := cmd.Flags()
	f.StringVarP((*string)(&format), "output", "o", output.Table.String(), fmt.Sprintf("Output format. One of: %v.", output.Formats()))

	return cmd
}
