// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/wavecomtech/omlox-client-go/internal/cli"
	"github.com/wavecomtech/omlox-client-go/internal/cli/output"
)

const getFencesHelp = `
This command retrieves fences from the Omlox Hub.
`

func newGetFencesCmd(settings cli.EnvSettings, out io.Writer) *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "fences",
		Short: "Retrieves fences from the Hub",
		Long:  getFencesHelp,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newOmloxClient(&settings)
			if err != nil {
				return err
			}

			fences, err := c.Fences.List(context.Background())
			if err != nil {
				return err
			}

			o, err := output.ParseFormat(format)
			if err != nil {
				return err
			}

			return o.Write(out, &output.FenceFormater{Fences: fences})
		},
	}

	f := cmd.Flags()
	f.StringVarP((*string)(&format), "output", "o", output.Table.String(), fmt.Sprintf("Output format. One of: %v.", output.Formats()))

	return cmd
}
